// Package ai - Gemini API client
package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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
	modelName := "gemini-2.5-flash"
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
	modelName := "gemini-2.5-flash"
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

CRITICAL: Your response MUST be ONLY a valid JSON object. Do NOT wrap it in markdown code blocks. Do NOT add any text before or after the JSON.

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
5. RESPOND WITH ONLY JSON - NO MARKDOWN, NO BACKTICKS, NO EXTRA TEXT

User Query: %s`, query)
}

// buildExplainPrompt creates the prompt for command explanation
func (g *GeminiClient) buildExplainPrompt(command string) string {
	return fmt.Sprintf(`You are an expert system administrator. Explain this shell command in a structured, educational format.

CRITICAL: Your response MUST be ONLY a valid JSON object. Do NOT wrap it in markdown code blocks. Do NOT add any text before or after the JSON.

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
- RESPOND WITH ONLY JSON - NO MARKDOWN, NO BACKTICKS, NO EXTRA TEXT

Command to explain: %s`, command)
}

// parseGenerateResponse parses the JSON response from the generate API
func (g *GeminiClient) parseGenerateResponse(resp *genai.GenerateContentResponse) (*GenerateResponse, error) {
	// Debug output if enabled - show complete response structure
	if g.config.Debug {
		fmt.Printf("DEBUG: === FULL API RESPONSE STRUCTURE ===\n")
		fmt.Printf("DEBUG: Number of candidates: %d\n", len(resp.Candidates))
		for i, candidate := range resp.Candidates {
			fmt.Printf("DEBUG: Candidate %d:\n", i)
			fmt.Printf("DEBUG:   Number of parts: %d\n", len(candidate.Content.Parts))
			for j, part := range candidate.Content.Parts {
				fmt.Printf("DEBUG:   Part %d text: %q\n", j, part.Text)
			}
		}
		fmt.Printf("DEBUG: === END API RESPONSE STRUCTURE ===\n")
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no content returned from API")
	}

	// Extract and parse JSON response
	jsonText := resp.Candidates[0].Content.Parts[0].Text
	if jsonText == "" {
		return nil, fmt.Errorf("empty response text")
	}

	if g.config.Debug {
		fmt.Printf("DEBUG: jsonText we're trying to parse:\n%s\n", jsonText)
		fmt.Printf("DEBUG: === END jsonText ===\n")
	}

	// Clean up the response - remove markdown code blocks if present
	cleanedJSON := cleanJSONResponse(jsonText)
	
	if g.config.Debug {
		fmt.Printf("DEBUG: cleanedJSON after removing markdown:\n%s\n", cleanedJSON)
		fmt.Printf("DEBUG: === END cleanedJSON ===\n")
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal([]byte(cleanedJSON), &geminiResp); err != nil {
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
	// Debug output if enabled - show complete response structure
	if g.config.Debug {
		fmt.Printf("DEBUG: === FULL API RESPONSE STRUCTURE ===\n")
		fmt.Printf("DEBUG: Number of candidates: %d\n", len(resp.Candidates))
		for i, candidate := range resp.Candidates {
			fmt.Printf("DEBUG: Candidate %d:\n", i)
			fmt.Printf("DEBUG:   Number of parts: %d\n", len(candidate.Content.Parts))
			for j, part := range candidate.Content.Parts {
				fmt.Printf("DEBUG:   Part %d text: %q\n", j, part.Text)
			}
		}
		fmt.Printf("DEBUG: === END API RESPONSE STRUCTURE ===\n")
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no content returned from API")
	}

	jsonText := resp.Candidates[0].Content.Parts[0].Text
	if jsonText == "" {
		return nil, fmt.Errorf("empty response text")
	}

	if g.config.Debug {
		fmt.Printf("DEBUG: jsonText we're trying to parse:\n%s\n", jsonText)
		fmt.Printf("DEBUG: === END jsonText ===\n")
	}

	// Clean up the response - remove markdown code blocks if present
	cleanedJSON := cleanJSONResponse(jsonText)
	
	if g.config.Debug {
		fmt.Printf("DEBUG: cleanedJSON after removing markdown:\n%s\n", cleanedJSON)
		fmt.Printf("DEBUG: === END cleanedJSON ===\n")
	}

	var explainResp struct {
		Explanation []ExplanationSection `json:"explanation"`
	}
	
	if err := json.Unmarshal([]byte(cleanedJSON), &explainResp); err != nil {
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

// cleanJSONResponse removes markdown code block formatting from API responses
func cleanJSONResponse(text string) string {
	// Remove markdown code blocks (```json ... ``` or ``` ... ```)
	text = strings.TrimSpace(text)
	
	// Check for and remove ```json prefix
	if strings.HasPrefix(text, "```json") {
		text = strings.TrimPrefix(text, "```json")
		text = strings.TrimSpace(text)
	}
	
	// Check for and remove ``` prefix (without json)
	if strings.HasPrefix(text, "```") {
		text = strings.TrimPrefix(text, "```")
		text = strings.TrimSpace(text)
	}
	
	// Check for and remove ``` suffix
	if strings.HasSuffix(text, "```") {
		text = strings.TrimSuffix(text, "```")
		text = strings.TrimSpace(text)
	}
	
	return text
}
