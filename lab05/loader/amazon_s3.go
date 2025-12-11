package main

import (
	"context"
	"fmt"
	"log"

	s3loader "github.com/cloudwego/eino-ext/components/document/loader/s3"
	"github.com/cloudwego/eino/components/document"
)

func main() {
	// 创建上下文
	ctx := context.Background()

	// 创建 S3 加载器
	// 根据 eino-ext 源码，使用 NewS3Loader 和 LoaderConfig
	// Region 是 *string 类型（指针），需要使用字符串指针
	// 如果不提供 Region，会使用默认的 AWS 配置
	region := "us-east-1" // AWS 区域
	loader, err := s3loader.NewS3Loader(ctx, &s3loader.LoaderConfig{
		Region: &region, // AWS 区域（指针类型）
		// 可选：如果提供 AWSAccessKey 和 AWSSecretKey，必须同时提供
		// AWSAccessKey: aws.String("your-access-key"),
		// AWSSecretKey: aws.String("your-secret-key"),
		// 如果不提供，会使用默认的 AWS 凭证配置（环境变量或配置文件）
	})
	if err != nil {
		log.Fatalf("创建加载器失败: %v", err)
	}

	// 从 S3 加载文档
	// URI 格式: s3://bucket-name/path/to/file.pdf
	// 注意：这是一个示例 URI，实际使用时请替换为你的 S3 bucket 和文件路径
	docs, err := loader.Load(ctx, document.Source{
		URI: "s3://my-bucket/documents/article.pdf",
	})
	if err != nil {
		log.Fatalf("加载文档失败: %v", err)
	}

	// 打印加载的文档
	fmt.Printf("加载了 %d 个文档\\n", len(docs))
	for i, doc := range docs {
		fmt.Printf("\\n=== 文档 %d ===\\n", i+1)
		// 只打印前 200 个字符，避免输出过长
		content := doc.Content
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		fmt.Printf("内容预览: %s\\n", content)
		fmt.Printf("内容长度: %d 字符\\n", len(doc.Content))
		fmt.Printf("元数据: %v\\n", doc.MetaData)
	}
}
