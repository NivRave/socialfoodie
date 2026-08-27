# socialfoodie

`socialfoodie` is an automated data pipeline and agentic knowledge base designed to scrape, parse, and store culinary content from social media (Instagram and Facebook). The primary goal is to extract unstructured recipe data from video descriptions and transform it into a highly structured format using LLMs.

Once structured, this data is exposed via a Model Context Protocol (MCP) server, allowing autonomous downstream agents to query recipes, suggest meals based on available ingredients, and filter by meal type.

## Architecture & Tech Stack

This project is built using a scalable, event-driven microservices architecture:

- **Scraping Layer:** Python (FastAPI + Instaloader + yt-dlp) for robust social media interaction across Instagram and Facebook.
- **Message Broker:** RabbitMQ to decouple scraping from the processing layer, equipped with a Dead Letter Exchange (DLX) and automatic retry backoff policies for resilience.
- **Backend & MCP Server:** Go, utilizing clean architecture principles for the core extraction worker and agentic interfaces.
- **Data Storage:** PostgreSQL with `pgvector` for structured data and semantic embeddings (using an HNSW index).
- **LLM Processing:** Google Gemini API for structured extraction, prompt-injection resilience, and vector embedding generation.
- **Observability:** Unified JSON structured logging across Python (`python-json-logger`) and Go (`slog`), with context propagation (`trace_id`).
- **Audit Tracking:** An event-driven `audit_logs` PostgreSQL table to track ingestion lifecycle events and record autonomous agent tool calls natively.
- **Infrastructure:** Dockerized microservices configured via `docker-compose`.

## Data Pipeline Flow

1. **Ingestion:** A REST endpoint triggers the Python scraper to fetch an Instagram or Facebook post (URL, raw caption, timestamp). Routing logic dynamically selects the correct scraping engine.
2. **Queueing:** Raw payloads, equipped with a `trace_id` and `platform` identifier, are pushed to a RabbitMQ exchange.
3. **LLM Extraction:** A Go-based worker consumes the raw text and passes it to the Gemini LLM to extract:
   - Recipe Name
   - Normalized Ingredients (Item, Amount, Unit)
   - Preparation Instructions
   - Inferred Meal Tags (ignoring spammy hashtags)
4. **Persistence:** The structured JSON and corresponding vector embeddings are stored in PostgreSQL using UPSERT semantics based on the unique source URL to ensure data integrity. Successful ingestions are permanently recorded in the `audit_logs`.

## Agentic Interface (Model Context Protocol)

The system exposes an MCP server for autonomous agents with the following paginated tools:

- `search_by_ingredients(ingredients, limit, offset)`: Returns recipes containing specific items.
- `semantic_recipe_search(query, limit, offset)`: Uses vector embeddings to find recipes based on abstract concepts (e.g., "comforting winter soup").
- `get_recipes_by_tags(tags, limit, offset)`: Filters recipes by inferred meal tags (e.g., high-protein, dinner).

Every tool call executed by an agent is tracked in the `audit_logs` database table (including tool arguments and results count).

## Development Workflow

Please review the [AGENTS.md](AGENTS.md) file before contributing. Key rules include:
- All work must be done on the `dev` branch.
- Database schemas must use `golang-migrate`.
- Integration and contract tests are required for asynchronous messaging.
- All features require local and Docker build verification and unit testing prior to commit.
