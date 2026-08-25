package mcp

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/NivRave/socialfoodie/backend/internal/db"
	"github.com/NivRave/socialfoodie/backend/internal/llm"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Handlers struct {
	db  *db.DB
	llm *llm.Client
}

func NewHandlers(db *db.DB, llmClient *llm.Client) *Handlers {
	return &Handlers{db: db, llm: llmClient}
}

func (h *Handlers) SearchByIngredientsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("invalid arguments format"), nil
	}
	ingredientsStr, ok := args["ingredients"].(string)
	if !ok {
		return mcp.NewToolResultError("ingredients must be a string"), nil
	}

	limit, ok := args["limit"].(float64)
	if !ok || limit <= 0 {
		limit = 10
	}

	parts := strings.Split(ingredientsStr, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	results, err := h.db.SearchByIngredients(ctx, parts, int(limit), 0)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return formatResults(results), nil
}

func (h *Handlers) SemanticSearchHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("invalid arguments format"), nil
	}
	query, ok := args["query"].(string)
	if !ok {
		return mcp.NewToolResultError("query must be a string"), nil
	}

	limit, ok := args["limit"].(float64)
	if !ok || limit <= 0 {
		limit = 10
	}

	embedding, err := h.llm.GenerateEmbedding(ctx, query)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	results, err := h.db.SemanticRecipeSearch(ctx, embedding, int(limit), 0)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return formatResults(results), nil
}

func (h *Handlers) GetRecipesByTagsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("invalid arguments format"), nil
	}
	tagsStr, ok := args["tags"].(string)
	if !ok {
		return mcp.NewToolResultError("tags must be a string"), nil
	}

	limit, ok := args["limit"].(float64)
	if !ok || limit <= 0 {
		limit = 10
	}

	parts := strings.Split(tagsStr, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	results, err := h.db.GetRecipesByTags(ctx, parts, int(limit), 0)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return formatResults(results), nil
}

func formatResults(results []db.RecipeResult) *mcp.CallToolResult {
	if len(results) == 0 {
		return mcp.NewToolResultText("No recipes found.")
	}

	b, _ := json.MarshalIndent(results, "", "  ")
	return mcp.NewToolResultText(string(b))
}

func RegisterTools(s *server.MCPServer, h *Handlers) {
	searchByIng := mcp.NewTool("search_by_ingredients",
		mcp.WithDescription("Search recipes by a list of ingredients"),
		mcp.WithString("ingredients", mcp.Required(), mcp.Description("Comma separated ingredients")),
		mcp.WithNumber("limit", mcp.Description("Max results")),
	)
	s.AddTool(searchByIng, h.SearchByIngredientsHandler)

	semanticSearch := mcp.NewTool("semantic_recipe_search",
		mcp.WithDescription("Semantic vector search for recipes using a natural language query"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query (e.g. 'a healthy chicken dish')")),
		mcp.WithNumber("limit", mcp.Description("Max results")),
	)
	s.AddTool(semanticSearch, h.SemanticSearchHandler)

	tagSearch := mcp.NewTool("get_recipes_by_tags",
		mcp.WithDescription("Search recipes by meal tags"),
		mcp.WithString("tags", mcp.Required(), mcp.Description("Comma separated tags (e.g. 'vegan, gluten-free')")),
		mcp.WithNumber("limit", mcp.Description("Max results")),
	)
	s.AddTool(tagSearch, h.GetRecipesByTagsHandler)
}
