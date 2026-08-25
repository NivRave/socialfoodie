package db_test

import (
	"context"

	"testing"
	"time"

	"github.com/NivRave/socialfoodie/backend/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestDBIntegration(t *testing.T) {
	ctx := context.Background()

	// Spin up a Postgres + pgvector container
	pgContainer, err := postgres.Run(ctx,
		"pgvector/pgvector:pg16",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	require.NoError(t, err)
	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate pgContainer: %s", err)
		}
	}()

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Run migrations using pgxpool directly (since golang-migrate is used in CLI, we can just execute the UP script)
	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)
	defer pool.Close()

	// Minimal migration for tests
	_, err = pool.Exec(ctx, `
		CREATE EXTENSION IF NOT EXISTS vector;
		CREATE TABLE recipes (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			url TEXT UNIQUE NOT NULL,
			shortcode TEXT UNIQUE,
			platform TEXT NOT NULL,
			raw_text TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			name TEXT,
			instructions TEXT,
			difficulty TEXT,
			prep_time_minutes INT
		);
		CREATE TABLE ingredients (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name TEXT UNIQUE NOT NULL
		);
		CREATE TABLE recipe_ingredients (
			recipe_id UUID REFERENCES recipes(id) ON DELETE CASCADE,
			ingredient_id UUID REFERENCES ingredients(id) ON DELETE CASCADE,
			quantity TEXT,
			unit TEXT,
			PRIMARY KEY (recipe_id, ingredient_id)
		);
		CREATE TABLE recipe_tags (
			recipe_id UUID REFERENCES recipes(id) ON DELETE CASCADE,
			tag TEXT NOT NULL,
			reasoning TEXT,
			PRIMARY KEY (recipe_id, tag)
		);
		CREATE TABLE recipe_embeddings (
			recipe_id UUID REFERENCES recipes(id) ON DELETE CASCADE,
			embedding vector(768) NOT NULL,
			PRIMARY KEY (recipe_id)
		);
		CREATE INDEX ON recipe_embeddings USING hnsw (embedding vector_cosine_ops);
	`)
	require.NoError(t, err, "failed to apply schema")

	// Initialize our DB package instance
	database := &db.DB{Pool: pool}

	// 1. Test InsertRecipe
	name := "Test Chicken"
	instructions := "Mix it up"
	difficulty := "Easy"
	prepTime := 15

	recipeID, err := database.InsertRecipe(ctx, "https://inst.com/p/123", "123", "instagram", "Raw caption text", &name, &instructions, &difficulty, &prepTime)
	require.NoError(t, err)
	require.NotEmpty(t, recipeID)

	// Test Duplicate Insert ignores
	recipeID2, err := database.InsertRecipe(ctx, "https://inst.com/p/123", "123", "instagram", "Raw caption text", &name, &instructions, &difficulty, &prepTime)
	require.NoError(t, err)
	assert.Equal(t, recipeID, recipeID2)

	// 2. Test InsertIngredients
	ingID, err := database.InsertIngredient(ctx, "Chicken")
	require.NoError(t, err)
	err = database.LinkRecipeIngredient(ctx, recipeID, ingID, "2", "cups")
	require.NoError(t, err)

	// 3. Test InsertTags
	err = database.InsertRecipeTag(ctx, recipeID, "High Protein", "Lots of chicken")
	require.NoError(t, err)

	// 4. Test InsertEmbedding
	// Create a dummy 768-dimensional array
	dummyEmbedding := make([]float32, 768)
	dummyEmbedding[0] = 0.5
	dummyEmbedding[1] = 0.5

	err = database.InsertEmbedding(ctx, recipeID, dummyEmbedding)
	require.NoError(t, err)
}
