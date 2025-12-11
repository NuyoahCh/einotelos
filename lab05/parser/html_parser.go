package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	htmlparser "github.com/cloudwego/eino-ext/components/document/parser/html"
	"github.com/cloudwego/eino/components/document/parser"
)

// main
func main02() {
	// 创建上下文
	ctx := context.Background()

	// 创建 HTML 解析器
	// 根据 eino-ext 源码，使用 NewParser 和 Config（不是 ParserConfig）
	// Selector 可选：指定要提取的内容选择器，例如 "body" 表示只提取 <body> 标签的内容
	// 如果不指定 Selector，会提取整个文档的内容
	htmlParser, err := htmlparser.NewParser(ctx, &htmlparser.Config{
		Selector: &htmlparser.BodySelector, // 可选：只提取 body 标签的内容
	})
	if err != nil {
		log.Fatalf("创建 HTML 解析器失败: %v", err)
	}

	// HTML 内容
	htmlContent := `
	<html lang="zh-CN">
		<head>
			<meta charset="UTF-8">
			<meta name="description" content="这是一个示例文档">
			<title>示例文档</title>
		</head>
		<body>
			<h1>标题</h1>
			<p>这是段落内容。</p>
			<ul>
				<li>列表项 1</li>
				<li>列表项 2</li>
			</ul>
		</body>
	</html>
	`

	// 解析 HTML
	reader := strings.NewReader(htmlContent)
	docs, err := htmlParser.Parse(ctx, reader, parser.WithURI("https://example.com/page.html"))
	if err != nil {
		log.Fatalf("解析 HTML 失败: %v", err)
	}

	// 打印解析结果
	fmt.Printf("解析了 %d 个文档\\n", len(docs))
	for i, doc := range docs {
		fmt.Printf("\\n=== 文档 %d ===\\n", i+1)
		fmt.Printf("内容: %s\\n", doc.Content)
		fmt.Printf("元数据: %v\\n", doc.MetaData)

		// 打印一些常见的元数据字段
		if title, ok := doc.MetaData[htmlparser.MetaKeyTitle].(string); ok {
			fmt.Printf("标题: %s\\n", title)
		}
		if desc, ok := doc.MetaData[htmlparser.MetaKeyDesc].(string); ok {
			fmt.Printf("描述: %s\\n", desc)
		}
		if lang, ok := doc.MetaData[htmlparser.MetaKeyLang].(string); ok {
			fmt.Printf("语言: %s\\n", lang)
		}
		if charset, ok := doc.MetaData[htmlparser.MetaKeyCharset].(string); ok {
			fmt.Printf("字符集: %s\\n", charset)
		}
	}
}
