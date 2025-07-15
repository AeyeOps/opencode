package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/shared/constant"
	"github.com/opencode-ai/opencode/internal/config"
	"github.com/opencode-ai/opencode/internal/llm/models"
	"github.com/opencode-ai/opencode/internal/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAIClientSendMessages(t *testing.T) {
	// Setup fake OpenAI server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			http.NotFound(w, r)
			return
		}
		resp := openai.ChatCompletion{
			ID:      "cmpl-test",
			Created: time.Now().Unix(),
			Model:   "gpt-4o",
			Object:  constant.ChatCompletion("chat.completion"),
			Choices: []openai.ChatCompletionChoice{
				{
					FinishReason: "stop",
					Index:        0,
					Logprobs: openai.ChatCompletionChoiceLogprobs{
						Content: []openai.ChatCompletionTokenLogprob{},
						Refusal: []openai.ChatCompletionTokenLogprob{},
					},
					Message: openai.ChatCompletionMessage{
						Content: "hello from server",
						Refusal: "",
						Role:    constant.Assistant("assistant"),
					},
				},
			},
			Usage: openai.CompletionUsage{
				CompletionTokens: 1,
				PromptTokens:     1,
				TotalTokens:      2,
				PromptTokensDetails: openai.CompletionUsagePromptTokensDetails{
					CachedTokens: 0,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	_, err := config.Load(t.TempDir(), false)
	require.NoError(t, err)

	model := models.OpenAIModels[models.GPT4o]
	p, err := NewProvider(models.ProviderOpenAI,
		WithModel(model),
		WithOpenAIOptions(WithOpenAIBaseURL(ts.URL)),
	)
	require.NoError(t, err)

	msg := message.Message{
		Role:  message.User,
		Parts: []message.ContentPart{message.TextContent{Text: "hello"}},
	}

	res, err := p.SendMessages(context.Background(), []message.Message{msg}, nil)
	require.NoError(t, err)

	assert.Equal(t, "hello from server", res.Content)
	assert.Equal(t, int64(1), res.Usage.OutputTokens)
}

func TestOpenAIShouldRetry(t *testing.T) {
	c := &openaiClient{}

	t.Run("non-openai error", func(t *testing.T) {
		retry, after, err := c.shouldRetry(1, assert.AnError)
		assert.False(t, retry)
		assert.Equal(t, int64(0), after)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("non retryable status", func(t *testing.T) {
		e := &openai.Error{StatusCode: 400, Response: &http.Response{Header: http.Header{}}, Request: httptest.NewRequest(http.MethodPost, "/", nil)}
		retry, after, err := c.shouldRetry(1, e)
		assert.False(t, retry)
		assert.Equal(t, int64(0), after)
		assert.Equal(t, e, err)
	})

	t.Run("max retries exceeded", func(t *testing.T) {
		e := &openai.Error{StatusCode: 429, Response: &http.Response{Header: http.Header{}}, Request: httptest.NewRequest(http.MethodPost, "/", nil)}
		retry, after, err := c.shouldRetry(maxRetries+1, e)
		assert.False(t, retry)
		assert.Equal(t, int64(0), after)
		assert.Error(t, err)
	})

	t.Run("retry with header", func(t *testing.T) {
		h := http.Header{"Retry-After": []string{"2"}}
		e := &openai.Error{StatusCode: 429, Response: &http.Response{Header: h}, Request: httptest.NewRequest(http.MethodPost, "/", nil)}
		retry, after, err := c.shouldRetry(1, e)
		assert.True(t, retry)
		assert.Equal(t, int64(2000), after)
		assert.NoError(t, err)
	})
}
