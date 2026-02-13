package http

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// StepImageGenerator 单步讲解图生成：描述 → 图片 URL/路径
type StepImageGenerator interface {
	Generate(ctx context.Context, prompt string) (pathOrURL string, err error)
}

// Server 提供题目输入、识图、解析等 HTTP 接口
type Server struct {
	Router       *chi.Mux
	UploadDir    string
	MaxSizeMB    int
	OCR          OCRRecognizer       // 可选
	ExplainGen   ExplainGenerator    // 可选
	ExplainStore ExplainStore        // 可选
	ImageGen     StepImageGenerator  // 可选，为每步生成讲解图
	HistoryStore HistoryStore       // 可选，解析历史
}

// NewServer 创建 HTTP 服务，uploadDir 为图片落盘目录，maxSizeMB 为单文件最大 MB；ocr/gen/store/imageGen/historyStore 可为 nil
func NewServer(uploadDir string, maxSizeMB int, ocr OCRRecognizer, explainGen ExplainGenerator, explainStore ExplainStore, imageGen StepImageGenerator, historyStore HistoryStore) *Server {
	if maxSizeMB <= 0 {
		maxSizeMB = 10
	}
	s := &Server{
		Router:       chi.NewRouter(),
		UploadDir:    uploadDir,
		MaxSizeMB:    maxSizeMB,
		OCR:          ocr,
		ExplainGen:   explainGen,
		ExplainStore: explainStore,
		ImageGen:     imageGen,
		HistoryStore: historyStore,
	}
	s.Router.Use(middleware.Logger, middleware.Recoverer)
	s.Router.Route("/api", func(r chi.Router) {
		r.Post("/upload", s.handleUpload)
		r.Get("/uploads/{filename}", s.handleServeUpload)
		r.Post("/submit", s.handleSubmit)
		r.Post("/explain", s.handleExplain)
		r.Get("/result/{id}", s.handleResult)
		r.Get("/history/find-upload", s.handleHistoryFindLatestUpload)
		r.Get("/history", s.handleHistoryList)
		r.Post("/history", s.handleHistoryCreate)
		r.Patch("/history/{id}", s.handleHistoryUpdateResult)
		r.Delete("/history/{id}", s.handleHistoryDelete)
	})
	return s
}

// Run 监听 addr
func (s *Server) Run(addr string) error {
	return http.ListenAndServe(addr, s.Router)
}
