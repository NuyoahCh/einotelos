package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/cloudwego/eino/components/document/parser"
)

// main
func main01() {
	// 创建上下文
	ctx := context.Background()

	// 创建文本解析器
	textParser := parser.TextParser{}

	// 解析文本内容
	textContent := "这是要解析的文本内容。\\n可以包含多行。"
	reader := strings.NewReader(textContent)

	// 使用 WithURI 指定文档 URI（可选，但建议提供）
	docs, err := textParser.Parse(ctx, reader, parser.WithURI("text://sample.txt"))
	if err != nil {
		log.Fatalf("解析失败: %v", err)
	}

	// 打印解析结果
	fmt.Printf("解析了 %d 个文档\\n", len(docs))
	for i, doc := range docs {
		fmt.Printf("\\n=== 文档 %d ===\\n", i+1)
		fmt.Printf("内容: %s\\n", doc.Content)
		fmt.Printf("元数据: %v\\n", doc.MetaData)
	}
}
