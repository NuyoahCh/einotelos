package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/schema"
)

func main() {
	ctx := context.Background()

	// 创建 ChatModel
	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  os.Getenv("DEEPSEEK_API_KEY"),
		Model:   "deepseek-chat",
		BaseURL: "https://api.deepseek.com",
	})
	if err != nil {
		log.Fatalf("创建失败: %v", err)
	}

	// 构建消息
	messages := []*schema.Message{
		schema.SystemMessage("你是一个懂得哲学的程序员。"),
		schema.UserMessage("什么是存在主义？"),
	}

	// 流式生成
	stream, err := chatModel.Stream(ctx, messages)
	if err != nil {
		log.Fatalf("流式生成失败: %v", err)
	}
	defer stream.Close() // 记得关闭流

	fmt.Print("AI 回复: ")

	// 逐块接收并打印
	for {
		chunk, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				// 流结束
				break
			}
			log.Fatalf("接收失败: %v", err)
		}

		// 打印内容（打字机效果）
		fmt.Print(chunk.Content)

		// 同时收集完整内容
		// fullContent.WriteString(chunk.Content)，通过字符串拼接收集完整内容
	}

	fmt.Println("\n\n完成！")
}
