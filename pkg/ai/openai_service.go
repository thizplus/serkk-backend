package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenAIService handles communication with OpenAI API
type OpenAIService interface {
	GenerateContent(ctx context.Context, prompt string, maxTokens int) (*GeneratedContent, error)
	GeneratePostContent(ctx context.Context, topic string) (*GeneratedContent, error)
	GeneratePostContentWithStyle(ctx context.Context, topic string, tone string, enableVariations bool) (*GeneratedContent, error)
	GenerateTitleVariations(ctx context.Context, topic string, count int, tone string) ([]string, error)
	GenerateBatchPosts(ctx context.Context, topics []string, tone string) ([]*GeneratedContent, error)
}

type openAIClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

type GeneratedContent struct {
	Title            string
	Content          string
	Tags             []string
	TitleVariations  []string          // Multiple title options
	Metadata         map[string]interface{} // Tone, style, etc.
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type openAIErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// NewOpenAIService creates a new OpenAI service instance
func NewOpenAIService(apiKey string, model string) OpenAIService {
	if model == "" {
		model = "gpt-4o-mini" // Default to cost-effective model
	}

	return &openAIClient{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GenerateContent generates content based on a custom prompt
func (c *openAIClient) GenerateContent(ctx context.Context, prompt string, maxTokens int) (*GeneratedContent, error) {
	if c.apiKey == "" {
		return nil, errors.New("OpenAI API key is not configured")
	}

	if maxTokens == 0 {
		maxTokens = 1000
	}

	reqBody := openAIRequest{
		Model: c.model,
		Messages: []openAIMessage{
			{
				Role:    "system",
				Content: "You are a helpful social media content creator. Generate engaging posts with titles, content, and relevant tags.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   maxTokens,
		Temperature: 0.8,
	}

	response, err := c.makeRequest(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	// Parse the response and extract title, content, and tags
	return c.parseGeneratedContent(response)
}

// GeneratePostContent generates a complete post based on a topic
func (c *openAIClient) GeneratePostContent(ctx context.Context, topic string) (*GeneratedContent, error) {
	prompt := fmt.Sprintf(`Create an engaging social media post about: "%s"

Please provide the response in the following JSON format:
{
  "title": "Post title (max 300 characters)",
  "content": "Post content (engaging and informative, 200-500 words)",
  "tags": ["tag1", "tag2", "tag3"]
}

Make the content engaging, informative, and suitable for a social media platform. Include 3-5 relevant hashtags as tags.`, topic)

	return c.GenerateContent(ctx, prompt, 1500)
}

// makeRequest makes HTTP request to OpenAI API
func (c *openAIClient) makeRequest(ctx context.Context, reqBody openAIRequest) (string, error) {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp openAIErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil {
			return "", fmt.Errorf("OpenAI API error: %s", errResp.Error.Message)
		}
		return "", fmt.Errorf("OpenAI API returned status %d: %s", resp.StatusCode, string(body))
	}

	var openAIResp openAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return "", errors.New("no response generated from OpenAI")
	}

	return openAIResp.Choices[0].Message.Content, nil
}

// parseGeneratedContent parses the AI response into structured content
func (c *openAIClient) parseGeneratedContent(response string) (*GeneratedContent, error) {
	var result struct {
		Title   string   `json:"title"`
		Content string   `json:"content"`
		Tags    []string `json:"tags"`
	}

	// Try to parse as JSON first
	if err := json.Unmarshal([]byte(response), &result); err == nil {
		return &GeneratedContent{
			Title:   result.Title,
			Content: result.Content,
			Tags:    result.Tags,
		}, nil
	}

	// If not JSON, try to extract from markdown code blocks
	content := response
	if start := bytes.Index([]byte(content), []byte("```json")); start != -1 {
		start += 7
		if end := bytes.Index([]byte(content[start:]), []byte("```")); end != -1 {
			jsonContent := content[start : start+end]
			if err := json.Unmarshal([]byte(jsonContent), &result); err == nil {
				return &GeneratedContent{
					Title:   result.Title,
					Content: result.Content,
					Tags:    result.Tags,
				}, nil
			}
		}
	}

	// Fallback: return raw content
	return &GeneratedContent{
		Title:   "AI Generated Post",
		Content: response,
		Tags:    []string{"ai-generated"},
	}, nil
}

// GeneratePostContentWithStyle generates content with specific tone and variations
func (c *openAIClient) GeneratePostContentWithStyle(ctx context.Context, topic string, tone string, enableVariations bool) (*GeneratedContent, error) {
	toneInstructions := map[string]string{
		"neutral":       "in a balanced, informative tone",
		"casual":        "in a friendly, conversational tone with casual language",
		"professional":  "in a professional, business-appropriate tone",
		"humorous":      "in a humorous, entertaining tone with wit and humor",
		"controversial": "in a provocative, thought-provoking tone that challenges assumptions",
	}

	toneInstruction := toneInstructions[tone]
	if toneInstruction == "" {
		toneInstruction = toneInstructions["neutral"]
	}

	variationRequest := ""
	if enableVariations {
		variationRequest = `  "titleVariations": ["variation 1", "variation 2", "variation 3", "variation 4", "variation 5"],`
	}

	prompt := fmt.Sprintf(`Create an engaging social media post %s about: "%s"

Please provide the response in the following JSON format:
{
  "title": "Main title (catchy, max 300 characters)",
%s
  "content": "Post content (engaging and informative, 200-500 words)",
  "tags": ["tag1", "tag2", "tag3", "tag4", "tag5"]
}

Guidelines:
- Make the title attention-grabbing and unique
- Use emojis sparingly but effectively
- Include relevant statistics or numbers if appropriate
- Make content scannable with short paragraphs
- Use power words and emotional triggers
- Include 3-5 relevant hashtags as tags
%s`, toneInstruction, topic, variationRequest, getStyleGuidelines(tone))

	return c.GenerateContent(ctx, prompt, 1800)
}

// GenerateTitleVariations generates multiple title options for a topic
func (c *openAIClient) GenerateTitleVariations(ctx context.Context, topic string, count int, tone string) ([]string, error) {
	if count <= 0 {
		count = 5
	}
	if count > 20 {
		count = 20
	}

	toneDesc := map[string]string{
		"neutral":       "balanced and informative",
		"casual":        "friendly and conversational",
		"professional":  "professional and business-appropriate",
		"humorous":      "funny and entertaining",
		"controversial": "provocative and thought-provoking",
	}[tone]

	if toneDesc == "" {
		toneDesc = "engaging"
	}

	prompt := fmt.Sprintf(`Generate %d unique and %s title variations for a social media post about: "%s"

Requirements:
- Each title should be unique and different from the others
- Maximum 300 characters per title
- Use different approaches: questions, statements, statistics, emotional triggers
- Include emojis where appropriate (but not in all titles)
- Make them click-worthy and shareable

Return ONLY a JSON array of strings:
["title 1", "title 2", "title 3", ...]`, count, toneDesc, topic)

	reqBody := openAIRequest{
		Model: c.model,
		Messages: []openAIMessage{
			{
				Role:    "system",
				Content: "You are an expert social media copywriter who creates viral, engaging titles.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   500,
		Temperature: 0.9, // Higher temperature for more variety
	}

	response, err := c.makeRequest(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	// Parse response as JSON array
	var titles []string
	if err := json.Unmarshal([]byte(response), &titles); err == nil {
		return titles, nil
	}

	// Try to extract from markdown code blocks
	if start := bytes.Index([]byte(response), []byte("```json")); start != -1 {
		start += 7
		if end := bytes.Index([]byte(response[start:]), []byte("```")); end != -1 {
			jsonContent := response[start : start+end]
			if err := json.Unmarshal([]byte(jsonContent), &titles); err == nil {
				return titles, nil
			}
		}
	}

	// Fallback: return a single title
	return []string{fmt.Sprintf("📢 %s: What You Need to Know", topic)}, nil
}

// GenerateBatchPosts generates multiple posts at once for efficiency
func (c *openAIClient) GenerateBatchPosts(ctx context.Context, topics []string, tone string) ([]*GeneratedContent, error) {
	if len(topics) == 0 {
		return nil, errors.New("no topics provided")
	}

	if len(topics) > 10 {
		topics = topics[:10] // Limit to 10 posts per batch
	}

	topicList := ""
	for i, topic := range topics {
		topicList += fmt.Sprintf("%d. %s\n", i+1, topic)
	}

	toneDesc := map[string]string{
		"neutral":       "balanced and informative",
		"casual":        "friendly and conversational",
		"professional":  "professional and business-appropriate",
		"humorous":      "funny and entertaining",
		"controversial": "provocative and thought-provoking",
	}[tone]

	if toneDesc == "" {
		toneDesc = "engaging"
	}

	prompt := fmt.Sprintf(`Generate %d unique %s social media posts for the following topics:

%s

For each topic, provide:
- A unique, catchy title (max 300 characters)
- Engaging content (200-500 words)
- 3-5 relevant tags

Return a JSON array with this structure:
[
  {
    "topic": "topic 1",
    "title": "title for topic 1",
    "content": "content for topic 1",
    "tags": ["tag1", "tag2", "tag3"]
  },
  ...
]

Make each post unique and avoid repetitive patterns.`, len(topics), toneDesc, topicList)

	reqBody := openAIRequest{
		Model: c.model,
		Messages: []openAIMessage{
			{
				Role:    "system",
				Content: "You are an expert social media content creator who generates diverse, engaging posts efficiently.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   len(topics) * 600, // Allocate tokens per post
		Temperature: 0.8,
	}

	response, err := c.makeRequest(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	// Parse response
	type BatchPost struct {
		Topic   string   `json:"topic"`
		Title   string   `json:"title"`
		Content string   `json:"content"`
		Tags    []string `json:"tags"`
	}

	var batchPosts []BatchPost
	if err := json.Unmarshal([]byte(response), &batchPosts); err == nil {
		results := make([]*GeneratedContent, len(batchPosts))
		for i, post := range batchPosts {
			results[i] = &GeneratedContent{
				Title:   post.Title,
				Content: post.Content,
				Tags:    post.Tags,
				Metadata: map[string]interface{}{
					"topic": post.Topic,
					"tone":  tone,
				},
			}
		}
		return results, nil
	}

	// Try to extract from markdown
	if start := bytes.Index([]byte(response), []byte("```json")); start != -1 {
		start += 7
		if end := bytes.Index([]byte(response[start:]), []byte("```")); end != -1 {
			jsonContent := response[start : start+end]
			if err := json.Unmarshal([]byte(jsonContent), &batchPosts); err == nil {
				results := make([]*GeneratedContent, len(batchPosts))
				for i, post := range batchPosts {
					results[i] = &GeneratedContent{
						Title:   post.Title,
						Content: post.Content,
						Tags:    post.Tags,
						Metadata: map[string]interface{}{
							"topic": post.Topic,
							"tone":  tone,
						},
					}
				}
				return results, nil
			}
		}
	}

	// Fallback: generate posts one by one
	results := make([]*GeneratedContent, 0, len(topics))
	for _, topic := range topics {
		content, err := c.GeneratePostContentWithStyle(ctx, topic, tone, false)
		if err != nil {
			continue // Skip failed posts
		}
		results = append(results, content)
	}

	return results, nil
}

// getStyleGuidelines returns style-specific guidelines
func getStyleGuidelines(tone string) string {
	guidelines := map[string]string{
		"controversial": `
Additional guidelines for controversial tone:
- Use rhetorical questions that challenge the status quo
- Present contrarian viewpoints
- Use phrases like "unpopular opinion", "hot take", "let's be honest"
- Create debate-worthy statements
- Use statistics that surprise or provoke`,
		"humorous": `
Additional guidelines for humorous tone:
- Use wordplay and puns where appropriate
- Include relatable situations
- Use self-deprecating humor
- Add unexpected twists
- Keep it light and fun`,
		"casual": `
Additional guidelines for casual tone:
- Use contractions and colloquial language
- Include personal anecdotes or examples
- Use "you" to address readers directly
- Keep sentences short and punchy`,
		"professional": `
Additional guidelines for professional tone:
- Use industry-specific terminology appropriately
- Include data and evidence
- Maintain credibility and authority
- Focus on actionable insights`,
	}

	if guide, exists := guidelines[tone]; exists {
		return guide
	}
	return ""
}
