package config

import (
	"os"
	"time"
)

// Models 从统一配置文件 config/models.yaml 加载的完整配置。
// 内含 ocr、llm、video 三块，分别供识图、解析、视频模块使用。
type Models struct {
	OCR   OCRConfig    `yaml:"ocr"`
	LLM   LLMConfig    `yaml:"llm"`
	Video VideoConfig  `yaml:"video"`
}

// OCRConfig OCR 识图配置：图片 → 题目文本
type OCRConfig struct {
	Provider     string `yaml:"provider"`
	Model        string `yaml:"model"`
	APIBase      string `yaml:"api_base"`
	APIKeyValue  string `yaml:"api_key"`      // 优先使用：直接从配置文件读取
	APIKeyEnv    string `yaml:"api_key_env"`   // 可选：api_key 为空时从该环境变量读取
	TimeoutSec   int    `yaml:"timeout_sec"`
	MaxRetries   int    `yaml:"max_retries"`
}

// APIKey 返回 OCR 使用的 API Key：优先使用配置文件中的 api_key，否则从 api_key_env 环境变量读取。
func (c OCRConfig) APIKey() string {
	if c.APIKeyValue != "" {
		return c.APIKeyValue
	}
	if c.APIKeyEnv != "" {
		return os.Getenv(c.APIKeyEnv)
	}
	return ""
}

// Timeout 返回超时时间
func (c OCRConfig) Timeout() time.Duration {
	if c.TimeoutSec <= 0 {
		return 30 * time.Second
	}
	return time.Duration(c.TimeoutSec) * time.Second
}

// LLMConfig LLM 配置块，内含题目解析等用途
type LLMConfig struct {
	Explanation LLMExplanationConfig `yaml:"explanation"`
}

// LLMExplanationConfig 题目解析 LLM：题目文本 → 分步解析
type LLMExplanationConfig struct {
	Provider          string  `yaml:"provider"`
	Model             string  `yaml:"model"`
	APIBase           string  `yaml:"api_base"`
	APIKeyValue       string  `yaml:"api_key"`      // 优先使用：直接从配置文件读取
	APIKeyEnv         string  `yaml:"api_key_env"`   // 可选：api_key 为空时从该环境变量读取
	Temperature       float64 `yaml:"temperature"`
	MaxTokens         int     `yaml:"max_tokens"`
	TimeoutSec        int     `yaml:"timeout_sec"`   // 单次解析请求超时（秒），≤0 时默认 180
	SystemPromptFile  string  `yaml:"system_prompt_file"`
}

// APIKey 返回 LLM 使用的 API Key：优先使用配置文件中的 api_key，否则从 api_key_env 环境变量读取。
func (c LLMExplanationConfig) APIKey() string {
	if c.APIKeyValue != "" {
		return c.APIKeyValue
	}
	if c.APIKeyEnv != "" {
		return os.Getenv(c.APIKeyEnv)
	}
	return ""
}

// Timeout 返回解析请求超时时间；≤0 时默认 180 秒（多模态/长文本可能较慢）。
func (c LLMExplanationConfig) Timeout() time.Duration {
	if c.TimeoutSec <= 0 {
		return 180 * time.Second
	}
	return time.Duration(c.TimeoutSec) * time.Second
}

// VideoConfig 视频生成配置（Phase 2 后续待办）
type VideoConfig struct {
	TTSProvider             string `yaml:"tts_provider"`
	TTSModel                string `yaml:"tts_model"`
	FFmpegBin               string `yaml:"ffmpeg_bin"`
	DefaultStepDurationSec   int    `yaml:"default_step_duration_sec"`
}
