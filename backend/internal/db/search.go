package db

import (
	"context"
)

type RecipeResult struct {
	ID           string
	Name         *string
	URL          string
	Instructions *string
}

func (db *DB) SearchByIngredients(ctx context.Context, ingredients []string, limit, offset int) ([]RecipeResult, error) {
	query := `
		SELECT DISTINCT r.id, r.name, r.url, r.instructions
		FROM recipes r
		JOIN recipe_ingredients ri ON ri.recipe_id = r.id
		JOIN ingredients i ON ri.ingredient_id = i.id
		WHERE i.name ILIKE ANY($1)
		LIMIT $2 OFFSET $3
	`

	likes := make([]string, len(ingredients))
	for i, ing := range ingredients {
		likes[i] = "%" + ing + "%"
	}

	rows, err := db.Pool.Query(ctx, query, likes, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []RecipeResult
	for rows.Next() {
		var r RecipeResult
		if err := rows.Scan(&r.ID, &r.Name, &r.URL, &r.Instructions); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, nil
}

func (db *DB) SemanticRecipeSearch(ctx context.Context, queryEmbedding []float32, limit, offset int) ([]RecipeResult, error) {
	query := `
		SELECT r.id, r.name, r.url, r.instructions
		FROM recipes r
		JOIN recipe_embeddings re ON re.recipe_id = r.id
		ORDER BY re.embedding <=> $1
		LIMIT $2 OFFSET $3
	`
	embStr := formatVector(queryEmbedding)

	rows, err := db.Pool.Query(ctx, query, embStr, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []RecipeResult
	for rows.Next() {
		var r RecipeResult
		if err := rows.Scan(&r.ID, &r.Name, &r.URL, &r.Instructions); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, nil
}

func (db *DB) GetRecipesByTags(ctx context.Context, tags []string, limit, offset int) ([]RecipeResult, error) {
	query := `
		SELECT DISTINCT r.id, r.name, r.url, r.instructions
		FROM recipes r
		JOIN recipe_tags rt ON rt.recipe_id = r.id
		WHERE rt.tag ILIKE ANY($1)
		LIMIT $2 OFFSET $3
	`
	likes := make([]string, len(tags))
	for i, t := range tags {
		likes[i] = "%" + t + "%"
	}

	rows, err := db.Pool.Query(ctx, query, likes, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []RecipeResult
	for rows.Next() {
		var r RecipeResult
		if err := rows.Scan(&r.ID, &r.Name, &r.URL, &r.Instructions); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, nil
}
