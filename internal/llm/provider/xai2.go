package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/opencode-ai/opencode/internal/config"
	"github.com/opencode-ai/opencode/internal/llm/models"
	"github.com/opencode-ai/opencode/internal/llm/tools"
	"github.com/opencode-ai/opencode/internal/logging"
	"github.com/opencode-ai/opencode/internal/message"
	"github.com/opencode-ai/opencode/internal/request"
)

const (
	xai2MaxRetries = 8
	// xAI specific error patterns
	quotaExceededPattern = "credits or reached its monthly spending limit"
)

// xai2ErrorPatterns contains patterns that indicate errors in content
var xai2ErrorPatterns = []string{
	"try again",
	"rate limit",
	"too many requests",
	"please retry",
	"service unavailable",
	"quota exceeded",
	"temporarily unavailable",
}

// xai2Client is a standalone xAI client that doesn't inherit from OpenAI
type xai2Client struct {
	client           openai.Client
	options          openaiOptions
	providerOptions  providerClientOptions
}

// xai2ContentError represents an error when xAI returns error messages as content
type xai2ContentError struct {
	StatusCode int
	Content    string
}

func (e *xai2ContentError) Error() string {
	return fmt.Sprintf("xAI returned error as content (status %d): %s", e.StatusCode, e.Content)
}

func NewXAI2Client(apiKey string, providerOptions providerClientOptions) (*xai2Client, error) {
	options := openaiOptions{
		baseURL: "https://api.x.ai/v1",
	}

	opts := []option.RequestOption{
		option.WithHeader("HTTP-Referer", "https://api.x.ai"),
	}

	if apiKey != "" {
		opts = append(opts, option.WithAPIKey(apiKey))
	}
	if options.baseURL != "" {
		opts = append(opts, option.WithBaseURL(options.baseURL))
	}

	client := openai.NewClient(opts...)

	return &xai2Client{
		client:          client,
		options:         options,
		providerOptions: providerOptions,
	}, nil
}

// Provider interface methods

func (x *xai2Client) SendMessages(ctx context.Context, messages []message.Message, tools []tools.BaseTool) (*ProviderResponse, error) {
	// Set current request info for display
	request.SetCurrent(string(x.providerOptions.model.Provider), x.providerOptions.model.APIModel, x.options.baseURL)
	
	openaiMessages := x.convertMessages(messages)
	openaiTools := x.convertTools(tools)
	
	params := x.preparedParams(openaiMessages, openaiTools)
	
	if x.isDebugMode() {
		x.logRequest()
	}
	
	completion, err := x.client.Chat.Completions.New(ctx, params)
	if err != nil {
		request.Clear()
		x.logError(err, params)
		return nil, err
	}
	
	request.Clear()
	response := x.toProviderResponse(*completion)
	return response, nil
}

func (x *xai2Client) StreamResponse(ctx context.Context, messages []message.Message, tools []tools.BaseTool) <-chan ProviderEvent {
	// Set current request info for display
	request.SetCurrent(string(x.providerOptions.model.Provider), x.providerOptions.model.APIModel, x.options.baseURL)
	
	openaiMessages := x.convertMessages(messages)
	openaiTools := x.convertTools(tools)
	
	params := x.preparedParams(openaiMessages, openaiTools)
	params.StreamOptions = openai.ChatCompletionStreamOptionsParam{
		IncludeUsage: openai.Bool(true),
	}
	
	return x.streamWithParams(ctx, params)
}

// streamWithParams implements custom retry logic for xAI
func (x *xai2Client) streamWithParams(ctx context.Context, params openai.ChatCompletionNewParams) <-chan ProviderEvent {
	eventChan := make(chan ProviderEvent)
	
	go func() {
		defer close(eventChan)
		attempts := 0
		
		for {
			attempts++
			
			// Debug logging
			if x.isDebugMode() {
				x.logDebug("Starting stream attempt %d", attempts)
			}
			
			// Create a new stream
			openaiStream := x.client.Chat.Completions.NewStreaming(ctx, params)
			
			acc := openai.ChatCompletionAccumulator{}
			currentContent := ""
			toolCalls := make([]message.ToolCall, 0)
			hasError := false
			var streamErr error
			
			// Process stream chunks
			for openaiStream.Next() {
				chunk := openaiStream.Current()
				acc.AddChunk(chunk)
				
				for _, choice := range chunk.Choices {
					if choice.Delta.Content != "" {
						// Check if content looks like an error
						if x.isErrorContent(choice.Delta.Content) {
							hasError = true
							streamErr = &xai2ContentError{
								StatusCode: 200, // Assume 200 since it came as content
								Content:    choice.Delta.Content,
							}
							break
						}
						
						eventChan <- ProviderEvent{
							Type:    EventContentDelta,
							Content: choice.Delta.Content,
						}
						currentContent += choice.Delta.Content
					}
				}
				
				if hasError {
					break
				}
			}
			
			// Check for stream error
			if streamErr == nil {
				streamErr = openaiStream.Err()
			}
			
			if streamErr != nil && !errors.Is(streamErr, io.EOF) {
				hasError = true
				
				// Debug logging
				if x.isDebugMode() {
					x.logDebug("Stream error on attempt %d: %v", attempts, streamErr)
				}
				
				// Check if we should retry
				retry, after, retryErr := x.shouldRetry(attempts, streamErr)
				
				if x.isDebugMode() {
					x.logDebug("shouldRetry returned: retry=%v, after=%d, retryErr=%v", retry, after, retryErr)
				}
				
				if retryErr != nil {
					request.Clear()
					eventChan <- ProviderEvent{Type: EventError, Error: retryErr}
					return
				}
				
				if retry {
					// Show retry message
					logging.WarnPersist(
						fmt.Sprintf("Retrying due to rate limit... attempt %d of %d", attempts, xai2MaxRetries),
						logging.PersistTimeArg, 
						time.Millisecond*time.Duration(after+100),
					)
					
					// Clear request info during retry
					request.Clear()
					
					// Wait before retrying
					select {
					case <-ctx.Done():
						eventChan <- ProviderEvent{Type: EventError, Error: ctx.Err()}
						return
					case <-time.After(time.Duration(after) * time.Millisecond):
						// Re-set request info for next attempt
						request.SetCurrent(string(x.providerOptions.model.Provider), x.providerOptions.model.APIModel, x.options.baseURL)
						continue // Try again
					}
				}
				
				// No retry - send error
				request.Clear()
				eventChan <- ProviderEvent{Type: EventError, Error: streamErr}
				return
			}
			
			// Success - stream completed
			if !hasError {
				// Check if the complete content is an error
				if x.isErrorContent(currentContent) {
					streamErr = &xai2ContentError{
						StatusCode: 200,
						Content:    currentContent,
					}
					eventChan <- ProviderEvent{Type: EventError, Error: streamErr}
					request.Clear()
					return
				}
				
				finishReason := x.finishReason(string(acc.ChatCompletion.Choices[0].FinishReason))
				if len(acc.ChatCompletion.Choices[0].Message.ToolCalls) > 0 {
					toolCalls = append(toolCalls, x.toolCalls(acc.ChatCompletion)...)
				}
				if len(toolCalls) > 0 {
					finishReason = message.FinishReasonToolUse
				}
				
				eventChan <- ProviderEvent{
					Type: EventComplete,
					Response: &ProviderResponse{
						Content:      currentContent,
						ToolCalls:    toolCalls,
						Usage:        x.usage(acc.ChatCompletion),
						FinishReason: finishReason,
					},
				}
				request.Clear()
				return
			}
		}
	}()
	
	return eventChan
}

// Override convertMessages to reload Grok prompt from file on each request
func (x *xai2Client) convertMessages(messages []message.Message) []openai.ChatCompletionMessageParamUnion {
	// Reload the system message from file for Grok models
	if externalPrompt := x.loadExternalGrokPrompt(); externalPrompt != "" {
		x.providerOptions.systemMessage = externalPrompt
	}
	
	// Prepend system message if it exists
	openaiMessages := []openai.ChatCompletionMessageParamUnion{}
	if x.providerOptions.systemMessage != "" {
		openaiMessages = append(openaiMessages, openai.SystemMessage(x.providerOptions.systemMessage))
	}
	
	// Convert the rest of the messages
	for _, msg := range messages {
		openaiMessages = append(openaiMessages, x.convertMessage(msg))
	}
	return openaiMessages
}

func (x *xai2Client) loadExternalGrokPrompt() string {
	// Search for the prompt file in the same locations as .opencode.json
	possiblePaths := []string{}
	
	if homeDir, err := os.UserHomeDir(); err == nil {
		// 1. $HOME/.opencode/grok4-system-prompt.md
		possiblePaths = append(possiblePaths, filepath.Join(homeDir, ".opencode", "grok4-system-prompt.md"))
		
		// 2. $XDG_CONFIG_HOME/opencode/grok4-system-prompt.md
		if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
			possiblePaths = append(possiblePaths, filepath.Join(xdgConfigHome, "opencode", "grok4-system-prompt.md"))
		}
		
		// 3. $HOME/.config/opencode/grok4-system-prompt.md
		possiblePaths = append(possiblePaths, filepath.Join(homeDir, ".config", "opencode", "grok4-system-prompt.md"))
	}
	
	// Try each path in order
	for _, path := range possiblePaths {
		if content, err := os.ReadFile(path); err == nil {
			return string(content)
		}
	}
	
	return ""
}

// Model returns the model configuration
func (x *xai2Client) Model() models.Model {
	return x.providerOptions.model
}

// shouldRetry handles xAI-specific retry logic
func (x *xai2Client) shouldRetry(attempts int, err error) (bool, int64, error) {
	// Check if we've exceeded max retries
	if attempts > xai2MaxRetries {
		return false, 0, fmt.Errorf("maximum retry attempts reached: %d retries", xai2MaxRetries)
	}
	
	// Check for xAI content errors
	var xaiErr *xai2ContentError
	if errors.As(err, &xaiErr) {
		// Use exponential backoff with jitter
		backoffMs := 2000 * (1 << (attempts - 1))
		jitterMs := int(float64(backoffMs) * 0.2)
		retryMs := backoffMs + jitterMs
		
		logging.WarnPersist(
			fmt.Sprintf("xAI content error detected: %s. Retrying... attempt %d of %d", 
				strings.TrimSpace(xaiErr.Content), attempts, xai2MaxRetries),
			logging.PersistTimeArg, 
			time.Millisecond*time.Duration(retryMs+100),
		)
		
		return true, int64(retryMs), nil
	}
	
	// Check for OpenAI API errors
	var apierr *openai.Error
	if errors.As(err, &apierr) {
		// Check if this is a quota/billing error from xAI
		if apierr.StatusCode == 429 && strings.Contains(err.Error(), quotaExceededPattern) {
			// This is a permanent error, don't retry
			return false, 0, fmt.Errorf("xAI quota exceeded: %s", err.Error())
		}
		
		// Handle standard rate limits
		if apierr.StatusCode == 429 || apierr.StatusCode == 500 {
			// Use exponential backoff
			backoffMs := 2000 * (1 << (attempts - 1))
			retryAfter := int64(backoffMs)
			
			// xAI doesn't provide retry-after headers in the error object
			
			return true, retryAfter, nil
		}
	}
	
	// Don't retry other errors
	return false, 0, err
}

// isErrorContent checks if the response content matches known error patterns
func (x *xai2Client) isErrorContent(content string) bool {
	if content == "" {
		return false
	}
	
	lowerContent := strings.ToLower(content)
	for _, pattern := range xai2ErrorPatterns {
		if strings.Contains(lowerContent, pattern) {
			return true
		}
	}
	return false
}

// Helper methods

func (x *xai2Client) convertMessage(msg message.Message) openai.ChatCompletionMessageParamUnion {
	content := msg.Content().Text
	toolCalls := msg.ToolCalls()
	
	switch msg.Role {
	case message.User:
		return openai.UserMessage(content)
	case message.Assistant:
		assistantMsg := openai.ChatCompletionAssistantMessageParam{
			Role: "assistant",
		}
		
		hasContent := false
		if content != "" {
			assistantMsg.Content = openai.ChatCompletionAssistantMessageParamContentUnion{
				OfString: openai.String(content),
			}
			hasContent = true
		}
		
		if len(toolCalls) > 0 {
			assistantMsg.ToolCalls = make([]openai.ChatCompletionMessageToolCallParam, len(toolCalls))
			for i, tc := range toolCalls {
				assistantMsg.ToolCalls[i] = openai.ChatCompletionMessageToolCallParam{
					ID:   tc.ID,
					Type: "function",
					Function: openai.ChatCompletionMessageToolCallFunctionParam{
						Name:      tc.Name,
						Arguments: tc.Input,
					},
				}
			}
			hasContent = true
		}
		
		// Only add the message if it has content or tool calls
		if hasContent {
			return openai.ChatCompletionMessageParamUnion{
				OfAssistant: &assistantMsg,
			}
		}
		return openai.AssistantMessage(content)
	case message.System:
		return openai.SystemMessage(content)
	case message.Tool:
		// xAI expects tool results to be sent as individual messages
		for _, result := range msg.ToolResults() {
			return openai.ToolMessage(result.Content, result.ToolCallID)
		}
		// Fallback if no tool results
		return openai.UserMessage(content)
	default:
		return openai.UserMessage(content)
	}
}

func (x *xai2Client) convertTools(tools []tools.BaseTool) []openai.ChatCompletionToolParam {
	openaiTools := make([]openai.ChatCompletionToolParam, len(tools))
	
	for i, tool := range tools {
		info := tool.Info()
		openaiTools[i] = openai.ChatCompletionToolParam{
			Function: openai.FunctionDefinitionParam{
				Name:        info.Name,
				Description: openai.String(info.Description),
				Parameters: openai.FunctionParameters{
					"type":       "object",
					"properties": info.Parameters,
					"required":   info.Required,
				},
			},
		}
	}
	
	return openaiTools
}

func (x *xai2Client) preparedParams(messages []openai.ChatCompletionMessageParamUnion, tools []openai.ChatCompletionToolParam) openai.ChatCompletionNewParams {
	params := openai.ChatCompletionNewParams{
		Model:    openai.ChatModel(x.providerOptions.model.APIModel),
		Messages: messages,
		Tools:    tools,
	}
	
	// For Grok 4 (reasoning model), use MaxCompletionTokens
	// For other models, use MaxTokens
	if x.providerOptions.model.CanReason {
		params.MaxCompletionTokens = openai.Int(x.providerOptions.maxTokens)
		// Grok 4 doesn't support reasoning_effort parameter
	} else {
		params.MaxTokens = openai.Int(x.providerOptions.maxTokens)
	}
	
	// xAI doesn't support:
	// - FrequencyPenalty
	// - PresencePenalty
	// - Stop sequences
	// - ResponseFormat (for Grok 4)
	
	return params
}

func (x *xai2Client) toProviderResponse(completion openai.ChatCompletion) *ProviderResponse {
	var content string
	var toolCalls []message.ToolCall
	
	if len(completion.Choices) > 0 {
		content = completion.Choices[0].Message.Content
		toolCalls = x.toolCalls(completion)
	}
	
	return &ProviderResponse{
		Content:      content,
		ToolCalls:    toolCalls,
		Usage:        x.usage(completion),
		FinishReason: x.finishReason(string(completion.Choices[0].FinishReason)),
	}
}

func (x *xai2Client) finishReason(reason string) message.FinishReason {
	switch reason {
	case "stop":
		return message.FinishReasonEndTurn
	case "length":
		return message.FinishReasonMaxTokens
	case "tool_calls", "function_call":
		return message.FinishReasonToolUse
	case "content_filter":
		return message.FinishReasonPermissionDenied
	default:
		return message.FinishReasonEndTurn
	}
}

func (x *xai2Client) toolCalls(completion openai.ChatCompletion) []message.ToolCall {
	if len(completion.Choices) == 0 || len(completion.Choices[0].Message.ToolCalls) == 0 {
		return nil
	}
	
	toolCalls := make([]message.ToolCall, len(completion.Choices[0].Message.ToolCalls))
	for i, tc := range completion.Choices[0].Message.ToolCalls {
		toolCalls[i] = message.ToolCall{
			ID:    tc.ID,
			Name:  tc.Function.Name,
			Input: tc.Function.Arguments,
			Type:  "function",
			Finished: true,
		}
	}
	return toolCalls
}

func (x *xai2Client) usage(completion openai.ChatCompletion) TokenUsage {
	return TokenUsage{
		InputTokens:  int64(completion.Usage.PromptTokens),
		OutputTokens: int64(completion.Usage.CompletionTokens),
		// xAI doesn't provide cache tokens info
		CacheCreationTokens: 0,
		CacheReadTokens:     0,
	}
}

// Logging methods

func (x *xai2Client) isDebugMode() bool {
	cfg := config.Get()
	return cfg != nil && cfg.Debug
}

func (x *xai2Client) logDebug(format string, args ...interface{}) {
	if !x.isDebugMode() {
		return
	}
	
	logFile, err := os.OpenFile("/tmp/xai-stream-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer logFile.Close()
	
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(logFile, "[%s] %s\n", timestamp, msg)
}

func (x *xai2Client) logRequest() {
	// Log request details
	
	dir := filepath.Join(os.TempDir(), "opencode")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return
	}
	
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("xai_%s_%s.json", x.providerOptions.model.ID, timestamp)
	path := filepath.Join(dir, filename)
	
	data := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"model":     x.providerOptions.model.ID,
		"provider":  "xai2",
		"baseURL":   x.options.baseURL,
	}
	
	jsonData, _ := json.MarshalIndent(data, "", "  ")
	os.WriteFile(path, jsonData, 0644)
}

func (x *xai2Client) logError(err error, params openai.ChatCompletionNewParams) {
	if !x.isDebugMode() {
		return
	}
	
	dir := filepath.Join(os.TempDir(), "opencode", "errors")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return
	}
	
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("xai_error_%s_%s.json", x.providerOptions.model.ID, timestamp)
	path := filepath.Join(dir, filename)
	
	data := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"model":     x.providerOptions.model.ID,
		"provider":  "xai2",
		"error":     err.Error(),
		"params":    params,
	}
	
	jsonData, _ := json.MarshalIndent(data, "", "  ")
	os.WriteFile(path, jsonData, 0644)
}