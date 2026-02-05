# Go Semantic Cache Gateway

## Description

Go Semantic Cache Gateway is a caching middleware for Large Language Model applications that reduces API costs and response latency through intelligent query matching. The system uses vector embeddings to identify not only exact duplicate queries but also semantically similar questions, serving cached responses when appropriate and only forwarding unique queries to the OpenAI API.

The gateway implements a two-tier caching strategy: exact match lookups using SHA-256 hashing for identical queries, and semantic similarity search using Redis vector operations for paraphrased or equivalently-meaning queries. This approach significantly reduces API calls while maintaining response quality.

## Performance Benchmarks

### Latency
```
Average Response Time: 1608.85ms
Median Response Time:  926.99ms
Fastest Response:      1.67ms (cache hit)
Slowest Response:      8900.01ms (API call)
```

Cache hits are served in milliseconds compared to seconds for API calls, providing 20-25x faster responses for cached queries.

### Cost Savings
```
Cost per 20 Queries (With Cache):    $0.001032
Cost per 20 Queries (Without Cache): $0.001370
Savings:                              $0.000338 (24.63% reduction)
```

### Cache Performance
- Hit rate: 70-80% for typical workloads
- API call reduction: 75%+
- Semantic matching threshold: 0.2 cosine distance

## Architecture

The system follows a middleware pattern where all requests pass through the caching layer before reaching the handler.

### Request Flow
```
Client Request
    ↓
Caching Middleware
    ├─ Parse request body
    ├─ Check exact match (SHA-256 hash lookup)
    │  └─ If found → Return cached response
    ├─ Generate embedding (OpenAI text-embedding-3-small)
    ├─ Vector similarity search (Redis HNSW)
    │  └─ If similar (cosine distance < 0.2) → Return cached response
    └─ Cache miss → Forward to handler
                        ↓
                    Handler
                        ├─ Call OpenAI completion API
                        ├─ Return response to client
                        └─ Cache response asynchronously (background)
```

### Cache Storage Structure

Each cache entry is stored as a Redis HASH:
```
Key: cache:<sha256-hash-of-query>

Fields:
  response:   LLM completion response text
  query:      Original user query
  embedding:  1536-dimension float32 vector (binary)
  created_at: Unix timestamp

TTL: 24 hours
```

## Tech Stack

**Backend**
- Go 1.25.4
- Standard library net/http for HTTP server
- OpenAI Go SDK for API interactions
- go-redis/v9 for Redis operations

**Infrastructure**
- Redis Stack (includes RediSearch module for vector operations)
- Docker and Docker Compose for containerization
- Multi-stage Docker builds for optimized images

**External APIs**
- OpenAI text-embedding-3-small (embeddings)
- OpenAI gpt-4o-mini (completions)

## Prerequisites

- Docker and Docker Compose installed
- OpenAI API key with access to embeddings and completions
- 4GB+ available RAM for Redis vector operations
- Go 1.25.4+ (only required for local development without Docker)

## Getting Started

### 1. Clone the Repository
```bash
git clone https://github.com/adi290491/go-semantic-cache.git
cd go-semantic-cache
```

### 2. Configure Environment Variables

Create a `.env` file in the project root:
```env
OPENAI_API_KEY=sk-proj-your-api-key-here
REDIS_HOSTNAME=redis-stack
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
PORT=8002
```

Replace `sk-proj-your-api-key-here` with your actual OpenAI API key.

### 3. Start the Services
```bash
docker-compose up -d
```

This starts two services:
- `semantic-cache`: The Go application (port 8002)
- `redis-stack`: Redis with RediSearch module (port 6379, Redis Insight on 8001)

### 4. Test the API

Make a test query:
```bash
curl -X POST http://localhost:8002/query \
  -H "Content-Type: application/json" \
  -d '{"query":"Who is the CEO of Apple?"}'
```

Expected response:
```json
{
  "response": "As of June 2024, the CEO of Apple is Tim Cook.",
  "query": "Who is the CEO of Apple?"
}
```

### 5. Test Cache Behavior

Run the same query again to see exact cache hit:
```bash
curl -X POST http://localhost:8002/query \
  -H "Content-Type: application/json" \
  -d '{"query":"Who is the CEO of Apple?"}'
```

This should return instantly (1-2ms) from cache.

Test semantic similarity with a paraphrased query:
```bash
curl -X POST http://localhost:8002/query \
  -H "Content-Type: application/json" \
  -d '{"query":"Who runs Apple?"}'
```

This should also hit the cache due to semantic similarity.

Look for log entries indicating cache hits/misses:
- `"Exact Cache hit"` - Query matched exactly
- `"Similar Cache hit"` - Query matched semantically
- `"Cache miss, calling handler"` - New query, calling OpenAI

### 6. Inspect Cache (Optional)

Access Redis Insight at `http://localhost:8001` to visualize cached data.

Or use Redis CLI:
```bash
docker-compose exec redis-stack redis-cli

# List all cache keys
KEYS cache:*

# View a specific cache entry
HGETALL cache:<key-value>

# Check vector index statistics
FT.INFO cache_idx
```

### 7. Run Benchmarks

Execute the benchmark script to measure performance:
```bash
chmod +x benchmark_cache.sh
./benchmark_cache.sh
```

This runs 20 test queries and provides detailed statistics on latency, cost savings, and cache hit rates.

## API Reference

### POST /query

Execute a query with semantic caching.

**Request Body**
```json
{
  "query": "string (required, max 4000 characters)"
}
```

**Success Response (200 OK)**
```json
{
  "response": "string",
  "query": "string"
}
```

**Error Responses**

400 Bad Request - Invalid or empty query
```json
{
  "message": "query cannot be empty",
  "status_code": 400
}
```

500 Internal Server Error - Server or API failure
```json
{
  "message": "failed to generate response",
  "status_code": 500
}
```

## License

MIT License - see LICENSE file for details.