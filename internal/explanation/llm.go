package explanation

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/gomath/gomath/internal/config"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// Generator 使用 LLM 题目解析配置块生成分步解析
type Generator struct {
	cfg config.LLMExplanationConfig
}

// NewGenerator 根据统一配置中的 llm.explanation 创建
func NewGenerator(cfg config.LLMExplanationConfig) *Generator {
	return &Generator{cfg: cfg}
}

// GenerateFromImage 基于题目图片直接生成分步解析（多模态：图片 + 提示），返回步骤序列
func (g *Generator) GenerateFromImage(ctx context.Context, imagePath string) (*Result, error) {
	if g.cfg.Provider == "" || g.cfg.Model == "" {
		return nil, fmt.Errorf("llm explanation not configured")
	}
	if g.cfg.Provider != "openai" {
		return generateStub("")
	}
	ctx, cancel := context.WithTimeout(ctx, g.cfg.Timeout())
	defer cancel()
	data, mime, err := readImageAsBase64(imagePath)
	if err != nil {
		return nil, fmt.Errorf("read image: %w", err)
	}
	opts := []openai.Option{
		openai.WithToken(g.cfg.APIKey()),
		openai.WithModel(g.cfg.Model),
	}
	if g.cfg.APIBase != "" {
		opts = append(opts, openai.WithBaseURL(strings.TrimSuffix(g.cfg.APIBase, "/")))
	}
	llm, err := openai.New(opts...)
	if err != nil {
		return nil, err
	}
	temperature := g.cfg.Temperature
	if temperature <= 0 {
		temperature = 0.3
	}
	maxTokens := g.cfg.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
	}
	prompt := buildPromptFromImage()
	dataURL := "data:" + mime + ";base64," + data
	content := llms.MessageContent{
		Role: llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{
			llms.TextPart(prompt),
			llms.ImageURLWithDetailPart(dataURL, "low"),
		},
	}
	out, err := llm.GenerateContent(ctx, []llms.MessageContent{content},
		llms.WithTemperature(temperature), llms.WithMaxTokens(maxTokens))
	if err != nil {
		return nil, err
	}
	if len(out.Choices) == 0 {
		return nil, fmt.Errorf("no response from llm")
	}
	return parseStepsResponse(out.Choices[0].Content)
}

func buildPromptFromImage() string {
	return `你是一个数学题解析助手。请根据图片中的数学题目，直接给出分步解析，并严格按以下 JSON 数组格式输出（不要其他前后文字），每步包含 title、content、image_prompt：
- title: 该步简短标题
- content: 该步详细解析，数学公式用 LaTeX，行内用 $...$，块级用 $$...$$
- image_prompt: 用于生成该步讲解图的英文描述（示意图、几何、函数图等）

直接输出 JSON 数组，例如：
[{"title":"步骤1","content":"...","image_prompt":"..."},{"title":"步骤2",...}]
`
}

// Generate 基于题目文本生成分步解析，返回步骤序列（含 title、content、image_prompt）
func (g *Generator) Generate(ctx context.Context, problemText string) (*Result, error) {
	if g.cfg.Provider == "" || g.cfg.Model == "" {
		return nil, fmt.Errorf("llm explanation not configured")
	}
	// 使用 langchaingo 调用大模型；此处以 OpenAI 为例，其他 provider 可扩展
	if g.cfg.Provider != "openai" {
		return generateStub(problemText)
	}
	ctx, cancel := context.WithTimeout(ctx, g.cfg.Timeout())
	defer cancel()
	opts := []openai.Option{
		openai.WithToken(g.cfg.APIKey()),
		openai.WithModel(g.cfg.Model),
	}
	if g.cfg.APIBase != "" {
		opts = append(opts, openai.WithBaseURL(g.cfg.APIBase))
	}
	llm, err := openai.New(opts...)
	if err != nil {
		return nil, err
	}
	temperature := g.cfg.Temperature
	if temperature <= 0 {
		temperature = 0.3
	}
	maxTokens := g.cfg.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
	}
	prompt := buildPrompt(problemText)
	out, err := llm.GenerateContent(ctx, []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	}, llms.WithTemperature(temperature), llms.WithMaxTokens(maxTokens))
	if err != nil {
		return nil, err
	}
	if len(out.Choices) == 0 {
		return nil, fmt.Errorf("no response from llm")
	}
	text := out.Choices[0].Content
	return parseStepsResponse(text)
}

func buildPrompt(problemText string) string {
	return `你是一个数学题解析助手。请对以下题目给出分步解析，并严格按以下 JSON 数组格式输出（不要其他前后文字），每步包含 title、content、image_prompt：
- title: 该步简短标题
- content: 该步详细解析，数学公式用 LaTeX，行内用 $...$，块级用 $$...$$
- image_prompt: 用于生成该步讲解图的英文描述（示意图、几何、函数图等）

题目：
` + problemText + `

直接输出 JSON 数组，例如：
[{"title":"步骤1","content":"...","image_prompt":"..."},{"title":"步骤2",...}]
`
}

func parseStepsResponse(text string) (*Result, error) {
	text = strings.TrimSpace(text)
	log.Printf("[explanation] llm raw output (len=%d): %s", len(text), text)
	if text == "" {
		return nil, fmt.Errorf("parse llm steps: empty response from model")
	}
	// 去除可能的 markdown 代码块（```json ... ``` 或 ``` ... ```）
	if strings.HasPrefix(text, "```") {
		text = strings.TrimPrefix(text, "```json")
		text = strings.TrimPrefix(text, "```")
		text = strings.TrimSuffix(text, "```")
		text = strings.TrimSpace(text)
	}
	// 若仍有前后说明文字，尝试只取第一个 '[' 到最后一个 ']' 的 JSON 数组
	if idx := strings.Index(text, "["); idx >= 0 {
		if last := strings.LastIndex(text, "]"); last > idx {
			text = text[idx : last+1]
		}
	}
	var steps []Step
	if err := json.Unmarshal([]byte(text), &steps); err != nil {
		return nil, fmt.Errorf("parse llm steps: %w (response length %d)", err, len(text))
	}
	res := &Result{Steps: make([]StepResult, 0, len(steps))}
	for _, s := range steps {
		res.Steps = append(res.Steps, StepResult{
			Title:       s.Title,
			Content:     s.Content,
			ImagePrompt: s.ImagePrompt,
		})
	}
	return res, nil
}

// generateStub 未配置 openai 时的占位
func generateStub(problemText string) (*Result, error) {
	_ = problemText
	return &Result{
		Steps: []StepResult{
			{Title: "步骤1", Content: "设 $x^2 - 5x + 6 = (x-a)(x-b)$，则 $a+b=5$，$ab=6$。", ImagePrompt: "quadratic equation factored form"},
			{Title: "步骤2", Content: "解得 $a=2,b=3$ 或 $a=3,b=2$，故 $x=2$ 或 $x=3$。", ImagePrompt: "number line with roots"},
		},
	}, nil
}
