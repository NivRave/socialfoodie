package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/NivRave/socialfoodie/backend/internal/db"
	"github.com/NivRave/socialfoodie/backend/internal/llm"
	mylogger "github.com/NivRave/socialfoodie/backend/internal/logger"
	"github.com/NivRave/socialfoodie/backend/internal/rabbitmq"
	"github.com/joho/godotenv"
)

type ScrapePayload struct {
	TraceID    string `json:"trace_id"`
	SourceURL  string `json:"source_url"`
	Platform   string `json:"platform"`
	RawCaption string `json:"raw_caption"`
}

func main() {
	godotenv.Load("../.env", "../../.env", ".env")
	mylogger.Setup()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	// Initialize RabbitMQ Consumer
	consumer, err := rabbitmq.NewConsumer()
	if err != nil {
		slog.Error("Failed to connect to RabbitMQ", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer consumer.Close()

	handler := func(body []byte) (bool, error) {
		var payload ScrapePayload
		if err := json.Unmarshal(body, &payload); err != nil {
			return false, fmt.Errorf("failed to unmarshal payload (permanent): %w", err)
		}

		// Add trace_id to context for downstream logging
		ctxWithTrace := context.WithValue(ctx, "trace_id", payload.TraceID)
		traceLogger := mylogger.WithTrace(ctxWithTrace)

		traceLogger.Info("Processing recipe", slog.String("url", payload.SourceURL))

		// 1. Extract recipe via Gemini
		recipe, err := llmClient.ExtractRecipe(ctxWithTrace, payload.RawCaption)
		if err != nil {
			traceLogger.Error("failed to extract recipe via LLM", slog.String("error", err.Error()))
			return true, fmt.Errorf("failed to extract recipe via LLM (transient): %w", err)
		}

		// 2. Insert into DB
		recipeID, err := database.InsertRecipe(ctxWithTrace, payload.SourceURL, payload.TraceID, payload.Platform, payload.RawCaption, &recipe.Name, &recipe.Instructions, &recipe.Difficulty, &recipe.PrepTime)
		if err != nil {
			traceLogger.Error("failed to insert recipe", slog.String("error", err.Error()))
			return true, fmt.Errorf("failed to insert recipe (transient): %w", err)
		}

		// 3. Insert Ingredients and link
		for _, ing := range recipe.Ingredients {
			ingID, err := database.InsertIngredient(ctxWithTrace, ing.Name)
			if err != nil {
				traceLogger.Warn("failed to insert ingredient", slog.String("ingredient", ing.Name), slog.String("error", err.Error()))
				continue
			}
			if err := database.LinkRecipeIngredient(ctxWithTrace, recipeID, ingID, ing.Quantity, ing.Unit); err != nil {
				traceLogger.Warn("failed to link ingredient", slog.String("ingredient", ing.Name), slog.String("error", err.Error()))
			}
		}

		// 4. Insert Tags
		for _, tag := range recipe.Tags {
			if err := database.InsertRecipeTag(ctxWithTrace, recipeID, tag.Tag, tag.Reasoning); err != nil {
				traceLogger.Warn("failed to insert tag", slog.String("tag", tag.Tag), slog.String("error", err.Error()))
			}
		}

		// 5. Generate Embedding using the parsed text
		embedText := fmt.Sprintf("Recipe: %s\nInstructions: %s\n", recipe.Name, recipe.Instructions)
		for _, tag := range recipe.Tags {
			embedText += fmt.Sprintf("Tag: %s ", tag.Tag)
		}

		embedding, err := llmClient.GenerateEmbedding(ctxWithTrace, embedText)
		if err != nil {
			traceLogger.Error("failed to generate embedding", slog.String("error", err.Error()))
			return true, fmt.Errorf("failed to generate embedding (transient): %w", err)
		}

		// 6. Save Embedding
		if err := database.InsertEmbedding(ctxWithTrace, recipeID, embedding); err != nil {
			traceLogger.Error("failed to insert embedding", slog.String("error", err.Error()))
			return true, fmt.Errorf("failed to insert embedding (transient): %w", err)
		}

		// Write to Audit Log (Ingestion Success)
		auditDetails := map[string]interface{}{
			"source_url": payload.SourceURL,
			"recipe_id":  recipeID,
			"name":       recipe.Name,
		}
		if err := database.InsertAuditLog(ctxWithTrace, payload.TraceID, "ingestion_success", "go_worker", auditDetails); err != nil {
			traceLogger.Error("failed to write audit log", slog.String("error", err.Error()))
		}

		traceLogger.Info("Successfully processed and saved recipe", slog.String("name", recipe.Name))
		return false, nil
	}

	if err := consumer.StartConsuming("recipe_scraping_queue", handler); err != nil {
		slog.Error("Failed to start consuming", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("Worker started. Waiting for messages...")

	// Wait for interrupt signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	slog.Info("Shutting down worker. Stopping consumer...")
	
	// Stop consuming new messages
	if err := consumer.StopConsuming(); err != nil {
		slog.Error("Error stopping consumer", slog.String("error", err.Error()))
	}
	
	// Wait for in-flight messages
	slog.Info("Waiting for in-flight messages to finish...")
	consumer.Wait()
	
	slog.Info("Worker shutdown complete.")
}
