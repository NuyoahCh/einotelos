package main

import (
	"context"
	"fmt"
	"log"

	fileloader "github.com/cloudwego/eino-ext/components/document/loader/file"
	htmlparser "github.com/cloudwego/eino-ext/components/document/parser/html"
	recursive "github.com/cloudwego/eino-ext/components/document/transformer/splitter/recursive"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/schema"
)

func main() {
	// 创建上下文
	ctx := context.Background()

	// 步骤 1: 创建文档加载器（从本地文件加载）
	// FileLoaderConfig 可以指定 Parser，如果不指定，默认使用 ExtParser
	fileLoader, err := fileloader.NewFileLoader(ctx, &fileloader.FileLoaderConfig{})
	if err != nil {
		log.Fatalf("创建文件加载器失败: %v", err)
	}

	// 步骤 2: 加载 HTML 文件
	docs, err := fileLoader.Load(ctx, document.Source{
		URI: "./testdata/article.html", // 根据实际文件路径修改
	})
	if err != nil {
		log.Fatalf("加载文档失败: %v", err)
	}
	fmt.Printf("加载了 %d 个文档\\n", len(docs))

	// 步骤 3: 创建 HTML 解析器（可选）
	// 注意：FileLoader 默认已经使用 ExtParser 解析了文件
	// 如果文件是 HTML 格式，FileLoader 会自动使用 HTML 解析器
	// 这里我们直接使用 Loader 返回的文档，不需要再次解析
	// 如果需要自定义解析，可以创建 HTML 解析器：
	htmlParser, err := htmlparser.NewParser(ctx, &htmlparser.Config{
		Selector: &htmlparser.BodySelector, // 只提取 body 标签的内容
	})
	if err != nil {
		log.Fatalf("创建 HTML 解析器失败: %v", err)
	}

	// 步骤 4: 使用 Loader 返回的文档
	// 注意：FileLoader 已经使用 ExtParser 解析了文件
	// 如果文件是 HTML 格式，FileLoader 会自动使用 HTML 解析器
	// 这里我们直接使用 Loader 返回的文档
	parsedDocs := docs
	fmt.Printf("解析后得到 %d 个文档\\n", len(parsedDocs))
	_ = htmlParser // 避免未使用变量的警告（如果需要重新解析，可以使用 htmlParser）

	// 步骤 5: 创建文档分割器
	splitter, err := recursive.NewSplitter(ctx, &recursive.Config{
		ChunkSize:   500,                            // 每个块最多 500 个字符
		OverlapSize: 50,                             // 块之间重叠 50 个字符
		Separators:  []string{"\\n", ".", "?", "!"}, // 分隔符列表（可选）
		KeepType:    recursive.KeepTypeNone,         // 丢弃分隔符
	})
	if err != nil {
		log.Fatalf("创建分割器失败: %v", err)
	}

	// 步骤 6: 分割文档
	var allChunks []*schema.Document
	for _, doc := range parsedDocs {
		chunks, err := splitter.Transform(ctx, []*schema.Document{doc})
		if err != nil {
			log.Printf("分割文档失败: %v", err)
			continue
		}
		allChunks = append(allChunks, chunks...)
	}
	fmt.Printf("分割后得到 %d 个文档块\\n\\n", len(allChunks))

	// 步骤 7: 打印结果
	for i, chunk := range allChunks {
		fmt.Printf("=== 块 %d ===\\n", i+1)
		fmt.Printf("内容长度: %d 字符\\n", len(chunk.Content))
		fmt.Printf("内容预览: %s...\\n", chunk.Content[:min(100, len(chunk.Content))])
		fmt.Printf("元数据: %v\\n\\n", chunk.MetaData)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
