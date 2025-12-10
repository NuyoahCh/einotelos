package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/schema"
)

func main() {
	// This is a placeholder for the chat quickstart application.

	// 1. 创建上下文
	ctx := context.Background()

	// 2. 创建 ChatModel 实例
	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		// 提供火山 ARK 的 APIKey 模型名称的可选项
		// APIKey: os.Getenv("ARK_API_KEY"),
		// Model:  os.Getenv("ARK_MODEL_NAME"),
		APIKey:  os.Getenv("DEEPSEEK_API_KEY"), // 获取环境变量中的 APIKey 配置
		Model:   "deepseek-chat",               // 指定使用的模型名称
		BaseURL: "https://api.deepseek.com",    // 自选的 API 服务器地址
	})
	if err != nil {
		log.Fatalf("创建 ChatModel 实例失败: %v", err)
	}

	// 3. 准备发送聊天请求
	messages := []*schema.Message{
		schema.SystemMessage("你是一个知识渊博的篮球解说员"),            // 系统消息，设定对话背景
		schema.UserMessage("你好，请介绍一下 Kobe Bryant 的职业生涯。"), // 用户消息，提出问题
	}

	// 4. 调用模型生成响应
	response, err := chatModel.Generate(ctx, messages)
	if err != nil {
		log.Fatalf("生成响应失败: %v", err)
	}

	// 5. 输出结果
	fmt.Printf("AI 响应: %s\n", response.Content)

	// 6. 输出 token 使用情况（可选项）
	if response.ResponseMeta != nil && response.ResponseMeta.Usage != nil {
		fmt.Printf("\nToken 使用统计:\n")
		fmt.Printf("  输入 Token: %d\n", response.ResponseMeta.Usage.PromptTokens)
		fmt.Printf("  输出 Token: %d\n", response.ResponseMeta.Usage.CompletionTokens)
		fmt.Printf("  总计 Token: %d\n", response.ResponseMeta.Usage.TotalTokens)
	}
}
