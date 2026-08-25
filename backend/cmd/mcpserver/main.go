package main

import (
	"context"
	"log"

	"github.com/NivRave/socialfoodie/backend/internal/db"
	"github.com/NivRave/socialfoodie/backend/internal/llm"
	mcp_handlers "github.com/NivRave/socialfoodie/backend/internal/mcp"
	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	godotenv.Load("../.env", "../../.env", ".env")

	ctx := context.Background()

	// Initialize DB
	database, err := db.New(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Pool.Close()

	// Initialize LLM
	llmClient, err := llm.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize LLM client: %v", err)
	}

	h := mcp_handlers.NewHandlers(database, llmClient)

	s := server.NewMCPServer("socialfoodie", "1.0.0", server.WithToolCapabilities(true))
	mcp_handlers.RegisterTools(s, h)

	log.Println("Starting MCP server on stdio...")
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
