package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadModels 从统一配置文件（如 config/models.yaml）加载大模型配置。
// 返回 ocr、llm、video 各块，供识图、解析、视频模块注入使用。
// 敏感信息（API Key）通过各块内的 api_key_env 指定环境变量名，由调用方通过 OCRConfig.APIKey()、LLMExplanationConfig.APIKey() 获取。
func LoadModels(configPath string) (*Models, error) {
	if configPath == "" {
		configPath = "config/models.yaml"
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read models config: %w", err)
	}
	var m Models
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse models config: %w", err)
	}
	return &m, nil
}

// MustLoadModels 加载配置，失败则 panic（适用于 main 启动时）。
func MustLoadModels(configPath string) *Models {
	if configPath == "" {
		// 支持从可执行文件所在目录相对查找
		exe, _ := os.Executable()
		if exe != "" {
			configPath = filepath.Join(filepath.Dir(exe), "config", "models.yaml")
		} else {
			configPath = "config/models.yaml"
		}
	}
	m, err := LoadModels(configPath)
	if err != nil {
		panic("load models config: " + err.Error())
	}
	return m
}
