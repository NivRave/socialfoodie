package llm

import (
	"encoding/json"
	"testing"
)

func TestExtractedRecipeUnmarshal(t *testing.T) {
	mockGeminiOutput := `{
		"name": "Chocolate Chip Cookies",
		"instructions": "Mix ingredients. Bake at 350 for 10 mins.",
		"difficulty": "Easy",
		"prep_time_minutes": 15,
		"ingredients": [
			{
				"name": "flour",
				"quantity": "2",
				"unit": "cups"
			},
			{
				"name": "chocolate chips",
				"quantity": "1",
				"unit": "cup"
			}
		],
		"tags": [
			{
				"tag": "dessert",
				"reasoning": "contains sugar and chocolate"
			}
		]
	}`

	var recipe ExtractedRecipe
	err := json.Unmarshal([]byte(mockGeminiOutput), &recipe)
	if err != nil {
		t.Fatalf("failed to unmarshal mock gemini output: %v", err)
	}

	if recipe.Name != "Chocolate Chip Cookies" {
		t.Errorf("Expected name 'Chocolate Chip Cookies', got '%s'", recipe.Name)
	}
	if recipe.PrepTime != 15 {
		t.Errorf("Expected prep time 15, got %d", recipe.PrepTime)
	}
	if len(recipe.Ingredients) != 2 {
		t.Errorf("Expected 2 ingredients, got %d", len(recipe.Ingredients))
	}
	if recipe.Ingredients[0].Name != "flour" {
		t.Errorf("Expected first ingredient 'flour', got '%s'", recipe.Ingredients[0].Name)
	}
	if len(recipe.Tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(recipe.Tags))
	}
	if recipe.Tags[0].Tag != "dessert" {
		t.Errorf("Expected tag 'dessert', got '%s'", recipe.Tags[0].Tag)
	}
}
