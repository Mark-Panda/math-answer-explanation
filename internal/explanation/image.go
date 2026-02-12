package explanation

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
)

// readImageAsBase64 读取图片并转为 base64，返回 base64 与 MIME
func readImageAsBase64(imagePath string) (data string, mime string, err error) {
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
