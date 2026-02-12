package http

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/gomath/gomath/internal/explanation"
)

// ExplainGenerator 生成分步解析（支持文本或图片直接解析）
type ExplainGenerator interface {
	Generate(ctx context.Context, problemText string) (*explanation.Result, error)
	GenerateFromImage(ctx context.Context, imagePath string) (*explanation.Result, error)
}

// ExplainStore 存储解析结果
type ExplainStore interface {
	Put(r *explanation.Result) string
	Get(id string) (*explanation.Result, bool)
}

// ExplainRequest 请求生成解析：problem_text 与 image_path 二选一；传 image_path 时直接让模型看图解析
type ExplainRequest struct {
	ProblemText string `json:"problem_text"`
	ImagePath   string `json:"image_path"` // 已上传图片路径（相对 upload 目录），与 problem_text 二选一
}

// ExplainResponse 返回任务 ID，前端可轮询 GET /api/result/:id
type ExplainResponse struct {
	TaskID string `json:"task_id"`
}

// ResultResponse 解析结果（步骤列表 + 每步文字与配图 URL）
type ResultResponse struct {
	Steps []StepResponse `json:"steps"`
}

// StepResponse 单步
type StepResponse struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	ImageURL string `json:"image_url,omitempty"`
}

func (s *Server) handleExplain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req ExplainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.ProblemText != "" && req.ImagePath != "" {
		http.Error(w, "provide either problem_text or image_path, not both", http.StatusBadRequest)
		return
	}
	if req.ProblemText == "" && req.ImagePath == "" {
		http.Error(w, "problem_text or image_path required", http.StatusBadRequest)
		return
	}
	if s.ExplainGen == nil || s.ExplainStore == nil {
		http.Error(w, "explanation not configured", http.StatusServiceUnavailable)
		return
	}
	var result *explanation.Result
	var err error
	if req.ImagePath != "" {
		absPath := filepath.Join(s.UploadDir, req.ImagePath)
		result, err = s.ExplainGen.GenerateFromImage(r.Context(), absPath)
	} else {
		result, err = s.ExplainGen.Generate(r.Context(), req.ProblemText)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// 若配置了讲解图生成，按步骤生成并绑定 URL
	if s.ImageGen != nil {
		for i := range result.Steps {
			prompt := result.Steps[i].ImagePrompt
			if prompt == "" {
				prompt = result.Steps[i].Title + ": " + result.Steps[i].Content
			}
			url, _ := s.ImageGen.Generate(r.Context(), prompt)
			result.Steps[i].ImageURL = url
		}
	}
	taskID := s.ExplainStore.Put(result)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ExplainResponse{TaskID: taskID})
}

func (s *Server) handleResult(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	if taskID == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}
	if s.ExplainStore == nil {
		http.Error(w, "not configured", http.StatusServiceUnavailable)
		return
	}
	result, ok := s.ExplainStore.Get(taskID)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	steps := make([]StepResponse, 0, len(result.Steps))
	for _, st := range result.Steps {
		steps = append(steps, StepResponse{
			Title:    st.Title,
			Content:  st.Content,
			ImageURL: st.ImageURL,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ResultResponse{Steps: steps})
}
