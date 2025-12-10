package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/schema"
)

// generateWithRetry 尝试多次调用 ChatModel 的 Generate 方法，直到成功或达到最大重试次数。
func generateWithRetry(ctx context.Context, chatModel *deepseek.ChatModel, messages []*schema.Message, maxRetries int) (*schema.Message, error) {
	// 记录最后一次错误
	var lastErr error

	// 重试逻辑
	for i := 0; i < maxRetries; i++ {
		response, err := chatModel.Generate(ctx, messages)
		if err == nil {
			return response, nil
		}

		lastErr = err
		log.Printf("尝试 %d/%d 失败: %v", i+1, maxRetries, err)

		// 指数退避
		if i < maxRetries-1 {
			backoff := time.Duration(1<<uint(i)) * time.Second
			log.Printf("等待 %v 后重试...", backoff)
			time.Sleep(backoff)
		}
	}

	return nil, fmt.Errorf("重试 %d 次后仍然失败: %w", maxRetries, lastErr)
}

func main() {
	ctx := context.Background()

	// 创建 Deepseek ChatModel 实例
	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  os.Getenv("DEEPSEEK_API_KEY"),
		Model:   "deepseek-chat",
		BaseURL: "https://api.deepseek.com",
		// 设置超时
		Timeout: 30 * time.Second,
	})
	if err != nil {
		log.Fatalf("创建失败: %v", err)
	}

	messages := []*schema.Message{
		schema.UserMessage("你好"),
	}

	// 带重试的生成
	response, err := generateWithRetry(ctx, chatModel, messages, 3)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Fatalf("请求超时")
		}
		log.Fatalf("生成失败: %v", err)
	}

	fmt.Printf("成功! 回答: %s\\n", response.Content)
}
