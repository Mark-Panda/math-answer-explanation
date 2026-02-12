package ocr

import (
	"context"
	"fmt"
	"os"

	"github.com/gomath/gomath/internal/config"
)

// Service 识图服务：图片 → 题目文本（含 LaTeX 公式），使用 OCR 配置块
type Service struct {
	cfg config.OCRConfig
}

// NewService 根据统一配置中的 ocr 块创建识图服务
func NewService(cfg config.OCRConfig) *Service {
	return &Service{cfg: cfg}
}

// Recognize 将图片转为题目文本。imagePath 为已上传文件的路径（绝对或相对 upload 目录）。
// 返回的文本中公式应以 LaTeX 表示（如 $...$ / $$...$$），供后续解析与前端渲染。
// 未配置 provider/model 时使用占位结果，便于联调；配置后调用真实多模态/视觉 API。
func (s *Service) Recognize(ctx context.Context, imagePath string) (string, error) {
	if _, err := os.Stat(imagePath); err != nil {
		return "", fmt.Errorf("image file: %w", err)
	}
	if s.cfg.Provider != "" && s.cfg.Model != "" {
		// OpenAI 或兼容 OpenAI 的视觉 API（如 ops-ai-gateway、火山等）
		text, err := callVisionAPI(ctx, s.cfg, imagePath)
		if err != nil {
			return "", fmt.Errorf("vision api: %w", err)
		}
		return text, nil
	}
	return recognizeStub(imagePath)
}

// recognizeStub 占位实现：未接入真实 API 时返回示例文本，便于联调
func recognizeStub(_ string) (string, error) {
	return "示例题目：求一元二次方程 $x^2 - 5x + 6 = 0$ 的解。", nil
}
