package main

import (
	"fmt"

	"github.com/cloudwego/eino/schema"
)

func main() {

	// 长文本
	longText := `
Go 语言是 Google 开发的一门编程语言。它具有以下特点：

1. 并发支持：Go 内置了 goroutine 和 channel，让并发编程变得简单。
2. 静态类型：在编译时进行类型检查，减少运行时错误。
3. 快速编译：编译速度快，提升开发效率。
4. 简洁语法：语法简单清晰，易于学习和维护。
5. 标准库丰富：提供了完善的标准库，覆盖常见需求。

Go 广泛应用于云计算、微服务、DevOps 等领域。
	`

	// 手动分割文档（简单实现）
	chunkSize := 100
	chunkOverlap := 20

	var chunks []*schema.Document
	textRunes := []rune(longText)

	for i := 0; i < len(textRunes); i += chunkSize - chunkOverlap {
		end := i + chunkSize
		if end > len(textRunes) {
			end = len(textRunes)
		}

		chunk := &schema.Document{
			Content: string(textRunes[i:end]),
			MetaData: map[string]any{
				"source": "go_intro.txt",
				"chunk":  len(chunks) + 1,
			},
		}
		chunks = append(chunks, chunk)

		if end >= len(textRunes) {
			break
		}
	}

	// 打印分割结果
	fmt.Printf("原文长度: %d 字符\\n", len(longText))
	fmt.Printf("分割成 %d 块\\n\\n", len(chunks))

	for i, chunk := range chunks {
		fmt.Printf("=== 块 %d ===\\n", i+1)
		fmt.Printf("内容: %s\\n", chunk.Content)
		fmt.Printf("长度: %d\\n\\n", len(chunk.Content))
	}
}
