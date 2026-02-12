package imagegen

import (
	"context"
)

// Generator 单步「描述 → 图片」生成，返回本地路径或 URL
type Generator interface {
	Generate(ctx context.Context, prompt string) (pathOrURL string, err error)
}

// NoopGenerator 占位：不生成图，返回空字符串
type NoopGenerator struct{}

func (NoopGenerator) Generate(_ context.Context, _ string) (string, error) {
	return "", nil
}
