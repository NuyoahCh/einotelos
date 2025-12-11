package main

import (
	"context"
	"fmt"
	"log"

	recursive "github.com/cloudwego/eino-ext/components/document/transformer/splitter/recursive"
	"github.com/cloudwego/eino/schema"
)

// main
func main01() {
	// 创建上下文
	ctx := context.Background()

	// 创建递归字符分割器
	// ChunkSize: 每个块的最大字符数
	// OverlapSize: 块之间的重叠字符数（用于保持上下文连续性）
	// Separators: 分隔符列表，按优先级顺序使用（默认：["\\n", ".", "?", "!"]）
	// KeepType: 分隔符保留方式（KeepTypeNone: 丢弃，KeepTypeStart: 保留在开头，KeepTypeEnd: 保留在结尾）
	splitter, err := recursive.NewSplitter(ctx, &recursive.Config{
		ChunkSize:   500,                           // 每个块最多 500 个字符
		OverlapSize: 50,                            // 块之间重叠 50 个字符
		Separators:  []string{"\\n", ".", "?", "!"}, // 分隔符列表（可选，默认值）
		KeepType:    recursive.KeepTypeNone,        // 丢弃分隔符
	})
	if err != nil {
		log.Fatalf("创建分割器失败: %v", err)
	}

	// 创建长文档
	longText := `
	Go 语言是 Google 开发的一门编程语言。它具有以下特点：

	1. 并发支持：Go 内置了 goroutine 和 channel，让并发编程变得简单。
	2. 静态类型：在编译时进行类型检查，减少运行时错误。
	3. 快速编译：编译速度快，提升开发效率。
	4. 简洁语法：语法简单清晰，易于学习和维护。
	5. 标准库丰富：提供了完善的标准库，覆盖常见需求。

	Go 广泛应用于云计算、微服务、DevOps 等领域。许多知名公司如 Google、Uber、Docker 等都使用 Go 语言开发核心服务。

	Go 语言的并发模型基于 CSP（Communicating Sequential Processes）理论，通过 goroutine 和 channel 实现轻量级并发。
	每个 goroutine 只占用几 KB 的内存，可以轻松创建成千上万个 goroutine。

	Go 语言的工具链包括编译器、格式化工具、测试工具等，都集成在 go 命令中，使用非常方便。
	`

	doc := &schema.Document{
		Content: longText,
		MetaData: map[string]any{
			"source": "go_intro.txt",
			"title":  "Go 语言简介",
		},
	}

	// 分割文档
	chunks, err := splitter.Transform(ctx, []*schema.Document{doc})
	if err != nil {
		log.Fatalf("分割文档失败: %v", err)
	}

	// 打印分割结果
	fmt.Printf("原文长度: %d 字符\\n", len(longText))
	fmt.Printf("分割成 %d 块\\n\\n", len(chunks))

	for i, chunk := range chunks {
		fmt.Printf("=== 块 %d ===\\n", i+1)
		fmt.Printf("内容: %s\\n", chunk.Content)
		fmt.Printf("长度: %d 字符\\n", len(chunk.Content))
		fmt.Printf("元数据: %v\\n\\n", chunk.MetaData)
	}
}

