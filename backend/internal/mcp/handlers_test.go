package mcp

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestSearchByIngredientsHandler_Validation(t *testing.T) {
	// We can test validation without needing a real DB by expecting argument errors
	h := &Handlers{} // DB is nil, but validation happens before DB is called

	ctx := context.Background()

	// Test 1: Invalid arguments format (nil arguments)
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "search_by_ingredients",
			Arguments: nil,
		},
	}

	result, err := h.SearchByIngredientsHandler(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error from handler, got %v", err)
	}
	if !result.IsError {
		t.Error("Expected error result for nil arguments")
	}

	// Test 2: Missing ingredients string
	req.Params.Arguments = map[string]interface{}{
		"limit": 10.0,
	}

	result, err = h.SearchByIngredientsHandler(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error from handler, got %v", err)
	}
	if !result.IsError {
		t.Error("Expected error result for missing ingredients")
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "must be a string") {
		t.Errorf("Expected specific error message, got %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestSemanticSearchHandler_Validation(t *testing.T) {
	h := &Handlers{}

	ctx := context.Background()
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "semantic_recipe_search",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := h.SemanticSearchHandler(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error from handler, got %v", err)
	}
	if !result.IsError {
		t.Error("Expected error result for missing query")
	}
}
