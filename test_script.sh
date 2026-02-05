##### Exact match

# curl -X POST http://localhost:8002/query \
# -H "Content-Type: application/json" \
# -d '{"query":"Who is the CEO of Apple?"}'

# # Wait 2 seconds, then exact same query - should hit exact cache
# sleep 2
# curl -X POST http://localhost:8002/query \
# -H "Content-Type: application/json" \
# -d '{"query":"Who is the CEO of Apple?"}'

##### Semantic Similarity
# Original query
# curl -X POST http://localhost:8002/query \
# -H "Content-Type: application/json" \
# -d '{"query":"Who is the CEO of Apple?"}'

# sleep 2

# # Semantically similar but different wording
# curl -X POST http://localhost:8002/query \
# -H "Content-Type: application/json" \
# -d '{"query":"Who runs Apple?"}'

# # sleep 2

# curl -X POST http://localhost:8002/query \
# -H "Content-Type: application/json" \
# -d '{"query":"Who is the chief executive of Apple?"}'

# sleep 2

# curl -X POST http://localhost:8002/query \
# -H "Content-Type: application/json" \
# -d '{"query":"Can you tell me who leads Apple?"}'

# ### Different topics

# # Query 1
curl -X POST http://localhost:8002/query \
-H "Content-Type: application/json" \
-d '{"query":"Who is the CEO of Apple?"}'

sleep 2

# Query 2 - completely different topic
curl -X POST http://localhost:8002/query \
-H "Content-Type: application/json" \
-d '{"query":"What is the capital of France?"}'

sleep 2

# Query 3 - another different topic
curl -X POST http://localhost:8002/query \
-H "Content-Type: application/json" \
-d '{"query":"How do airplanes fly?"}'

