package db

import (
	"context"
	"fmt"
)

func (db *DB) InsertRecipe(ctx context.Context, url, shortcode, platform, rawText string, name, instructions *string, difficulty *string, prepTime *int) (string, error) {
	var id string
	query := `
		INSERT INTO recipes (url, shortcode, platform, raw_text, name, instructions, difficulty, prep_time_minutes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (url) DO UPDATE SET 
			name = EXCLUDED.name,
			instructions = EXCLUDED.instructions
		RETURNING id
	`
	err := db.Pool.QueryRow(ctx, query, url, shortcode, platform, rawText, name, instructions, difficulty, prepTime).Scan(&id)
	return id, err
}

func (db *DB) InsertIngredient(ctx context.Context, name string) (string, error) {
	var id string
	query := `
		INSERT INTO ingredients (name) VALUES ($1)
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`
	err := db.Pool.QueryRow(ctx, query, name).Scan(&id)
	return id, err
}

func (db *DB) LinkRecipeIngredient(ctx context.Context, recipeID, ingredientID, quantity, unit string) error {
	query := `
		INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity, unit)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (recipe_id, ingredient_id) DO NOTHING
	`
	_, err := db.Pool.Exec(ctx, query, recipeID, ingredientID, quantity, unit)
	return err
}

func (db *DB) InsertRecipeTag(ctx context.Context, recipeID, tag, reasoning string) error {
	query := `
		INSERT INTO recipe_tags (recipe_id, tag, reasoning)
		VALUES ($1, $2, $3)
		ON CONFLICT (recipe_id, tag) DO NOTHING
	`
	_, err := db.Pool.Exec(ctx, query, recipeID, tag, reasoning)
	return err
}

func (db *DB) InsertEmbedding(ctx context.Context, recipeID string, embedding []float32) error {
	query := `
		INSERT INTO recipe_embeddings (recipe_id, embedding)
		VALUES ($1, $2)
		ON CONFLICT (recipe_id) DO UPDATE SET embedding = EXCLUDED.embedding
	`
	embStr := formatVector(embedding)
	_, err := db.Pool.Exec(ctx, query, recipeID, embStr)
	return err
}

func formatVector(vec []float32) string {
	s := "["
	for i, v := range vec {
		if i > 0 {
			s += ","
		}
		s += fmt.Sprintf("%f", v)
	}
	s += "]"
	return s
}
