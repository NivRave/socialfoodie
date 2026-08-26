# MVP (V1) Architecture Plan: socialfoodie

This document outlines the Minimum Viable Product (MVP) architecture for `socialfoodie`.

## 1. Functional Requirements (MVP)
*   **Ingestion:** Ability to scrape a specific Instagram post (via URL) and extract the raw caption text and video timestamp. This will be triggered via a simple REST endpoint on the Python scraper service.
*   **Processing:** Asynchronously process the scraped text using an LLM to extract structured recipe data:
    *   Recipe Name
    *   Ingredients (Item, Amount, Unit) - *Note: The LLM must normalize ingredient names (e.g., "Potato", "potatos", "potato" should all map to a standardized base term like "potato") to prevent duplication.*
    *   Preparation Instructions
    *   Meal Tags (e.g., breakfast, dessert) - *Note: The LLM must deduce these tags from the recipe/ingredients and provide its reasoning, ignoring uninformative or spammy Instagram hashtags.*
*   **Storage:** Store the structured recipe and generate vector embeddings for semantic search.
*   **Agentic API (MCP):** Expose a Model Context Protocol server with at least three tools:
    *   `search_by_ingredients(ingredients)`
    *   `semantic_recipe_search(query)`
    *   `get_recipes_by_tags(tags)`

## 2. Non-Functional Requirements (MVP)
*   **Reliability:** The scraping and LLM processing must be decoupled. A Dead Letter Exchange (DLX) and retry backoff policy must be implemented in RabbitMQ to handle LLM rate limits or poison pill payloads safely.
*   **Observability:** A `trace_id` must be propagated from the Python scraper, through RabbitMQ, and into the Go workers for distributed debugging.
*   **Audit Logging:** The system must maintain an event-driven `audit_logs` table to track the ingestion lifecycle (e.g., successful/failed scraping events) and Agent usage (e.g., recording which MCP tools were executed, at what time, and with what arguments).
*   **Security:** The system must implement robust prompt injection protections to prevent malicious payloads embedded in Instagram captions from manipulating the LLM.
*   **Scalability:** Containerized services (Docker) that can be run locally via `docker-compose` for easy development and deployment.
*   **Scraping Authentication:** The Python scraper must function without Instagram credentials (POC required). If Instagram blocks unauthenticated scraping, alternative strategies (like proxies or sessions) will be evaluated later.
*   **Data Integrity:** The system must prevent duplicate scraping and parsing by enforcing uniqueness on the `source_url` for every recipe.

## 3. Tech Stack Selection
*   **Scraping Service:** Python wrapped in a lightweight REST API (e.g., FastAPI) using a library like `instaloader`.
*   **Message Broker:** RabbitMQ.
*   **Core Backend & MCP Server:** Go.
*   **Database:** PostgreSQL with `pgvector` extension.
*   **LLM Provider:** Google Gemini API.

## 4. High-Level Architecture
```mermaid
graph TD
    Trigger([External Trigger]) -->|POST /scrape {url}| Scraper[Python Scraper API]
    Scraper -->|Pushes raw JSON| Queue[(RabbitMQ)]
    Queue -->|Consumes message| GoWorker[Go Extractor Worker]
    GoWorker <-->|Prompt, Schema, Anti-Injection| LLM[Gemini API]
    GoWorker -->|Saves structured JSON & Vector| DB[(PostgreSQL + pgvector)]
    
    Agent([Autonomous Agent]) <-->|MCP Protocol| MCPServer[Go MCP Server]
    MCPServer <-->|SQL/Vector Queries| DB
```

## 5. Database Schema (PostgreSQL + pgvector)

```sql
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE recipes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_url TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    raw_caption TEXT,
    instructions TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE ingredients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipe_id UUID REFERENCES recipes(id) ON DELETE CASCADE,
    item_name TEXT NOT NULL,
    amount NUMERIC, 
    unit TEXT
);

CREATE TABLE recipe_tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipe_id UUID REFERENCES recipes(id) ON DELETE CASCADE,
    tag_name TEXT NOT NULL,
    llm_reasoning TEXT
);

-- Assuming Gemini embeddings (e.g., text-embedding-004) use 768 dimensions
CREATE TABLE recipe_embeddings (
    recipe_id UUID PRIMARY KEY REFERENCES recipes(id) ON DELETE CASCADE,
    embedding vector(768)
);

-- HNSW index for scalable approximate nearest neighbor search
CREATE INDEX ON recipe_embeddings USING hnsw (embedding vector_cosine_ops);

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trace_id TEXT, -- For linking with RabbitMQ trace_ids
    event_type TEXT NOT NULL, -- e.g., 'mcp_tool_call', 'ingestion_success', 'ingestion_failed'
    event_source TEXT NOT NULL, -- e.g., 'mcp_server', 'go_worker', 'python_scraper'
    details JSONB, -- Flexible payload for tool arguments, error messages, or metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

## 6. API & Interfaces

### 6.1 Scraper REST API (Python)
*   **Endpoint:** `POST /scrape`
*   **Payload:** `{ "url": "https://www.instagram.com/p/XYZ123/" }`
*   **Response:** `202 Accepted` (The actual scraping and publishing to the queue happens in the background).

### 6.2 RabbitMQ Message Payload
```json
{
  "trace_id": "uuid-v4-string",
  "source_url": "https://www.instagram.com/p/XYZ123/",
  "raw_caption": "...",
  "timestamp": "2023-10-12T15:00:00Z",
  "scraped_at": "2024-01-01T12:00:00Z"
}
```

### 6.3 MCP Server Tools (Go)
*   **`search_by_ingredients`**
    *   *Input:* `ingredients` (array of strings), `limit` (int), `offset` (int)
    *   *Output:* Paginated list of recipes containing all or most of the specified ingredients. Uses SQL `ILIKE` or full-text search against the `ingredients` table.
*   **`semantic_recipe_search`**
    *   *Input:* `query` (string), `limit` (int), `offset` (int)
    *   *Output:* Paginated list of recipes. The Go server generates an embedding for the query using the Gemini API, then performs a cosine similarity search (`<=>`) against the `recipe_embeddings` table.
*   **`get_recipes_by_tags`**
    *   *Input:* `tags` (array of strings), `limit` (int), `offset` (int)
    *   *Output:* Paginated list of recipes that match the provided tags.
