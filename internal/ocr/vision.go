package ocr

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
)

// 识别用 prompt：要求输出题目文字，公式用 LaTeX
const visionPrompt = `请识别图片中的数学题目，将题目文字完整、准确地输出。
要求：数学公式必须用 LaTeX 表示，行内公式用 $...$，独立公式用 $$...$$。
只输出题目内容本身，不要添加解析或答案。`

// readImageAsBase64 读取图片并转为 base64，返回 base64 字符串与 MIME 类型
func readImageAsBase64(imagePath string) (base64Str string, mime string, err error) {
	b, err := os.ReadFile(imagePath)
	if err != nil {
		return "", "", err
	}
	ext := strings.ToLower(filepath.Ext(imagePath))
	switch ext {
	case ".jpg", ".jpeg":
		mime = "image/jpeg"
	case ".png":
		mime = "image/png"
	case ".webp":
		mime = "image/webp"
	default:
		mime = "image/png"
	}
	return base64.StdEncoding.EncodeToString(b), mime, nil
}
