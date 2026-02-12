package explanation

// Step 解析步骤：标题、正文（Markdown+LaTeX）、配图描述
type Step struct {
	Title      string `json:"title"`
	Content    string `json:"content"`     // Markdown，公式用 $...$ / $$...$$
	ImagePrompt string `json:"image_prompt"` // 用于生成该步讲解图的描述
}

// Result 分步解析结果，与步骤一一对应的配图在生成后填入 ImageURL
type Result struct {
	Steps []StepResult `json:"steps"`
}

// StepResult 单步展示：文字 + 配图 URL（可选）
type StepResult struct {
	Title      string `json:"title"`
	Content    string `json:"content"`
	ImageURL   string `json:"image_url,omitempty"`   // 讲解图 URL，空表示暂无图
	ImagePrompt string `json:"image_prompt,omitempty"`
}
