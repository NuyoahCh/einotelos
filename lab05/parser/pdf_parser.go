package main

import (
	"context"
	"fmt"
	"log"
	"os"

	pdfparser "github.com/cloudwego/eino-ext/components/document/parser/pdf"
	"github.com/cloudwego/eino/components/document/parser"
)

func main() {
	// 创建上下文
	ctx := context.Background()

	// 创建 PDF 解析器
	// 根据 eino-ext 源码，使用 NewPDFParser 和 Config（不是 ParserConfig）
	// ToPages: 如果为 true，会将 PDF 按页分割成多个文档；如果为 false，会将所有页面合并为一个文档
	pdfParser, err := pdfparser.NewPDFParser(ctx, &pdfparser.Config{
		ToPages: false, // false: 合并所有页面为一个文档；true: 每页一个文档
	})
	if err != nil {
		log.Fatalf("创建 PDF 解析器失败: %v", err)
	}

	// 打开 PDF 文件
	// 注意：这是一个示例路径，实际使用时请替换为你的 PDF 文件路径
	file, err := os.Open("./testdata/sample.pdf")
	if err != nil {
		log.Fatalf("打开文件失败: %v\\n提示：请确保 ./testdata/sample.pdf 文件存在", err)
	}
	defer file.Close()

	// 解析 PDF
	// 可以使用 WithToPages 选项在运行时覆盖配置中的 ToPages 设置
	docs, err := pdfParser.Parse(ctx, file,
		parser.WithURI("./testdata/sample.pdf"),
		// pdfparser.WithToPages(true), // 可选：运行时指定按页分割
	)
	if err != nil {
		log.Fatalf("解析 PDF 失败: %v", err)
	}

	// 打印解析结果
	fmt.Printf("解析了 %d 个文档\\n", len(docs))
	for i, doc := range docs {
		fmt.Printf("\\n=== 文档 %d ===\\n", i+1)
		fmt.Printf("内容长度: %d 字符\\n", len(doc.Content))
		if len(doc.Content) > 0 {
			previewLen := min(100, len(doc.Content))
			fmt.Printf("内容预览: %s...\\n", doc.Content[:previewLen])
		}
		fmt.Printf("元数据: %v\\n", doc.MetaData)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
