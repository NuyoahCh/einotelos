package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/schema"
)

func main() {
	ctx := context.Background()

	// 1. 创建 ARK Embedding 模型
	embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		APIKey: os.Getenv("ARK_API_KEY"),
		Model:  os.Getenv("ARK_EMBEDDING_MODEL"),
	})
	if err != nil {
		log.Fatalf("创建 Embedder 失败: %v", err)
	}

	// 2. 准备知识库文档
	docs := []*schema.Document{
		{
			Content: "Go 语言由 Google 开发，2009年发布。它是一门静态类型、编译型语言。",
			MetaData: map[string]any{
				"source": "go_basics",
				"type":   "intro",
			},
		},
		{
			Content: "Go 的并发模型基于 CSP（通信顺序进程）。主要通过 goroutine 和 channel 实现。",
			MetaData: map[string]any{
				"source": "go_concurrency",
				"type":   "advanced",
			},
		},
		{
			Content: "Eino 是用 Go 开发的 AI 应用框架，提供了丰富的组件和灵活的编排能力。",
			MetaData: map[string]any{
				"source": "eino_intro",
				"type":   "framework",
			},
		},
	}

	// 3. 向量化所有文档（内存向量存储）
	fmt.Println("=== 向量化文档 ===")
	var vectors [][]float64
	for _, doc := range docs {
		vec, err := embedder.EmbedStrings(ctx, []string{doc.Content})
		if err != nil {
			log.Fatalf("向量化失败: %v", err)
		}
		vectors = append(vectors, vec[0])
	}
	fmt.Printf("✅ 成功向量化 %d 个文档\\n\\n", len(vectors))

	// 4. 执行语义检索
	query := "Go 语言的并发是如何实现的？"
	fmt.Printf("=== 检索: %s ===\\n", query)

	// 向量化查询
	queryVec, err := embedder.EmbedStrings(ctx, []string{query})
	if err != nil {
		log.Fatalf("查询向量化失败: %v", err)
	}

	// 计算相似度并排序
	type docScore struct {
		doc   *schema.Document
		score float64
	}

	var scores []docScore
	for i, doc := range docs {
		similarity := cosineSimilarity(queryVec[0], vectors[i])
		scores = append(scores, docScore{doc: doc, score: similarity})
	}

	// 简单排序获取 Top 2
	for i := 0; i < len(scores); i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[i].score < scores[j].score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}

	topK := 2
	if len(scores) < topK {
		topK = len(scores)
	}

	// 获取检索到的文档
	retrievedDocs := make([]*schema.Document, topK)
	for i := 0; i < topK; i++ {
		retrievedDocs[i] = scores[i].doc
	}

	// 显示检索结果
	fmt.Printf("✅ 检索到 %d 个相关文档:\\n\\n", len(retrievedDocs))
	for i, doc := range retrievedDocs {
		fmt.Printf("文档 %d:\\n", i+1)
		fmt.Printf("  内容: %s\\n", doc.Content)
		fmt.Printf("  元数据: %v\\n\\n", doc.MetaData)
	}

	// 5. 使用检索结果增强 LLM 回答（RAG）
	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  os.Getenv("DEEPSEEK_API_KEY"),
		Model:   "deepseek-chat",
		BaseURL: "https://api.deepseek.com",
	})
	if err != nil {
		log.Fatalf("创建 ChatModel 失败: %v", err)
	}

	// 构建上下文
	contextStr := ""
	for _, doc := range retrievedDocs {
		contextStr += doc.Content + "\\n"
	}

	messages := []*schema.Message{
		schema.SystemMessage(fmt.Sprintf("你是一个专业的技术助手。请基于以下知识回答用户问题：\\n\\n%s", contextStr)),
		schema.UserMessage(query),
	}

	fmt.Println("=== AI 回答（基于检索的知识）===")
	response, err := chatModel.Generate(ctx, messages)
	if err != nil {
		log.Fatalf("生成失败: %v", err)
	}

	fmt.Println(response.Content)
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

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
