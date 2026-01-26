package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/NuyoahCh/einotelos/lab11/common"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// ImageUnderstanding 演示如何使用 Eino 框架进行图像理解
// 本示例展示了如何向 LLM 发送包含图像的消息，并获取对图像内容的理解和描述
func main() {
	ctx := context.Background()

	// 创建支持多模态的 ChatModel 实例
	// DeepSeek 支持图像理解功能
	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  os.Getenv("DEEPSEEK_API_KEY"),
		Model:   "deepseek-chat", // 使用支持视觉的模型
		BaseURL: "https://api.deepseek.com",
	})
	if err != nil {
		log.Fatalf("创建 ChatModel 实例失败: %v", err)
	}

	// 准备包含图像的消息
	// 图像可以通过 URL 或 Base64 编码提供
	imageURL := "https://example.com/sample-image.jpg" // 替换为实际的图像 URL

	messages := []*schema.Message{
		schema.SystemMessage("你是一个专业的图像分析助手，能够详细描述图像内容。"),
		{
			Role: schema.User,
			UserInputMultiContent: []schema.MessageInputPart{
				common.NewTextPart("请详细描述这张图片的内容，包括主要对象、场景、颜色和氛围。"),
				common.NewImageURLPart(imageURL),
			},
		},
	}

	// 调用模型进行图像理解
	fmt.Println("正在分析图像...")
	response, err := chatModel.Generate(ctx, messages)
	if err != nil {
		log.Fatalf("图像理解失败: %v", err)
	}

	// 输出图像分析结果
	fmt.Printf("\n图像分析结果:\n%s\n", response.Content)

	// 输出 token 使用情况
	if response.ResponseMeta != nil && response.ResponseMeta.Usage != nil {
		fmt.Printf("\nToken 使用统计:\n")
		fmt.Printf("  输入 Token: %d\n", response.ResponseMeta.Usage.PromptTokens)
		fmt.Printf("  输出 Token: %d\n", response.ResponseMeta.Usage.CompletionTokens)
		fmt.Printf("  总计 Token: %d\n", response.ResponseMeta.Usage.TotalTokens)
	}

	// 进阶示例：多轮对话中的图像理解
	fmt.Println("\n--- 进阶示例：基于图像的多轮对话 ---")
	multiRoundImageChat(ctx, chatModel, imageURL)
}

// multiRoundImageChat 演示如何在多轮对话中使用图像
func multiRoundImageChat(ctx context.Context, chatModel model.ChatModel, imageURL string) {
	// 第一轮：询问图像内容
	messages := []*schema.Message{
		schema.SystemMessage("你是一个图像分析专家。"),
		{
			Role: schema.User,
			UserInputMultiContent: []schema.MessageInputPart{
				common.NewTextPart("这张图片中有什么？"),
				common.NewImageURLPart(imageURL),
			},
		},
	}

	response1, err := chatModel.Generate(ctx, messages)
	if err != nil {
		log.Printf("第一轮对话失败: %v", err)
		return
	}
	fmt.Printf("\n第一轮 - AI: %s\n", response1.Content)

	// 第二轮：基于第一轮的回答继续提问
	messages = append(messages, response1)
	messages = append(messages, schema.UserMessage("图片中的主要颜色是什么？"))

	response2, err := chatModel.Generate(ctx, messages)
	if err != nil {
		log.Printf("第二轮对话失败: %v", err)
		return
	}
	fmt.Printf("第二轮 - AI: %s\n", response2.Content)

	// 第三轮：询问更具体的细节
	messages = append(messages, response2)
	messages = append(messages, schema.UserMessage("能否估计图片的拍摄时间或场景？"))

	response3, err := chatModel.Generate(ctx, messages)
	if err != nil {
		log.Printf("第三轮对话失败: %v", err)
		return
	}
	fmt.Printf("第三轮 - AI: %s\n", response3.Content)
}
