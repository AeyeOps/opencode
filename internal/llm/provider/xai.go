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
	"github.com/opencode-ai/opencode/internal/llm/tools"
	"github.com/opencode-ai/opencode/internal/logging"
	"github.com/opencode-ai/opencode/internal/message"
	"github.com/opencode-ai/opencode/internal/request"
)

type xaiClient struct {
	*openaiClient
}

type XAIClient ProviderClient

// xaiContentError represents an error when xAI returns error messages as content
type xaiContentError struct {
	StatusCode int
	Content    string
}

func (e *xaiContentError) Error() string {
	return fmt.Sprintf("xAI returned error content: %s", e.Content)
}

// xaiErrorPatterns contains patterns that indicate rate limiting or errors in content
var xaiErrorPatterns = []string{
	"try again",
	"rate limit",
	"too many requests",
	"please retry",
	"service unavailable",
	"quota exceeded",
	"temporarily unavailable",
}

func newXAIClient(opts providerClientOptions) XAIClient {
	// Create base OpenAI client with xAI endpoint
	openaiOpts := openaiOptions{
		reasoningEffort: "medium",
		baseURL:         "https://api.x.ai/v1",
	}

	for _, o := range opts.openaiOptions {
		o(&openaiOpts)
	}

	// Build OpenAI client options
	clientOptions := []option.RequestOption{
		option.WithBaseURL("https://api.x.ai/v1"),
	}
	if opts.apiKey != "" {
		clientOptions = append(clientOptions, option.WithAPIKey(opts.apiKey))
	}

	base := &openaiClient{
		providerOptions: opts,
		options:         openaiOpts,
		client:          openai.NewClient(clientOptions...),
	}

	return &xaiClient{openaiClient: base}
}

// Override convertMessages to reload Grok prompt from file on each request
func (x *xaiClient) convertMessages(messages []message.Message) []openai.ChatCompletionMessageParamUnion {
	// Reload the system message from file for Grok models
	if externalPrompt := x.loadExternalGrokPrompt(); externalPrompt != "" {
		x.providerOptions.systemMessage = externalPrompt
	}

	// Call the parent implementation
	return x.openaiClient.convertMessages(messages)
}

func (x *xaiClient) loadExternalGrokPrompt() string {
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

// isErrorContent checks if the response content matches known error patterns
func (x *xaiClient) isErrorContent(content string) bool {
	if content == "" {
		return false
	}

	lowerContent := strings.ToLower(content)
	for _, pattern := range xaiErrorPatterns {
		if strings.Contains(lowerContent, pattern) {
			return true
		}
	}
	return false
}

// Override preparedParams to filter out unsupported parameters for xAI
func (x *xaiClient) preparedParams(messages []openai.ChatCompletionMessageParamUnion, tools []openai.ChatCompletionToolParam) openai.ChatCompletionNewParams {
	params := openai.ChatCompletionNewParams{
		Model:    openai.ChatModel(x.providerOptions.model.APIModel),
		Messages: messages,
		Tools:    tools,
	}

	// Log request info if debug logging is enabled
	if cfg := x.getConfig(); cfg != nil && cfg.Debug {
		x.logRequest("xAI", x.providerOptions.model.APIModel, len(tools) > 0)
	}

	// xAI doesn't support reasoning_effort parameter, but Grok-4 is always in reasoning mode
	// Use MaxTokens instead of MaxCompletionTokens for xAI
	params.MaxTokens = openai.Int(x.providerOptions.maxTokens)

	// xAI API doesn't support these parameters - they remain unset (nil)
	// - PresencePenalty
	// - FrequencyPenalty
	// - Stop
	// - ReasoningEffort

	return params
}

// Override send to use our custom preparedParams
func (x *xaiClient) send(ctx context.Context, messages []message.Message, tools []tools.BaseTool) (*ProviderResponse, error) {
	// Set current request info for display
	request.SetCurrent(string(x.providerOptions.model.Provider), x.providerOptions.model.APIModel, "https://api.x.ai/v1")

	// Convert messages and tools
	openaiMessages := x.convertMessages(messages)
	openaiTools := x.convertTools(tools)

	// Use our custom preparedParams that filters out unsupported parameters
	params := x.preparedParams(openaiMessages, openaiTools)

	// Use our custom sendWithParams that detects content-based errors
	response, err := x.sendWithParams(ctx, params)
	if err != nil {
		x.logError(err, params)
		request.Clear() // Clear request info on error
	}
	if err == nil {
		request.Clear() // Clear request info on successful completion
	}
	return response, err
}

// Override sendWithParams to detect content-based errors
func (x *xaiClient) sendWithParams(ctx context.Context, params openai.ChatCompletionNewParams) (*ProviderResponse, error) {
	// Call the parent implementation
	response, err := x.openaiClient.sendWithParams(ctx, params)

	// Check for content-based errors even on successful responses
	if err == nil && response != nil && x.isErrorContent(response.Content) {
		// Log the content error
		contentErr := &xaiContentError{
			StatusCode: 429, // Treat as rate limit
			Content:    response.Content,
		}
		x.logError(contentErr, params)

		// Convert to error that can trigger retry
		return nil, contentErr
	}

	return response, err
}

// Override stream to use our custom preparedParams
func (x *xaiClient) stream(ctx context.Context, messages []message.Message, tools []tools.BaseTool) <-chan ProviderEvent {
	// Set current request info for display
	request.SetCurrent(string(x.providerOptions.model.Provider), x.providerOptions.model.APIModel, "https://api.x.ai/v1")

	// Debug logging
	logFile, _ := os.OpenFile("/tmp/xai-stream-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if logFile != nil {
		fmt.Fprintf(logFile, "\n[%s] ========== STREAM REQUEST STARTED ==========\n", time.Now().Format("2006-01-02 15:04:05.000"))
		fmt.Fprintf(logFile, "[%s] Model: %s\n", time.Now().Format("2006-01-02 15:04:05.000"), x.providerOptions.model.APIModel)
		fmt.Fprintf(logFile, "[%s] Number of messages: %d\n", time.Now().Format("2006-01-02 15:04:05.000"), len(messages))
		logFile.Close()
	}

	// Convert messages and tools
	openaiMessages := x.convertMessages(messages)
	openaiTools := x.convertTools(tools)

	// Use our custom preparedParams that filters out unsupported parameters
	params := x.preparedParams(openaiMessages, openaiTools)
	params.StreamOptions = openai.ChatCompletionStreamOptionsParam{
		IncludeUsage: openai.Bool(true),
	}

	// Use our custom streamWithParams that includes retry logic
	// This will handle retries for both HTTP errors and content errors
	return x.streamWithParams(ctx, params)
}

// wrapStreamEvents monitors streaming responses for error patterns
func (x *xaiClient) wrapStreamEvents(baseEvents <-chan ProviderEvent) <-chan ProviderEvent {
	wrappedEvents := make(chan ProviderEvent)

	go func() {
		defer close(wrappedEvents)
		var accumulatedContent strings.Builder
		var bufferedEvents []ProviderEvent
		var isBuffering bool

		// Raw logging for debugging
		logFile, _ := os.OpenFile("/tmp/xai-stream-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if logFile != nil {
			defer logFile.Close()
			fmt.Fprintf(logFile, "\n[%s] ========== NEW STREAM ==========\n", time.Now().Format("2006-01-02 15:04:05.000"))
			fmt.Fprintf(logFile, "[%s] wrapStreamEvents: Started processing baseEvents\n", time.Now().Format("2006-01-02 15:04:05.000"))
		}

		eventCount := 0
		for event := range baseEvents {
			eventCount++
			// Log every event
			if logFile != nil {
				fmt.Fprintf(logFile, "[%s] Event Type: %s\n", time.Now().Format("2006-01-02 15:04:05.000"), event.Type)
				if event.Type == EventContentDelta {
					fmt.Fprintf(logFile, "  Content Delta: %q\n", event.Content)
				} else if event.Type == EventComplete && event.Response != nil {
					fmt.Fprintf(logFile, "  Complete Content: %q\n", event.Response.Content)
					fmt.Fprintf(logFile, "  Finish Reason: %v\n", event.Response.FinishReason)
				} else if event.Type == EventError {
					fmt.Fprintf(logFile, "  Error: %v\n", event.Error)
				}
			}

			switch event.Type {
			case EventContentDelta:
				// Always buffer content deltas initially
				accumulatedContent.WriteString(event.Content)
				bufferedEvents = append(bufferedEvents, event)
				isBuffering = true

				// Don't forward any deltas until we know it's not an error

			case EventComplete:
				fullContent := accumulatedContent.String()
				if event.Response != nil {
					fullContent = event.Response.Content
				}

				if logFile != nil {
					fmt.Fprintf(logFile, "[%s] Accumulated Content: %q\n", time.Now().Format("2006-01-02 15:04:05.000"), fullContent)
					fmt.Fprintf(logFile, "[%s] Is Error Content: %v\n", time.Now().Format("2006-01-02 15:04:05.000"), x.isErrorContent(fullContent))
				}

				if x.isErrorContent(fullContent) {
					// Don't forward any buffered events - convert to error
					wrappedEvents <- ProviderEvent{
						Type: EventError,
						Error: &xaiContentError{
							StatusCode: 429,
							Content:    fullContent,
						},
					}
				} else if isBuffering {
					// Forward all buffered content deltas
					for _, bufferedEvent := range bufferedEvents {
						wrappedEvents <- bufferedEvent
					}
					// Forward the complete event
					wrappedEvents <- event
				} else {
					// No buffered events, just forward complete
					wrappedEvents <- event
				}

				// Reset for next message
				accumulatedContent.Reset()
				bufferedEvents = nil
				isBuffering = false

			case EventError:
				// Clear any buffered events on error
				accumulatedContent.Reset()
				bufferedEvents = nil
				isBuffering = false
				wrappedEvents <- event

			default:
				// Forward other events immediately (tool calls, etc)
				wrappedEvents <- event
			}
		}

		// Log stream completion
		if logFile != nil {
			fmt.Fprintf(logFile, "[%s] wrapStreamEvents: Finished processing %d events from baseEvents\n",
				time.Now().Format("2006-01-02 15:04:05.000"), eventCount)
		}

		// If stream ended with buffered events (shouldn't happen normally)
		if len(bufferedEvents) > 0 && !x.isErrorContent(accumulatedContent.String()) {
			for _, bufferedEvent := range bufferedEvents {
				wrappedEvents <- bufferedEvent
			}
		}
	}()

	return wrappedEvents
}

// shouldRetry handles xAI-specific errors while maintaining OpenAI compatibility
func (x *xaiClient) shouldRetry(attempts int, err error) (bool, int64, error) {
	// Check for xAI content errors
	var xaiErr *xaiContentError
	if errors.As(err, &xaiErr) {
		if attempts > maxRetries {
			return false, 0, fmt.Errorf("maximum retry attempts reached for xAI rate limit: %d retries", maxRetries)
		}

		// Use exponential backoff with jitter
		backoffMs := 2000 * (1 << (attempts - 1))
		jitterMs := int(float64(backoffMs) * 0.2)
		retryMs := backoffMs + jitterMs

		// Log the retry attempt
		logging.WarnPersist(
			fmt.Sprintf("xAI content error detected: %s. Retrying... attempt %d of %d",
				strings.TrimSpace(xaiErr.Content), attempts, maxRetries),
			logging.PersistTimeArg,
			time.Millisecond*time.Duration(retryMs+100),
		)

		return true, int64(retryMs), nil
	}

	// Check for OpenAI API errors
	var apierr *openai.Error
	if errors.As(err, &apierr) {
		// Check if this is a quota/billing error from xAI
		if apierr.StatusCode == 429 && strings.Contains(err.Error(), "credits or reached its monthly spending limit") {
			// This is a permanent error, don't retry
			return false, 0, fmt.Errorf("xAI quota exceeded: %s", err.Error())
		}
	}

	// Fall back to OpenAI's shouldRetry for standard errors
	return x.openaiClient.shouldRetry(attempts, err)
}

// Override streamWithParams to use our custom shouldRetry and error detection
// We completely replace the parent implementation to control retry logic
func (x *xaiClient) streamWithParams(ctx context.Context, params openai.ChatCompletionNewParams) <-chan ProviderEvent {
	eventChan := make(chan ProviderEvent)

	go func() {
		defer close(eventChan)
		attempts := 0

		for {
			attempts++

			// Create a new stream directly
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
						eventChan <- ProviderEvent{
							Type:    EventContentDelta,
							Content: choice.Delta.Content,
						}
						currentContent += choice.Delta.Content
					}
				}
			}

			// Check for error
			streamErr = openaiStream.Err()
			if streamErr != nil && !errors.Is(streamErr, io.EOF) {
				hasError = true

				// Log the error
				logFile, _ := os.OpenFile("/tmp/xai-stream-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
				if logFile != nil {
					fmt.Fprintf(logFile, "[%s] xAI streamWithParams: Error on attempt %d: %v\n",
						time.Now().Format("2006-01-02 15:04:05.000"), attempts, streamErr)
					logFile.Close()
				}

				// Check if we should retry using our custom logic
				retry, after, retryErr := x.shouldRetry(attempts, streamErr)

				// Log retry decision
				logFile, _ = os.OpenFile("/tmp/xai-stream-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
				if logFile != nil {
					fmt.Fprintf(logFile, "[%s] xAI streamWithParams: shouldRetry returned: retry=%v, after=%d, retryErr=%v\n",
						time.Now().Format("2006-01-02 15:04:05.000"), retry, after, retryErr)
					logFile.Close()
				}

				if retryErr != nil {
					request.Clear()
					eventChan <- ProviderEvent{Type: EventError, Error: retryErr}
					return
				}

				if retry {
					// Show retry message
					logging.WarnPersist(
						fmt.Sprintf("Retrying due to rate limit... attempt %d of %d", attempts, maxRetries),
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
						request.SetCurrent(string(x.providerOptions.model.Provider), x.providerOptions.model.APIModel, "https://api.x.ai/v1")
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

// Helper methods from openaiClient that we need
func (x *xaiClient) finishReason(reason string) message.FinishReason {
	return x.openaiClient.finishReason(reason)
}

func (x *xaiClient) toolCalls(completion openai.ChatCompletion) []message.ToolCall {
	return x.openaiClient.toolCalls(completion)
}

func (x *xaiClient) usage(completion openai.ChatCompletion) TokenUsage {
	return x.openaiClient.usage(completion)
}

// logError logs xAI-specific errors with detailed information
func (x *xaiClient) logError(err error, params openai.ChatCompletionNewParams) {
	// Only log errors in debug mode
	if cfg := x.getConfig(); cfg != nil && cfg.Debug {
		logPath := x.getLogPath()
		if logPath != "" {
			debugFile, _ := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if debugFile != nil {
				defer debugFile.Close()

				fmt.Fprintf(debugFile, "[%s] xAI Error: %v\n", time.Now().Format("15:04:05"), err)

				// Log xAI-specific content errors
				var xaiErr *xaiContentError
				if errors.As(err, &xaiErr) {
					fmt.Fprintf(debugFile, "[%s] xAI Content Error Details:\n", time.Now().Format("15:04:05"))
					fmt.Fprintf(debugFile, "  - Error Type: Content-based rate limit\n")
					fmt.Fprintf(debugFile, "  - Status Code: %d\n", xaiErr.StatusCode)
					fmt.Fprintf(debugFile, "  - Content: %s\n", strings.TrimSpace(xaiErr.Content))
					fmt.Fprintf(debugFile, "  - Model: %s\n", params.Model)
				}

				// Log the request parameters for debugging
				jsonParams, _ := json.Marshal(params)
				fmt.Fprintf(debugFile, "[%s] xAI Request Params: %s\n", time.Now().Format("15:04:05"), string(jsonParams))

				// Log the full response body if available for standard errors
				var apierr *openai.Error
				if errors.As(err, &apierr) && apierr.Response != nil {
					responseBody := apierr.DumpResponse(true)
					fmt.Fprintf(debugFile, "[%s] xAI Response Body: %s\n", time.Now().Format("15:04:05"), string(responseBody))
				}
			}
		}
	}
}

// getConfig returns the current config
func (x *xaiClient) getConfig() *config.Config {
	return config.Get()
}

// getLogPath returns the path to the debug log file
func (x *xaiClient) getLogPath() string {
	cfg := x.getConfig()
	if cfg != nil && cfg.Data.Directory != "" {
		return filepath.Join(cfg.Data.Directory, "requests.log")
	}
	return ""
}

// logRequest logs request information to the debug log
func (x *xaiClient) logRequest(provider, model string, hasTools bool) {
	logPath := x.getLogPath()
	if logPath != "" {
		debugFile, _ := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if debugFile != nil {
			defer debugFile.Close()
			fmt.Fprintf(debugFile, "[%s] Provider: %s, Model: %s, URL: https://api.x.ai/v1, HasTools: %v\n",
				time.Now().Format("15:04:05"),
				provider,
				model,
				hasTools)
		}
	}
}
