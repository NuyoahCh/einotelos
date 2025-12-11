package main

import (
	"context"
	"fmt"
	"log"
	"os"

	htmlparser "github.com/cloudwego/eino-ext/components/document/parser/html"
	pdfparser "github.com/cloudwego/eino-ext/components/document/parser/pdf"
	"github.com/cloudwego/eino/components/document/parser"
)

func main() {
	// 创建上下文
	ctx := context.Background()

	// 创建各个解析器
	// HTML 解析器：使用 Config（不是 ParserConfig）
	htmlParser, err := htmlparser.NewParser(ctx, &htmlparser.Config{})
	if err != nil {
		log.Fatalf("创建 HTML 解析器失败: %v", err)
	}

	// PDF 解析器：使用 NewPDFParser 和 Config（不是 NewParser 和 ParserConfig）
	pdfParser, err := pdfparser.NewPDFParser(ctx, &pdfparser.Config{
		ToPages: false, // 合并所有页面为一个文档
	})
	if err != nil {
		log.Fatalf("创建 PDF 解析器失败: %v", err)
	}

	// 创建扩展解析器，注册不同扩展名的解析器
	extParser, err := parser.NewExtParser(ctx, &parser.ExtParserConfig{
		Parsers: map[string]parser.Parser{
			".html": htmlParser, // HTML 文件使用 HTML 解析器
			".htm":  htmlParser, // HTM 文件也使用 HTML 解析器
			".pdf":  pdfParser,  // PDF 文件使用 PDF 解析器
		},
		// FallbackParser 是可选的，如果不提供，默认使用 TextParser
	})
	if err != nil {
		log.Fatalf("创建扩展解析器失败: %v", err)
	}

	// 解析不同格式的文件
	files := []string{
		"./testdata/sample.html",
		"./testdata/sample.pdf",
		"./testdata/sample.txt", // 没有注册的扩展名，会使用 FallbackParser
	}

	for _, filePath := range files {
		file, err := os.Open(filePath)
		if err != nil {
			log.Printf("打开文件失败 %s: %v", filePath, err)
			continue
		}

		// 必须使用 WithURI 指定文件路径，ExtParser 需要根据扩展名选择解析器
		docs, err := extParser.Parse(ctx, file, parser.WithURI(filePath))
		file.Close()

		if err != nil {
			log.Printf("解析文件失败 %s: %v", filePath, err)
			continue
		}

		fmt.Printf("\\n文件: %s\\n", filePath)
		fmt.Printf("解析了 %d 个文档\\n", len(docs))
		for i, doc := range docs {
			fmt.Printf("  文档 %d: 长度 %d 字符\\n", i+1, len(doc.Content))
		}
	}
}
