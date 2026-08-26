package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/NivRave/socialfoodie/backend/internal/db"
	"github.com/NivRave/socialfoodie/backend/internal/llm"
	mylogger "github.com/NivRave/socialfoodie/backend/internal/logger"
	mcp_handlers "github.com/NivRave/socialfoodie/backend/internal/mcp"
	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	godotenv.Load("../.env", "../../.env", ".env")
	mylogger.Setup()

	ctx := context.Background()

	// Initialize DB
	database, err := db.New(ctx)
	if err != nil {
		slog.Error("Failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer database.Pool.Close()

	// Initialize LLM
	llmClient, err := llm.NewClient(ctx)
	if err != nil {
		slog.Error("Failed to initialize LLM client", slog.String("error", err.Error()))
		os.Exit(1)
	}

	h := mcp_handlers.NewHandlers(database, llmClient)

	s := server.NewMCPServer("socialfoodie", "1.0.0", server.WithToolCapabilities(true))
	mcp_handlers.RegisterTools(s, h)

	slog.Info("Starting MCP server on stdio...")
	if err := server.ServeStdio(s); err != nil {
		slog.Error("Server error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
