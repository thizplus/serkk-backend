package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
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
	Model              string          `json:"model"`
	Messages           []openAIMessage `json:"messages"`
	MaxCompletionTokens int            `json:"max_completion_tokens,omitempty"`
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
		MaxCompletionTokens: maxTokens,
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
		Title      string   `json:"title"`
		Content    string   `json:"content"`     // Old format (backward compatible)
		Paragraphs []string `json:"paragraphs"` // New format
		Tags       []string `json:"tags"`
	}

	// Clean response (trim whitespace)
	cleanedResponse := strings.TrimSpace(response)

	// Try 1: Parse as JSON directly
	if err := json.Unmarshal([]byte(cleanedResponse), &result); err == nil {
		content := result.Content
		// If paragraphs provided, join with double newlines
		if len(result.Paragraphs) > 0 {
			content = strings.Join(result.Paragraphs, "\n\n")
		}
		return &GeneratedContent{
			Title:   result.Title,
			Content: content,
			Tags:    result.Tags,
		}, nil
	}

	// Try 2: Extract from ```json code block
	if start := bytes.Index([]byte(cleanedResponse), []byte("```json")); start != -1 {
		start += 7
		if end := bytes.Index([]byte(cleanedResponse[start:]), []byte("```")); end != -1 {
			jsonContent := bytes.TrimSpace([]byte(cleanedResponse[start : start+end]))
			if err := json.Unmarshal(jsonContent, &result); err == nil {
				content := result.Content
				if len(result.Paragraphs) > 0 {
					content = strings.Join(result.Paragraphs, "\n\n")
				}
				return &GeneratedContent{
					Title:   result.Title,
					Content: content,
					Tags:    result.Tags,
				}, nil
			}
		}
	}

	// Try 3: Extract from ``` code block (without json marker)
	if start := bytes.Index([]byte(cleanedResponse), []byte("```")); start != -1 {
		start += 3
		if end := bytes.Index([]byte(cleanedResponse[start:]), []byte("```")); end != -1 {
			jsonContent := bytes.TrimSpace([]byte(cleanedResponse[start : start+end]))
			if err := json.Unmarshal(jsonContent, &result); err == nil {
				content := result.Content
				if len(result.Paragraphs) > 0 {
					content = strings.Join(result.Paragraphs, "\n\n")
				}
				return &GeneratedContent{
					Title:   result.Title,
					Content: content,
					Tags:    result.Tags,
				}, nil
			}
		}
	}

	// Try 4: Find JSON object with regex pattern
	startIdx := bytes.Index([]byte(cleanedResponse), []byte("{"))
	endIdx := bytes.LastIndex([]byte(cleanedResponse), []byte("}"))
	if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
		jsonContent := bytes.TrimSpace([]byte(cleanedResponse[startIdx : endIdx+1]))
		if err := json.Unmarshal(jsonContent, &result); err == nil {
			content := result.Content
			if len(result.Paragraphs) > 0 {
				content = strings.Join(result.Paragraphs, "\n\n")
			}
			return &GeneratedContent{
				Title:   result.Title,
				Content: content,
				Tags:    result.Tags,
			}, nil
		}
	}

	// If all parsing failed, return error instead of raw content
	return nil, fmt.Errorf("failed to parse AI response as JSON. Response: %s", cleanedResponse)
}

// GeneratePostContentWithStyle generates content with specific tone and variations
func (c *openAIClient) GeneratePostContentWithStyle(ctx context.Context, topic string, tone string, enableVariations bool) (*GeneratedContent, error) {
	// Tone descriptions for human-like writing
	toneInstructions := map[string]string{
		"neutral":       "balanced, informative",
		"casual":        "friendly, conversational",
		"professional":  "professional, business-appropriate",
		"humorous":      "humorous, entertaining with wit",
		"controversial": "provocative, thought-provoking",
	}

	toneInstruction := toneInstructions[tone]
	if toneInstruction == "" {
		toneInstruction = toneInstructions["neutral"]
	}

	// Persona based on tone
	personaMap := map[string]string{
		"neutral":       "Thai observer with balanced perspective",
		"casual":        "Thai friend sharing thoughts over coffee",
		"professional":  "Thai business columnist",
		"humorous":      "Thai comedian with social commentary",
		"controversial": "Thai columnist satirical - sharp, surgical, witty",
	}

	persona := personaMap[tone]
	if persona == "" {
		persona = personaMap["neutral"]
	}

	prompt := fmt.Sprintf(`%s

Persona: %s
Tone: %s
Topic: "%s"

Write a social media post using HUMAN cadence.

Return JSON:
{
  "title": "...",
  "paragraphs": ["paragraph 1", "paragraph 2", "paragraph 3"],
  "tags": ["...", "...", "..."]
}

CONTENT RULES:
- Must feel like written by a human with emotion + opinion
- No repeated sentence patterns
- No generic AI phrases
- Use rhetorical hooks, tension, punchlines
- Use Thai linguistic flow, natural spacing, and character
- Include sharp insight that reflects human lived experience
- Use pattern-breaking phrasing
- Write in Thai language
- Title: catchy, max 200 characters
- Paragraphs: 4-8 paragraphs, vary length (some short punch lines, some longer)
- Tags: 3-5 keywords in Thai (NO # symbols, just plain words)

FORMAT EXAMPLE:
{
  "title": "ทำไมค่าทิปถึงทำให้คนทำงานรู้สึกเหมือนโดนฟาดแต่ต้องยิ้มรับ?",
  "paragraphs": [
    "ประเทศไทยนี่แปลกอย่าง—อะไรที่ควรเป็น \"ระบบ\" ดันถูกโยนให้เป็น \"น้ำใจ\" ของลูกค้าแทนซะงั้น",
    "ค่าทิปก็เหมือนกัน พูดตรงๆ มันดูดีเฉพาะบนกระดาษ",
    "คนเสิร์ฟต้องลุ้นทุกโต๊ะ จะได้ไหม? วันนี้จะมีดวงพอให้ค่าไฟไหม?",
    "ถึงเวลาพูดกันตรงๆ ว่า คนทำงานควรได้ \"ค่าจ้าง\" ไม่ใช่ \"ค่าลุ้น\""
  ],
  "tags": ["ค่าทิป", "ระบบแรงงาน", "เศรษฐกิจไทย"]
}

Start writing now.`, HumanWriterBlueprint, persona, toneInstruction, topic)

	return c.GenerateContent(ctx, prompt, 2000)
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
		MaxCompletionTokens: 500,
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
		MaxCompletionTokens: len(topics) * 600, // Allocate tokens per post
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
