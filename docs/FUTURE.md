# Future Features & Ideas

This document tracks ideas, features, and improvements that are out of scope for the MVP but should be considered for future iterations of `socialfoodie`.

*   **Mark Recipes as Used:** Add the ability to mark recipes that were used/cooked. This could hook into an alerting or meal-planning system to ensure variety and track what has been tested.
*   **Centralized Log Aggregation:** Implement a logging stack (e.g., Grafana Loki + Promtail) in `docker-compose.yml` to ingest the structured JSON logs from the scraper and Go workers. This will provide a single pane of glass for monitoring system health and debugging distributed trace IDs.
