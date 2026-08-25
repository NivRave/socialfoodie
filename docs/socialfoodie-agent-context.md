# Project: socialfoodie

## 1. Project Overview
`socialfoodie` is an automated data pipeline and agentic knowledge base designed to scrape, parse, and store culinary content from social media (specifically Instagram). The primary goal is to extract unstructured recipe data from video descriptions and transform it into a highly structured format using LLMs. 

Once structured, this data is exposed via a Model Context Protocol (MCP) server, allowing autonomous downstream agents to query recipes, suggest meals based on available ingredients, and filter by meal type.

## 2. System Architecture & Tech Stack
To ensure a highly scalable, event-driven architecture, `socialfoodie` should be built using the following stack:
*   **Scraping Layer:** Python (Playwright / Instaloader) for robust social media interaction.
*   **Message Broker:** RabbitMQ to decouple the scraping layer from the processing layer.
*   **Backend & MCP Server:** Go, utilizing clean architecture principles for the core extraction service and agentic interfaces.
*   **Data Storage:** PostgreSQL (with `pgvector` for semantic search) or MongoDB for flexible document storage.
*   **Infrastructure:** Dockerized microservices.

## 3. Data Pipeline
1.  **Ingestion:** The Python scraper fetches Instagram Reels/Posts (Targeting: URL, raw caption text, timestamp).
2.  **Queueing:** Raw payloads are pushed to a RabbitMQ exchange.
3.  **LLM Extraction:** A Go-based worker consumes the raw text and passes it to an LLM enforcing a strict structured output schema (e.g., JSON Schema).
4.  **Parsing:** The LLM extracts:
    *   Recipe Name
    *   List of Ingredients (Item, Amount, Unit)
    *   Preparation Instructions
    *   Meal Tags / Types (e.g., high-protein, dinner)
5.  **Persistence:** The structured JSON and corresponding vector embeddings are stored in the database.

## 4. Agentic Interface (Model Context Protocol)
To enable seamless integration with the Antigravity IDE and other autonomous agents, the system exposes an MCP server with tools such as:
*   `search_by_ingredients(ingredients []string)`: Returns recipes containing specific items.
*   `semantic_recipe_search(query string)`: Uses vector embeddings to find recipes based on abstract concepts (e.g., "comforting winter soup").
*   `get_recipes_by_type(meal_type string)`: Filters by categorized tags.

## 5. Agent Instructions
When assisting with this codebase, the agent must:
*   Prioritize modularity, SOLID principles, and event-driven design.
*   Ensure the LLM extraction prompts are highly resilient to emojis, slang, missing punctuation, and informal units of measurement common on Instagram.
*   Maintain clean, idiomatic Go for the backend services and robust Python for the scraping edge.
