# Go Semantic Cache Gateway

A high-performance API Gateway middleware written in Go that optimizes Large Language Model (LLM) costs and latency. It intercepts incoming queries, converts them into vector embeddings, and performs a semantic similarity search against a Redis Vector Database. If a semantically similar query is found, the cached response is returned instantly, bypassing the expensive LLM generation step.

## Performance Benchmarks

The following benchmarks were recorded using a dataset of 20 mixed queries (recurring and unique) running on a local Docker environment.

| Metric | Cache Miss (Direct LLM) | Cache Hit (Semantic Gateway) | Improvement |
| :--- | :--- | :--- | :--- |
| **Max Latency** | 8,900 ms | 1.6 ms | ~5,500x Faster |
| **Average Latency** | 3,000+ ms | < 50 ms | 98% Reduction |
| **Cost (per token)** | Standard API Rate | $0.00 | 100% Savings |

**Summary Statistics:**
* **Total Queries:** 20
* **Cache Hit Rate:** 25.00%
* **Total Cost Savings:** ~24.63% (on mixed traffic)
* **Lowest Latency Recorded:** 1.66ms

## Architecture

The system utilizes a "Dual-Path" retrieval strategy to ensure maximum performance:

1.  **Request Ingestion:** The Go server receives a JSON payload containing the user prompt.
2.  **Vector Embedding:** The prompt is sent to OpenAI's Embedding API (`text-embedding-3-small`) to generate a 1536-dimensional vector.
3.  **Vector Search (RediSearch):** The system queries Redis Stack using HNSW (Hierarchical Navigable Small World) indexing to find vectors with a cosine similarity score below the configured threshold (default: 0.1).
4.  **Decision Logic:**
    * **HIT:** If a similar vector is found, the stored text response is returned immediately.
    * **MISS:** The request is forwarded to the OpenAI Chat Completion API.
5.  **Async Write:** On a cache miss, the new response is returned to the user immediately. A background Goroutine asynchronously writes the new Vector + Text pair to Redis to prevent blocking the response.

## Tech Stack

* **Language:** Golang (1.23)
* **Database:** Redis Stack (Redis + RediSearch + ReJSON)
* **AI Provider:** OpenAI API (Embeddings & Chat Completion)
* **Containerization:** Docker & Docker Compose

## Prerequisites

* Go 1.23+
* Docker & Docker Compose
* OpenAI API Key

## Getting Started

### 1. Clone the Repository

```bash
git clone [https://github.com/yourusername/go-semantic-cache.git](https://github.com/yourusername/go-semantic-cache.git)
cd go-semantic-cache