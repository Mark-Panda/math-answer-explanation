package http

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// 允许的图片类型
var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

// UploadResponse 上传成功响应
type UploadResponse struct {
	Path string `json:"path"` // 相对路径，用于后续识图等
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	maxBytes := int64(s.MaxSizeMB) * 1024 * 1024
	if err := r.ParseMultipartForm(maxBytes); err != nil {
		http.Error(w, "file too large or invalid form", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing or invalid file field", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ct := header.Header.Get("Content-Type")
	if ct != "" {
		ct = strings.TrimSpace(strings.Split(ct, ";")[0])
	}
	if !allowedImageTypes[ct] {
		http.Error(w, "unsupported file type, use JPEG/PNG/WebP", http.StatusBadRequest)
		return
	}

	if header.Size > maxBytes {
		http.Error(w, "file too large", http.StatusBadRequest)
		return
	}

	if err := os.MkdirAll(s.UploadDir, 0755); err != nil {
		http.Error(w, "server config error", http.StatusInternalServerError)
		return
	}
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		switch ct {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/webp":
			ext = ".webp"
		default:
			ext = ".bin"
		}
	}
	name := uuid.New().String() + ext
	fpath := filepath.Join(s.UploadDir, name)
	dst, err := os.Create(fpath)
	if err != nil {
		http.Error(w, "failed to save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(fpath)
		http.Error(w, "failed to save file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UploadResponse{Path: name})
}

// handleServeUpload 提供已上传图片的访问，用于前端预览；filename 仅允许单级路径（无 / 与 ..）
func (s *Server) handleServeUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	filename := chi.URLParam(r, "filename")
	if filename == "" || strings.Contains(filename, "..") || strings.ContainsRune(filename, '/') {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	absPath := filepath.Join(s.UploadDir, filename)
	// 确保解析后的路径仍在 UploadDir 内，防止路径穿越
	uploadDirAbs, _ := filepath.Abs(s.UploadDir)
	absPath, _ = filepath.Abs(absPath)
	sep := string(filepath.Separator)
	if absPath != uploadDirAbs && !strings.HasPrefix(absPath, uploadDirAbs+sep) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	f, err := os.Open(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || info.IsDir() {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	ct := "image/jpeg"
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".png":
		ct = "image/png"
	case ".webp":
		ct = "image/webp"
	}
	w.Header().Set("Content-Type", ct)
	http.ServeContent(w, r, filename, info.ModTime(), f)
}
