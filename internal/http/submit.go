package http

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"
)

// OCRRecognizer 识图能力：图片路径 → 题目文本
type OCRRecognizer interface {
	Recognize(ctx context.Context, imagePath string) (string, error)
}

// SubmitRequest 提交题目：仅文本，或先传图后的图片路径（相对 upload 目录）
type SubmitRequest struct {
	Text      string `json:"text"`       // 直接题目文字
	ImagePath string `json:"image_path"` // 已上传图片路径（相对 upload 目录），与 text 二选一
}

// SubmitResponse 统一返回题目文本，供前端展示/编辑或发起解析
type SubmitResponse struct {
	ProblemText string `json:"problem_text"`
}

func (s *Server) handleSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req SubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.Text != "" && req.ImagePath != "" {
		http.Error(w, "provide either text or image_path, not both", http.StatusBadRequest)
		return
	}
	if req.Text == "" && req.ImagePath == "" {
		http.Error(w, "provide text or image_path", http.StatusBadRequest)
		return
	}

	if req.Text != "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SubmitResponse{ProblemText: req.Text})
		return
	}

	// 先传图再以识图结果作为文本
	if s.OCR == nil {
		http.Error(w, "ocr not configured", http.StatusServiceUnavailable)
		return
	}
	absPath := filepath.Join(s.UploadDir, req.ImagePath)
	text, err := s.OCR.Recognize(r.Context(), absPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SubmitResponse{ProblemText: text})
}
