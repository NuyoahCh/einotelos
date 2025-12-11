package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
)

func main() {
	ctx := context.Background()

	// 创建 ARK Embedding 模型
	embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		APIKey: os.Getenv("ARK_API_KEY"),
		Model:  os.Getenv("ARK_EMBEDDING_MODEL"), // 例如: doubao-embedding-large
	})
	if err != nil {
		log.Fatalf("创建 Embedder 失败: %v", err)
	}

	// 待向量化的文本
	texts := []string{
		"Go 是一门编程语言",
		"Python 是一门编程语言",
		"今天天气真好",
	}

	// 生成向量
	vectors, err := embedder.EmbedStrings(ctx, texts)
	if err != nil {
		log.Fatalf("向量化失败: %v", err)
	}

	// 打印结果
	for i, text := range texts {
		fmt.Printf("文本 %d: %s\\n", i+1, text)
		fmt.Printf("  向量维度: %d\\n", len(vectors[i]))
		fmt.Printf("  前5维: %v\\n\\n", vectors[i][:5])
	}

	// 计算相似度
	similarity12 := cosineSimilarity(vectors[0], vectors[1])
	similarity13 := cosineSimilarity(vectors[0], vectors[2])

	fmt.Printf("文本1 和 文本2 的相似度: %.4f\\n", similarity12)
	fmt.Printf("文本1 和 文本3 的相似度: %.4f\\n", similarity13)
}

// 计算余弦相似度
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (sqrt(normA) * sqrt(normB))
}

func sqrt(x float64) float64 {
	return math.Sqrt(x)
}
