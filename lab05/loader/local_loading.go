package main

import (
	"context"
	"fmt"
	"log"

	fileloader "github.com/cloudwego/eino-ext/components/document/loader/file"
	"github.com/cloudwego/eino/components/document"
)

// main
func main01() {
	// 创建上下文
	ctx := context.Background()

	// 创建本地文件加载器
	// 根据 eino-ext 源码，使用 NewFileLoader 和 FileLoaderConfig
	loader, err := fileloader.NewFileLoader(ctx, &fileloader.FileLoaderConfig{
		// UseNameAsID: true, // 可选：使用文件名作为文档 ID
	})
	if err != nil {
		log.Fatalf("创建加载器失败: %v", err)
	}

	// 加载文档
	// Source.URI 应该是文件的绝对路径或相对路径
	docs, err := loader.Load(ctx, document.Source{
		URI: "./testdata/sample.txt", // 文件路径
	})
	if err != nil {
		log.Fatalf("加载文档失败: %v", err)
	}

	// 打印加载的文档
	fmt.Printf("加载了 %d 个文档\\n", len(docs))
	for i, doc := range docs {
		fmt.Printf("\\n=== 文档 %d ===\\n", i+1)
		fmt.Printf("内容: %s\\n", doc.Content)
		fmt.Printf("元数据: %v\\n", doc.MetaData)

		// 打印一些常见的元数据字段
		if fileName, ok := doc.MetaData[fileloader.MetaKeyFileName].(string); ok {
			fmt.Printf("文件名: %s\\n", fileName)
		}
		if extension, ok := doc.MetaData[fileloader.MetaKeyExtension].(string); ok {
			fmt.Printf("扩展名: %s\\n", extension)
		}
		if source, ok := doc.MetaData[fileloader.MetaKeySource].(string); ok {
			fmt.Printf("源路径: %s\\n", source)
		}
	}
}
