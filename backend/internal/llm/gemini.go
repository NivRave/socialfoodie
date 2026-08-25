package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"google.golang.org/genai"
)

type Client struct {
	client *genai.Client
}

func NewClient(ctx context.Context) (*Client, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is not set")
	}

	client, err := genai.NewClient(ctx, nil) // Uses GEMINI_API_KEY automatically
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	return &Client{client: client}, nil
}

type ExtractedRecipe struct {
	Name         string `json:"name"`
	Instructions string `json:"instructions"`
	Difficulty   string `json:"difficulty"`
	PrepTime     int    `json:"prep_time_minutes"`
	Ingredients  []struct {
		Name     string `json:"name"`
		Quantity string `json:"quantity"`
		Unit     string `json:"unit"`
	} `json:"ingredients"`
	Tags []struct {
		Tag       string `json:"tag"`
		Reasoning string `json:"reasoning"`
	} `json:"tags"`
}

func (c *Client) ExtractRecipe(ctx context.Context, rawText string) (*ExtractedRecipe, error) {
	prompt := fmt.Sprintf(`Extract the recipe details from the following raw text. 
Return ONLY a valid JSON object matching the requested schema.

Raw Text: %s`, rawText)

	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"name":              {Type: genai.TypeString},
				"instructions":      {Type: genai.TypeString},
				"difficulty":        {Type: genai.TypeString},
				"prep_time_minutes": {Type: genai.TypeInteger},
				"ingredients": {
					Type: genai.TypeArray,
					Items: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"name":     {Type: genai.TypeString},
							"quantity": {Type: genai.TypeString},
							"unit":     {Type: genai.TypeString},
						},
					},
				},
				"tags": {
					Type: genai.TypeArray,
					Items: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"tag":       {Type: genai.TypeString},
							"reasoning": {Type: genai.TypeString},
						},
					},
				},
			},
		},
	}

	resp, err := c.client.Models.GenerateContent(ctx, "gemini-3.6-flash", genai.Text(prompt), config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from Gemini")
	}

	part := resp.Candidates[0].Content.Parts[0]
	var jsonText string
	if part.Text != "" {
		jsonText = part.Text
	} else {
		return nil, fmt.Errorf("unexpected part type or empty text")
	}

	var recipe ExtractedRecipe
	if err := json.Unmarshal([]byte(jsonText), &recipe); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w\nResponse was: %s", err, jsonText)
	}

	return &recipe, nil
}

func (c *Client) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	dim := int32(768)
	config := &genai.EmbedContentConfig{
		OutputDimensionality: &dim,
	}
	resp, err := c.client.Models.EmbedContent(ctx, "gemini-embedding-2", genai.Text(text), config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	if len(resp.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return resp.Embeddings[0].Values, nil
}
