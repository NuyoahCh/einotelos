package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/schema"
)

// DocumentQA 文档问答系统
type DocumentQA struct {
	embedder  *ark.Embedder
	chatModel *deepseek.ChatModel
	documents []*schema.Document
	vectors   [][]float64
}

func NewDocumentQA(arkAPIKey, arkModel, deepseekAPIKey string) (*DocumentQA, error) {
	ctx := context.Background()

	// 创建 ARK Embedding
	embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		APIKey: arkAPIKey,
		Model:  arkModel,
	})
	if err != nil {
		return nil, err
	}

	// 创建 DeepSeek ChatModel
	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  deepseekAPIKey,
		Model:   "deepseek-chat",
		BaseURL: "https://api.deepseek.com",
	})
	if err != nil {
		return nil, err
	}

	return &DocumentQA{
		embedder:  embedder,
		chatModel: chatModel,
	}, nil
}

// LoadDocuments 加载文档
func (qa *DocumentQA) LoadDocuments(ctx context.Context, docs []*schema.Document) error {
	fmt.Printf("正在加载 %d 个文档...\\n", len(docs))

	// 提取文本内容
	texts := make([]string, len(docs))
	for i, doc := range docs {
		texts[i] = doc.Content
	}

	// 向量化
	vectors, err := qa.embedder.EmbedStrings(ctx, texts)
	if err != nil {
		return err
	}

	qa.documents = docs
	qa.vectors = vectors

	fmt.Printf("成功加载并向量化 %d 个文档\\n", len(docs))
	return nil
}

// Query 查询
func (qa *DocumentQA) Query(ctx context.Context, question string) (string, error) {
	// 1. 向量化问题
	questionVectors, err := qa.embedder.EmbedStrings(ctx, []string{question})
	if err != nil {
		return "", err
	}
	questionVector := questionVectors[0]

	// 2. 计算相似度，找到最相关的文档
	type docScore struct {
		doc   *schema.Document
		score float64
	}

	scores := make([]docScore, len(qa.documents))
	for i := range qa.documents {
		similarity := cosineSimilarity(questionVector, qa.vectors[i])
		scores[i] = docScore{
			doc:   qa.documents[i],
			score: similarity,
		}
	}

	// 3. 排序，取 Top 3
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	topDocs := make([]*schema.Document, 0, 3)
	for i := 0; i < 3 && i < len(scores); i++ {
		if scores[i].score > 0.7 { // 相似度阈值
			topDocs = append(topDocs, scores[i].doc)
		}
	}

	if len(topDocs) == 0 {
		return "抱歉，我在文档中找不到相关信息。", nil
	}

	// 4. 构建上下文
	context := "相关文档内容：\\n\\n"
	for i, doc := range topDocs {
		context += fmt.Sprintf("%d. %s\\n\\n", i+1, doc.Content)
	}

	// 5. 生成回答
	messages := []*schema.Message{
		schema.SystemMessage(fmt.Sprintf(`你是一个专业的文档问答助手。
请根据以下文档内容回答用户问题。如果文档中没有相关信息，请如实告知。

%s`, context)),
		schema.UserMessage(question),
	}

	response, err := qa.chatModel.Generate(ctx, messages)
	if err != nil {
		return "", err
	}

	return response.Content, nil
}

func main() {
	// 创建问答系统
	qa, err := NewDocumentQA(
		os.Getenv("ARK_API_KEY"),
		os.Getenv("ARK_EMBEDDING_MODEL"),
		os.Getenv("DEEPSEEK_API_KEY"),
	)
	if err != nil {
		log.Fatalf("创建问答系统失败: %v", err)
	}

	// 加载知识库
	docs := []*schema.Document{
		{Content: "Eino 是基于 Go 语言的 AI 应用开发框架，由字节跳动开源。"},
		{Content: "Eino 提供了 ChatModel、Embedding、Retriever 等丰富的组件。"},
		{Content: "Eino 支持 Chain 和 Graph 两种编排方式，可以灵活组合组件。"},
		{Content: "React Agent 是 Eino 中的智能代理，能够自主调用工具完成任务。"},
		{Content: "Eino 支持多种大模型，包括 OpenAI、ARK、Ollama 等。"},
	}

	ctx := context.Background()
	if err := qa.LoadDocuments(ctx, docs); err != nil {
		log.Fatalf("加载文档失败: %v", err)
	}

	// 交互式问答
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("\\n=== 文档问答系统（输入 'exit' 退出）===")

	for {
		fmt.Print("问题: ")
		if !scanner.Scan() {
			break
		}

		question := strings.TrimSpace(scanner.Text())
		if question == "exit" {
			fmt.Println("再见！")
			break
		}

		if question == "" {
			continue
		}

		answer, err := qa.Query(ctx, question)
		if err != nil {
			fmt.Printf("查询失败: %v\\n", err)
			continue
		}

		fmt.Printf("\\n回答: %s\\n\\n", answer)
	}
}

func cosineSimilarity(a, b []float64) float64 {
	// 实现同前面的示例
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
