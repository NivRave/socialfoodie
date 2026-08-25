package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/NivRave/socialfoodie/backend/internal/db"
	"github.com/NivRave/socialfoodie/backend/internal/llm"
	"github.com/NivRave/socialfoodie/backend/internal/rabbitmq"
	"github.com/joho/godotenv"
)

type ScrapePayload struct {
	TraceID    string `json:"trace_id"`
	SourceURL  string `json:"source_url"`
	RawCaption string `json:"raw_caption"`
}

func main() {
	godotenv.Load("../.env", "../../.env", ".env")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	// Initialize RabbitMQ Consumer
	consumer, err := rabbitmq.NewConsumer()
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer consumer.Close()

	handler := func(body []byte) error {
		var payload ScrapePayload
		if err := json.Unmarshal(body, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		log.Printf("Processing recipe from %s", payload.SourceURL)

		// 1. Extract recipe via Gemini
		recipe, err := llmClient.ExtractRecipe(ctx, payload.RawCaption)
		if err != nil {
			return fmt.Errorf("failed to extract recipe via LLM: %w", err)
		}

		// 2. Insert into DB
		recipeID, err := database.InsertRecipe(ctx, payload.SourceURL, payload.TraceID, "instagram", payload.RawCaption, &recipe.Name, &recipe.Instructions, &recipe.Difficulty, &recipe.PrepTime)
		if err != nil {
			return fmt.Errorf("failed to insert recipe: %w", err)
		}

		// 3. Insert Ingredients and link
		for _, ing := range recipe.Ingredients {
			ingID, err := database.InsertIngredient(ctx, ing.Name)
			if err != nil {
				log.Printf("Warning: failed to insert ingredient %s: %v", ing.Name, err)
				continue
			}
			if err := database.LinkRecipeIngredient(ctx, recipeID, ingID, ing.Quantity, ing.Unit); err != nil {
				log.Printf("Warning: failed to link ingredient %s: %v", ing.Name, err)
			}
		}

		// 4. Insert Tags
		for _, tag := range recipe.Tags {
			if err := database.InsertRecipeTag(ctx, recipeID, tag.Tag, tag.Reasoning); err != nil {
				log.Printf("Warning: failed to insert tag %s: %v", tag.Tag, err)
			}
		}

		// 5. Generate Embedding using the parsed text
		embedText := fmt.Sprintf("Recipe: %s\nInstructions: %s\n", recipe.Name, recipe.Instructions)
		for _, tag := range recipe.Tags {
			embedText += fmt.Sprintf("Tag: %s ", tag.Tag)
		}

		embedding, err := llmClient.GenerateEmbedding(ctx, embedText)
		if err != nil {
			return fmt.Errorf("failed to generate embedding: %w", err)
		}

		// 6. Save Embedding
		if err := database.InsertEmbedding(ctx, recipeID, embedding); err != nil {
			return fmt.Errorf("failed to insert embedding: %w", err)
		}

		log.Printf("Successfully processed and saved recipe: %s", recipe.Name)
		return nil
	}

	if err := consumer.StartConsuming("recipe_scraping_queue", handler); err != nil {
		log.Fatalf("Failed to start consuming: %v", err)
	}

	log.Println("Worker started. Waiting for messages...")

	// Wait for interrupt signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	log.Println("Shutting down worker...")
}
