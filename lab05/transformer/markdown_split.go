package main

import (
	"context"
	"fmt"
	"log"

	markdown "github.com/cloudwego/eino-ext/components/document/transformer/splitter/markdown"
	"github.com/cloudwego/eino/schema"
)

func main() {
	// 创建上下文
	ctx := context.Background()

	// 创建 Markdown 标题分割器
	// Headers: 指定要识别的标题标记及其在元数据中的名称
	//   - 键：标题标记（如 "#", "##", "###"），只能由 '#' 组成
	//   - 值：在文档元数据中的字段名
	// TrimHeaders: 是否在结果中移除标题行（false: 保留标题，true: 移除标题）
	// 注意：Markdown 分割器是基于标题结构分割的，不支持 ChunkSize 和 ChunkOverlap
	splitter, err := markdown.NewHeaderSplitter(ctx, &markdown.HeaderConfig{
		Headers: map[string]string{
			"#":   "title",      // 一级标题
			"##":  "section",    // 二级标题
			"###": "subsection", // 三级标题
		},
		TrimHeaders: false, // 保留标题行
	})
	if err != nil {
		log.Fatalf("创建分割器失败: %v", err)
	}

	// Markdown 文档
	markdownContent := `
# Go 语言简介

Go 语言是 Google 开发的一门编程语言。

## 特点

### 并发支持
Go 内置了 goroutine 和 channel，让并发编程变得简单。

### 静态类型
在编译时进行类型检查，减少运行时错误。

## 应用领域

Go 广泛应用于云计算、微服务、DevOps 等领域。
	`

	doc := &schema.Document{
		Content: markdownContent,
		MetaData: map[string]any{
			"source": "go_intro.md",
		},
	}

	// 分割文档
	chunks, err := splitter.Transform(ctx, []*schema.Document{doc})
	if err != nil {
		log.Fatalf("分割文档失败: %v", err)
	}

	// 打印分割结果
	fmt.Printf("原文长度: %d 字符\\n", len(markdownContent))
	fmt.Printf("分割成 %d 块\\n\\n", len(chunks))

	for i, chunk := range chunks {
		fmt.Printf("=== 块 %d ===\\n", i+1)
		fmt.Printf("内容: %s\\n", chunk.Content)
		fmt.Printf("长度: %d 字符\\n\\n", len(chunk.Content))
	}
}
