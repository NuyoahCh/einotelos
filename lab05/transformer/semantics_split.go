package main

import (
	"context"
	"fmt"
	"log"
	"os"

	semantic "github.com/cloudwego/eino-ext/components/document/transformer/splitter/semantic"
	ark "github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino/schema"
)

// main
func main02() {
	// 创建上下文
	ctx := context.Background()

	// 创建 Embedding 模型（语义分割器需要）
	// 注意：需要设置环境变量 ARK_API_KEY 和 ARK_MODEL
	// ARK_API_KEY: 从 https://cloud.bytedance.net/ark/region:ark+cn-beijing/endpoint 获取
	// ARK_MODEL: 模型端点 ID，例如 "ep-20240909094235-xxxx"（必须是支持 embedding 的模型）
	embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		APIKey: os.Getenv("ARK_API_KEY"), // 例如: "xxxxxx-xxxx-xxxx-xxxx-xxxxxxx"
		Model:  os.Getenv("ARK_MODEL"),   // 例如: "ep-20240909094235-xxxx"
	})
	if err != nil {
		log.Fatalf("创建 Embedder 失败: %v", err)
	}

	// 创建语义分割器
	// Embedding: 用于计算语义相似度的 Embedding 模型
	// Percentile: 分割阈值，如果两个块之间的差异大于 X 百分位数，则分割（默认 0.9）
	// BufferSize: 在计算 embedding 时，每个块前后拼接的块数量（用于提高准确性）
	// MinChunkSize: 最小块大小，小于此大小的块会被合并到相邻块
	// Separators: 分隔符列表，按顺序使用（默认：["\\n", ".", "?", "!"]）
	splitter, err := semantic.NewSplitter(ctx, &semantic.Config{
		Embedding:    embedder,                       // 用于计算语义相似度的 Embedding 模型
		Percentile:   0.7,                            // 分割阈值（0-1 之间，默认 0.9）
		BufferSize:   1,                              // 每个块前后拼接的块数量（可选）
		MinChunkSize: 50,                             // 最小块大小（可选）
		Separators:   []string{"\\n", ".", "?", "!"}, // 分隔符列表（可选，默认值）
	})
	if err != nil {
		log.Fatalf("创建分割器失败: %v", err)
	}

	// 创建文档
	doc := &schema.Document{
		Content: `
		第一段：Go 语言是 Google 开发的一门编程语言。它具有并发支持、静态类型、快速编译等特点。

		第二段：Go 语言广泛应用于云计算、微服务、DevOps 等领域。许多知名公司都使用 Go 语言开发核心服务。

		第三段：Go 语言的并发模型基于 CSP 理论，通过 goroutine 和 channel 实现轻量级并发。

		第四段：Go 语言的工具链包括编译器、格式化工具、测试工具等，都集成在 go 命令中。
		`,
		MetaData: map[string]any{
			"source": "go_intro.txt",
		},
	}

	// 分割文档
	chunks, err := splitter.Transform(ctx, []*schema.Document{doc})
	if err != nil {
		log.Fatalf("分割文档失败: %v", err)
	}

	// 打印分割结果
	fmt.Printf("原文长度: %d 字符\\n", len(doc.Content))
	fmt.Printf("分割成 %d 块\\n\\n", len(chunks))

	for i, chunk := range chunks {
		fmt.Printf("=== 块 %d ===\\n", i+1)
		fmt.Printf("内容: %s\\n", chunk.Content)
		fmt.Printf("长度: %d 字符\\n\\n", len(chunk.Content))
	}
}
