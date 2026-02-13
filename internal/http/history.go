package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gomath/gomath/internal/history"
)

// HistoryStore 历史存储接口
type HistoryStore interface {
	List() []history.Item
	Add(it history.Item) string
	UpdateResult(id string, result *history.Result, taskID string) bool
	FindLatestUploadByPath(path string) *history.Item
}

// HistoryListResponse 列表响应
type HistoryListResponse struct {
	Items []history.Item `json:"items"`
}

// HistoryCreateRequest 创建请求
type HistoryCreateRequest struct {
	Type string `json:"type"` // "upload" | "text"
	Path string `json:"path,omitempty"`
	Text string `json:"text,omitempty"`
	At   int64  `json:"at"`
}

// HistoryCreateResponse 创建响应
type HistoryCreateResponse struct {
	ID string `json:"id"`
}

// HistoryUpdateResultRequest 更新结果请求
type HistoryUpdateResultRequest struct {
	Result *history.Result `json:"result"`
	TaskID string         `json:"task_id"`
}

func (s *Server) handleHistoryList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.HistoryStore == nil {
		http.Error(w, "history not configured", http.StatusServiceUnavailable)
		return
	}
	items := s.HistoryStore.List()
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(HistoryListResponse{Items: items})
}

func (s *Server) handleHistoryCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.HistoryStore == nil {
		http.Error(w, "history not configured", http.StatusServiceUnavailable)
		return
	}
	var req HistoryCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.Type != "upload" && req.Type != "text" {
		http.Error(w, "type must be upload or text", http.StatusBadRequest)
		return
	}
	it := history.Item{Type: req.Type, Path: req.Path, Text: req.Text, At: req.At}
	id := s.HistoryStore.Add(it)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(HistoryCreateResponse{ID: id})
}

func (s *Server) handleHistoryUpdateResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch && r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.HistoryStore == nil {
		http.Error(w, "history not configured", http.StatusServiceUnavailable)
		return
	}
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}
	var req HistoryUpdateResultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.Result == nil {
		http.Error(w, "result required", http.StatusBadRequest)
		return
	}
	ok := s.HistoryStore.UpdateResult(id, req.Result, req.TaskID)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleHistoryFindLatestUpload 供前端“当前上传”解析时拿到要更新的 history id（可选，前端也可用创建时返回的 id）
func (s *Server) handleHistoryFindLatestUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.HistoryStore == nil {
		http.Error(w, "history not configured", http.StatusServiceUnavailable)
		return
	}
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "path required", http.StatusBadRequest)
		return
	}
	it := s.HistoryStore.FindLatestUploadByPath(path)
	if it == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": it.ID})
}
