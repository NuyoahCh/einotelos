// Package config 提供应用配置管理
package config

import (
	"errors"
	"os"
)

// Config 应用配置
type Config struct {
	// 火山引擎 ARK 配置
	ARK ARKConfig

	// 向量数据库配置
	VectorDB VectorDBConfig

	// RAG 配置
	RAG RAGConfig
}

// ARKConfig 火山引擎 ARK 模型配置
type ARKConfig struct {
	APIKey         string
	ChatModel      string // 对话模型
	EmbeddingModel string // 向量模型
	BaseURL        string
}

// VectorDBConfig 向量数据库配置
type VectorDBConfig struct {
	Type string // memory, redis, vikingdb
	Addr string
}

// RAGConfig RAG 检索配置
type RAGConfig struct {
	TopK           int
	ScoreThreshold float64
	ChunkSize      int
	ChunkOverlap   int
}

// Load 从环境变量加载配置
func Load() (*Config, error) {
	apiKey := os.Getenv("ARK_API_KEY")
	if apiKey == "" {
		return nil, errors.New("缺少环境变量 ARK_API_KEY")
	}

	chatModel := os.Getenv("ARK_MODEL_NAME")
	if chatModel == "" {
		chatModel = "doubao-pro-32k" // 默认使用豆包 Pro
	}

	embeddingModel := os.Getenv("ARK_EMBEDDING_MODEL")
	if embeddingModel == "" {
		embeddingModel = "doubao-embedding" // 默认向量模型
	}

	return &Config{
		ARK: ARKConfig{
			APIKey:         apiKey,
			ChatModel:      chatModel,
			EmbeddingModel: embeddingModel,
			BaseURL:        getEnvOrDefault("ARK_BASE_URL", "https://ark.cn-beijing.volces.com/api/v3"),
		},
		VectorDB: VectorDBConfig{
			Type: getEnvOrDefault("VECTOR_DB_TYPE", "memory"),
			Addr: getEnvOrDefault("VECTOR_DB_ADDR", "localhost:6379"),
		},
		RAG: RAGConfig{
			TopK:           5,
			ScoreThreshold: 0.5,
			ChunkSize:      500,
			ChunkOverlap:   50,
		},
	}, nil
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
