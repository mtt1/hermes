// Package ai - Gemini API client
package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/genai"
	"hermes/internal/safety"
)

// GeminiClient implements the Client interface for Google's Gemini API
type GeminiClient struct {
	config Config
	client *genai.Client
}

// geminiResponse represents the structured JSON response from Gemini API
type geminiResponse struct {
	Command     string `json:"command"`
	Safety      string `json:"safety"`
	Explanation string `json:"explanation"`
}

// ExplanationSection represents a section of the structured explanation
type ExplanationSection struct {
	Text    string   `json:"text"`
	Details []string `json:"details"`
}

// NewGeminiClient creates a new Gemini API client using the official Google Gen AI SDK
func NewGeminiClient(config Config) (*GeminiClient, error) {
	// API key presence is validated before creating the client
	ctx := context.Background()
	
	// Initialize the official Google Gen AI client
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  config.APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GeminiClient{
		config: config,
		client: client,
	}, nil
}

// GenerateCommand generates a shell command from natural language
func (g *GeminiClient) GenerateCommand(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	prompt := g.buildGeneratePrompt(req.Query)
	
	// Select model - use Flash for speed, Pro for quality
	modelName := "gemini-2.0-flash"
	if g.config.Model != "" {
		modelName = g.config.Model
	}
	
	// Create parts for the request
	parts := []*genai.Part{
		{Text: prompt},
	}
	content := []*genai.Content{{Parts: parts}}
	
	resp, err := g.client.Models.GenerateContent(ctx, modelName, content, nil)
	if err != nil {
		return nil, err // Fail fast and transparent
	}
	
	return g.parseGenerateResponse(resp)
}

// ExplainCommand explains what a shell command does
func (g *GeminiClient) ExplainCommand(ctx context.Context, req ExplainRequest) (*ExplainResponse, error) {
	prompt := g.buildExplainPrompt(req.Command)
	
	// Select model - use Flash for speed, Pro for quality
	modelName := "gemini-2.0-flash"
	if g.config.Model != "" {
		modelName = g.config.Model
	}
	
	// Create parts for the request
	parts := []*genai.Part{
		{Text: prompt},
	}
	content := []*genai.Content{{Parts: parts}}
	
	resp, err := g.client.Models.GenerateContent(ctx, modelName, content, nil)
	if err != nil {
		return nil, err // Fail fast and transparent
	}
	
	return g.parseExplainResponse(resp)
}

// Close cleans up any resources used by the client
func (g *GeminiClient) Close() error {
	// The genai client doesn't have a Close method, so we do nothing
	return nil
}

// buildGeneratePrompt creates the prompt for command generation
func (g *GeminiClient) buildGeneratePrompt(query string) string {
	return fmt.Sprintf(`You are an expert system administrator that translates natural language queries into shell commands.

Your response MUST be a valid JSON object with exactly this schema:
{
  "command": "<the generated shell command>",
  "safety": "<SAFE | ATTENTION>",
  "explanation": "<brief explanation of the command and safety reasoning>"
}

Safety Guidelines:
- SAFE: Read-only operations, basic file listing, navigation, help commands
- ATTENTION: File modifications, system changes, network operations, anything requiring sudo

Important Rules:
1. Generate the EXACT command needed, no explanations outside the JSON
2. Commands should be compatible with bash/zsh
3. Use standard Unix utilities when possible
4. Be conservative with safety assessment - prefer ATTENTION when uncertain

User Query: %s`, query)
}

// buildExplainPrompt creates the prompt for command explanation
func (g *GeminiClient) buildExplainPrompt(command string) string {
	return fmt.Sprintf(`You are an expert system administrator. Explain this shell command in a structured, educational format.

Your response MUST be a valid JSON object with exactly this schema:
{
  "explanation": [
    {
      "text": "main command or section description",
      "details": ["flag explanations", "option explanations"]
    }
  ]
}

Structure Guidelines:
- Each main command/section gets its own object in the explanation array
- Put the main description in "text" field
- Put flag/option explanations in "details" array
- For piped commands, separate each part into different objects
- Use clear, educational language

Command to explain: %s`, command)
}

// parseGenerateResponse parses the JSON response from the generate API
func (g *GeminiClient) parseGenerateResponse(resp *genai.GenerateContentResponse) (*GenerateResponse, error) {
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no content returned from API")
	}

	// Extract and parse JSON response
	jsonText := resp.Candidates[0].Content.Parts[0].Text
	if jsonText == "" {
		return nil, fmt.Errorf("empty response text")
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal([]byte(jsonText), &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Convert safety level
	var safetyLevel safety.SafetyLevel
	switch geminiResp.Safety {
	case "SAFE":
		safetyLevel = safety.Safe
	case "ATTENTION":
		safetyLevel = safety.Attention
	default:
		safetyLevel = safety.Attention // Default to attention for unknown values
	}

	return &GenerateResponse{
		Command:     geminiResp.Command,
		SafetyLevel: safetyLevel,
		Reasoning:   geminiResp.Explanation,
	}, nil
}

// parseExplainResponse parses the JSON response from the explain API
func (g *GeminiClient) parseExplainResponse(resp *genai.GenerateContentResponse) (*ExplainResponse, error) {
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no content returned from API")
	}

	jsonText := resp.Candidates[0].Content.Parts[0].Text
	if jsonText == "" {
		return nil, fmt.Errorf("empty response text")
	}

	var explainResp struct {
		Explanation []ExplanationSection `json:"explanation"`
	}
	
	if err := json.Unmarshal([]byte(jsonText), &explainResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Format the structured explanation into bullet points
	explanation := g.formatExplanation(explainResp.Explanation)

	return &ExplainResponse{
		Explanation: explanation,
	}, nil
}

// formatExplanation converts structured explanation to bullet point format
func (g *GeminiClient) formatExplanation(sections []ExplanationSection) string {
	var result string
	
	for _, section := range sections {
		result += fmt.Sprintf("• %s\n", section.Text)
		for _, detail := range section.Details {
			result += fmt.Sprintf("  • %s\n", detail)
		}
	}
	
	return result
}