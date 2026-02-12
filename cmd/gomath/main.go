package main

import (
	"fmt"
	"os"

	"github.com/gomath/gomath/internal/config"
	"github.com/gomath/gomath/internal/explanation"
	"github.com/gomath/gomath/internal/http"
	"github.com/gomath/gomath/internal/ocr"
)

func main() {
	configPath := os.Getenv("GOMATH_MODELS_CONFIG")
	if configPath == "" {
		configPath = "config/models.yaml"
	}
	models, err := config.LoadModels(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(models.Status())
	ocrSvc := ocr.NewService(models.OCR)
	explainGen := explanation.NewGenerator(models.LLM.Explanation)
	explainStore := explanation.NewStore()
	// 讲解图生成：暂用 nil，后续接入文生图 API 后注入
	var imageGen http.StepImageGenerator = nil

	uploadDir := os.Getenv("GOMATH_UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "uploads"
	}
	srv := http.NewServer(uploadDir, 10, ocrSvc, explainGen, explainStore, imageGen)
	addr := os.Getenv("GOMATH_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	fmt.Println("gomath server listening on", addr)
	if err := srv.Run(addr); err != nil {
		fmt.Fprintf(os.Stderr, "server: %v\n", err)
		os.Exit(1)
	}
}
