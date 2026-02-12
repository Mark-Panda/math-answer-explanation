package ocr

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gomath/gomath/internal/config"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// callVisionAPI 使用 langchaingo 调用 OpenAI 兼容的视觉模型（与 LLM 题目解析同一框架）
func callVisionAPI(ctx context.Context, cfg config.OCRConfig, imagePath string) (string, error) {
	data, mime, err := readImageAsBase64(imagePath)
	if err != nil {
		return "", err
	}
	key := cfg.APIKey()
	if key == "" {
		return "", fmt.Errorf("ocr api_key or api_key_env not set")
	}

	opts := []openai.Option{
		openai.WithToken(key),
		openai.WithModel(cfg.Model),
	}
	if cfg.APIBase != "" {
		opts = append(opts, openai.WithBaseURL(strings.TrimSuffix(cfg.APIBase, "/")))
	}
	llm, err := openai.New(opts...)
	if err != nil {
		return "", fmt.Errorf("openai client: %w", err)
	}

	dataURL := "data:" + mime + ";base64," + data
	content := llms.MessageContent{
		Role: llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{
			llms.TextPart(visionPrompt),
			llms.ImageURLWithDetailPart(dataURL, "low"),
		},
	}

	maxRetries := cfg.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 1
	}
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			default:
				time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
			}
		}
		out, err := llm.GenerateContent(ctx, []llms.MessageContent{content},
			llms.WithMaxTokens(2048))
		if err != nil {
			lastErr = err
			errStr := err.Error()
			// 5xx 或服务端临时错误可重试
			retryable := strings.Contains(errStr, "500") || strings.Contains(errStr, "502") ||
				strings.Contains(errStr, "503") || strings.Contains(errStr, "Gateway") ||
				strings.Contains(errStr, "unavailable") || strings.Contains(errStr, "internal error")
			if retryable && attempt+1 < maxRetries {
				continue
			}
			break
		}
		if len(out.Choices) == 0 {
			lastErr = fmt.Errorf("no choices in response")
			continue
		}
		text := strings.TrimSpace(out.Choices[0].Content)
		if text == "" {
			lastErr = fmt.Errorf("empty content")
			continue
		}
		return text, nil
	}
	return "", lastErr
}
