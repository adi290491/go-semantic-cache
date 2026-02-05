#!/bin/bash

set -e

BASE_URL="http://localhost:8002"
RESULTS_DIR="benchmark_results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RESULTS_FILE="$RESULTS_DIR/benchmark_$TIMESTAMP.txt"

mkdir -p "$RESULTS_DIR"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "======================================"
echo "🚀 Semantic Cache Benchmark"
echo "======================================"
echo "Timestamp: $TIMESTAMP"
echo "Results will be saved to: $RESULTS_FILE"
echo ""

# Clear Redis cache before starting
echo "🧹 Clearing Redis cache..."
docker-compose exec -T redis-stack redis-cli FLUSHDB > /dev/null
echo "✅ Cache cleared"
echo ""

# Test queries - mix of exact duplicates and similar queries
declare -a QUERIES=(
    "Who is the CEO of Apple?"
    "Who is the CEO of Apple?"  # Exact duplicate
    "Who runs Apple?"  # Similar
    "Who is the chief executive of Apple?"  # Similar
    "Who leads Apple?"  # Similar
    "What is the capital of France?"
    "What is the capital of France?"  # Exact duplicate
    "What's France's capital city?"  # Similar
    "Tell me the capital of France"  # Similar
    "Who is the CEO of Google?"
    "Who is the CEO of Google?"  # Exact duplicate
    "Who runs Google?"  # Similar
    "What is quantum computing?"
    "What is quantum computing?"  # Exact duplicate
    "Explain quantum computing"  # Similar
    "How does quantum computing work?"  # Similar
    "What are the planets in our solar system?"
    "What are the planets in our solar system?"  # Exact duplicate
    "List the planets in the solar system"  # Similar
    "How many planets are in the solar system?"  # Similar
)

TOTAL_QUERIES=${#QUERIES[@]}
CACHE_HITS=0
CACHE_MISSES=0
TOTAL_LATENCY=0
OPENAI_CALLS=0

declare -a LATENCIES=()

# Initialize results file
{
    echo "======================================"
    echo "Semantic Cache Benchmark Results"
    echo "======================================"
    echo "Timestamp: $TIMESTAMP"
    echo "Total Queries: $TOTAL_QUERIES"
    echo ""
    echo "Query Results:"
    echo "--------------------------------------"
} > "$RESULTS_FILE"

echo "🧪 Running $TOTAL_QUERIES queries..."
echo ""

for i in "${!QUERIES[@]}"; do
    query="${QUERIES[$i]}"
    query_num=$((i + 1))
    
    echo -ne "${BLUE}[$query_num/$TOTAL_QUERIES]${NC} "
    
    # Make request and measure time
    start_time=$(date +%s%N)
    
    response=$(curl -s -w "\n%{http_code}\n%{time_total}" -X POST "$BASE_URL/query" \
        -H "Content-Type: application/json" \
        -d "{\"query\":\"$query\"}")
    
    end_time=$(date +%s%N)
    
    # Parse response
    lines=$(echo "$response" | wc -l | tr -d ' ')
    body_lines=$((lines - 2))

    body=$(echo "$response" | head -n $body_lines)
    http_code=$(echo "$response" | tail -n2 | head -n1)
    time_total=$(echo "$response" | tail -n1)
        
    # Convert to milliseconds
    latency_ms=$(echo "$time_total * 1000" | bc)
    LATENCIES+=("$latency_ms")
    TOTAL_LATENCY=$(echo "$TOTAL_LATENCY + $latency_ms" | bc)
    
    # Check logs to determine if it was a cache hit or miss
    # (Sleep briefly to let logs flush)
    sleep 0.5
    
    # Get last few log lines
    recent_logs=$(docker-compose logs --tail=5 semantic-cache 2>/dev/null | grep -E "(Exact Cache hit|Similar Cache hit|Cache miss, calling handler)" | tail -n1)
    
    if echo "$recent_logs" | grep -q "Exact Cache hit"; then
        status="${GREEN}✓ EXACT HIT${NC}"
        CACHE_HITS=$((CACHE_HITS + 1))
    elif echo "$recent_logs" | grep -q "Similar Cache hit"; then
        status="${GREEN}✓ SIMILAR HIT${NC}"
        CACHE_HITS=$((CACHE_HITS + 1))
    else
        status="${YELLOW}✗ MISS${NC}"
        CACHE_MISSES=$((CACHE_MISSES + 1))
        OPENAI_CALLS=$((OPENAI_CALLS + 1))
    fi
    
    echo -e "$status | ${latency_ms}ms | Query: \"$query\""
    
    # Log to file
    {
        echo "Query $query_num: $query"
        echo "Status: $http_code | Latency: ${latency_ms}ms"
        echo "Type: $status"
        echo "---"
    } >> "$RESULTS_FILE"
    
    # Small delay between requests
    sleep 1
done

echo ""
echo "======================================"
echo "📊 Results Summary"
echo "======================================"

# Calculate statistics
AVG_LATENCY=$(echo "scale=2; $TOTAL_LATENCY / $TOTAL_QUERIES" | bc)
CACHE_HIT_RATE=$(echo "scale=2; ($CACHE_HITS * 100) / $TOTAL_QUERIES" | bc)

# Sort latencies to find median
IFS=$'\n' sorted_latencies=($(sort -n <<<"${LATENCIES[*]}"))
unset IFS

MEDIAN_INDEX=$((TOTAL_QUERIES / 2))
MEDIAN_LATENCY=${sorted_latencies[$MEDIAN_INDEX]}

# Find min and max
MIN_LATENCY=${sorted_latencies[0]}
MAX_LATENCY=${sorted_latencies[${#sorted_latencies[@]}-1]}

# Calculate average for cache hits vs misses
# (This is approximate - you'd need to track separately for exact numbers)

echo ""
echo "Total Queries:       $TOTAL_QUERIES"
echo -e "Cache Hits:          ${GREEN}$CACHE_HITS${NC}"
echo -e "Cache Misses:        ${YELLOW}$CACHE_MISSES${NC}"
echo -e "OpenAI API Calls:    ${RED}$OPENAI_CALLS${NC}"
echo ""
echo -e "Cache Hit Rate:      ${GREEN}${CACHE_HIT_RATE}%${NC}"
echo ""
echo "Latency Statistics:"
echo "  Average:           ${AVG_LATENCY}ms"
echo "  Median:            ${MEDIAN_LATENCY}ms"
echo "  Min:               ${MIN_LATENCY}ms"
echo "  Max:               ${MAX_LATENCY}ms"
echo ""

# Cost calculation (approximate)
# text-embedding-3-small: ~$0.00002 per 1K tokens
# gpt-4o-mini: ~$0.15 per 1M input tokens, ~$0.60 per 1M output tokens
# Assume avg 50 input tokens, 100 output tokens per query

EMBEDDING_COST_PER_QUERY=0.000001  # $0.000001 per embedding
COMPLETION_INPUT_COST=0.0000075    # $0.15 / 1M * 50 tokens
COMPLETION_OUTPUT_COST=0.00006     # $0.60 / 1M * 100 tokens
COMPLETION_TOTAL_COST=$(echo "$COMPLETION_INPUT_COST + $COMPLETION_OUTPUT_COST" | bc)

# Total cost with cache
TOTAL_COST_WITH_CACHE=$(echo "($TOTAL_QUERIES * $EMBEDDING_COST_PER_QUERY) + ($OPENAI_CALLS * $COMPLETION_TOTAL_COST)" | bc)

# Cost without cache (every query calls OpenAI)
TOTAL_COST_WITHOUT_CACHE=$(echo "$TOTAL_QUERIES * ($EMBEDDING_COST_PER_QUERY + $COMPLETION_TOTAL_COST)" | bc)

COST_SAVINGS=$(echo "$TOTAL_COST_WITHOUT_CACHE - $TOTAL_COST_WITH_CACHE" | bc)
COST_SAVINGS_PERCENT=$(echo "scale=2; ($COST_SAVINGS * 100) / $TOTAL_COST_WITHOUT_CACHE" | bc)

echo "Cost Analysis (Approximate):"
echo "  Cost WITH cache:    \$$(printf '%.6f' $TOTAL_COST_WITH_CACHE)"
echo "  Cost WITHOUT cache: \$$(printf '%.6f' $TOTAL_COST_WITHOUT_CACHE)"
echo -e "  ${GREEN}Savings:            \$$(printf '%.6f' $COST_SAVINGS) (${COST_SAVINGS_PERCENT}%)${NC}"
echo ""

# Extrapolate to larger scale
echo "📈 Extrapolated Savings (at scale):"
echo ""

for scale in 1000 10000 100000; do
    scaled_with=$(echo "$TOTAL_COST_WITH_CACHE * $scale / $TOTAL_QUERIES" | bc -l)
    scaled_without=$(echo "$TOTAL_COST_WITHOUT_CACHE * $scale / $TOTAL_QUERIES" | bc -l)
    scaled_savings=$(echo "$scaled_without - $scaled_with" | bc -l)
    
    echo "  At $scale queries/day:"
    echo "    Cost WITH cache:    \$$(printf '%.2f' $scaled_with)"
    echo "    Cost WITHOUT cache: \$$(printf '%.2f' $scaled_without)"
    echo -e "    ${GREEN}Daily Savings:      \$$(printf '%.2f' $scaled_savings)${NC}"
    echo -e "    ${GREEN}Monthly Savings:    \$$(printf '%.2f' $(echo "$scaled_savings * 30" | bc -l))${NC}"
    echo -e "    ${GREEN}Annual Savings:     \$$(printf '%.2f' $(echo "$scaled_savings * 365" | bc -l))${NC}"
    echo ""
done

echo "======================================"

# Append summary to results file
{
    echo ""
    echo "======================================"
    echo "Summary Statistics"
    echo "======================================"
    echo "Total Queries: $TOTAL_QUERIES"
    echo "Cache Hits: $CACHE_HITS"
    echo "Cache Misses: $CACHE_MISSES"
    echo "Cache Hit Rate: ${CACHE_HIT_RATE}%"
    echo "OpenAI API Calls: $OPENAI_CALLS"
    echo ""
    echo "Latency:"
    echo "  Average: ${AVG_LATENCY}ms"
    echo "  Median: ${MEDIAN_LATENCY}ms"
    echo "  Min: ${MIN_LATENCY}ms"
    echo "  Max: ${MAX_LATENCY}ms"
    echo ""
    echo "Cost:"
    echo "  With Cache: \$$(printf '%.6f' $TOTAL_COST_WITH_CACHE)"
    echo "  Without Cache: \$$(printf '%.6f' $TOTAL_COST_WITHOUT_CACHE)"
    echo "  Savings: \$$(printf '%.6f' $COST_SAVINGS) (${COST_SAVINGS_PERCENT}%)"
} >> "$RESULTS_FILE"

echo ""
echo "✅ Benchmark complete! Results saved to: $RESULTS_FILE"
echo ""