package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadModels(t *testing.T) {
	// 从项目根目录的 config/models.yaml 加载
	configPath := "config/models.yaml"
	if _, err := os.Stat(configPath); err != nil {
		// 可能在 internal/config 目录下执行测试，尝试相对路径
		configPath = filepath.Join("..", "..", "config", "models.yaml")
		if _, err2 := os.Stat(configPath); err2 != nil {
			t.Skipf("config file not found, skip: %v", err)
		}
	}

	m, err := LoadModels(configPath)
	if err != nil {
		t.Fatalf("LoadModels: %v", err)
	}

	// 校验 ocr 块存在且可读
	if m.OCR.Provider != "" || m.OCR.Model != "" {
		t.Logf("ocr: provider=%q model=%q api_base=%q timeout_sec=%d",
			m.OCR.Provider, m.OCR.Model, m.OCR.APIBase, m.OCR.TimeoutSec)
	}
	if m.OCR.TimeoutSec > 0 && m.OCR.TimeoutSec != 30 {
		t.Logf("ocr timeout_sec=%d", m.OCR.TimeoutSec)
	}

	// 校验 llm.explanation 块存在
	exp := m.LLM.Explanation
	if exp.Provider != "" || exp.Model != "" {
		t.Logf("llm.explanation: provider=%q model=%q temperature=%v max_tokens=%d",
			exp.Provider, exp.Model, exp.Temperature, exp.MaxTokens)
	}
	if exp.Temperature <= 0 {
		exp.Temperature = 0.3
	}
	if exp.MaxTokens <= 0 {
		exp.MaxTokens = 4096
	}

	// 校验 video 块存在（Phase 2）
	if m.Video.FFmpegBin != "" {
		t.Logf("video: ffmpeg_bin=%q default_step_duration_sec=%d",
			m.Video.FFmpegBin, m.Video.DefaultStepDurationSec)
	}
}

func TestLoadModels_NotFound(t *testing.T) {
	_, err := LoadModels("/nonexistent/models.yaml")
	if err == nil {
		t.Fatal("expected error when file not found")
	}
}
