package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/schema"
)

func main() {
	ctx := context.Background()

	// 示例1: 基础配置
	fmt.Println("=== 示例1: 基础配置 ===")
	basicExample(ctx)

	// 示例2: 高级配置
	fmt.Println("\\n=== 示例2: 高级配置 ===")
	advancedExample(ctx)

	// 示例3: 创意写作配置
	fmt.Println("\\n=== 示例3: 创意写作配置 ===")
	creativeExample(ctx)
}

// 基础配置示例
func basicExample(ctx context.Context) {
	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  os.Getenv("DEEPSEEK_API_KEY"),
		Model:   "deepseek-chat",
		BaseURL: "https://api.deepseek.com",
	})
	if err != nil {
		log.Fatalf("创建失败: %v", err)
	}

	messages := []*schema.Message{
		schema.SystemMessage("你是一个友好的 AI 助手"),
		schema.UserMessage("用一句话介绍 Eino 框架"),
	}

	response, err := chatModel.Generate(ctx, messages)
	if err != nil {
		log.Fatalf("生成失败: %v", err)
	}

	fmt.Printf("AI 响应: %s\\n", response.Content)
	printTokenUsage(response)
}

// 高级配置示例 - 精确控制输出
func advancedExample(ctx context.Context) {
	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		// 基础配置
		APIKey:  os.Getenv("DEEPSEEK_API_KEY"),
		Model:   "deepseek-chat",
		BaseURL: "https://api.deepseek.com",
		Timeout: 30 * time.Second,

		// 生成参数
		Temperature: 0.7, // 控制输出随机性，范围 [0.0, 2.0]，越高越随机
		TopP:        0.9, // 核采样参数，范围 [0.0, 1.0]，越低越聚焦
		MaxTokens:   500, // 限制最大生成 token 数量，范围 [1, 8192]

		// 停止序列 - 遇到这些文本时停止生成
		Stop: []string{"\\n\\n", "总结:"},

		// 惩罚参数 - 控制重复度
		PresencePenalty:  0.6, // 存在惩罚，范围 [-2.0, 2.0]，正值增加新话题可能性
		FrequencyPenalty: 0.5, // 频率惩罚，范围 [-2.0, 2.0]，正值减少重复词语
	})
	if err != nil {
		log.Fatalf("创建失败: %v", err)
	}

	messages := []*schema.Message{
		schema.SystemMessage("你是一个专业的技术文档撰写专家"),
		schema.UserMessage("详细介绍 Eino 框架的核心特性，包括架构、组件和优势"),
	}

	response, err := chatModel.Generate(ctx, messages)
	if err != nil {
		log.Fatalf("生成失败: %v", err)
	}

	fmt.Printf("AI 响应: %s\\n", response.Content)
	printTokenUsage(response)
}

// 创意写作配置示例 - 高随机性
func creativeExample(ctx context.Context) {
	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  os.Getenv("DEEPSEEK_API_KEY"),
		Model:   "deepseek-chat",
		BaseURL: "https://api.deepseek.com",

		// 高温度设置，适合创意写作
		Temperature: 1.2,  // 更高的随机性
		TopP:        0.95, // 保留更多可能性
		MaxTokens:   800,

		// 减少重复惩罚，允许一定的重复（适合故事情节）
		PresencePenalty:  0.3,
		FrequencyPenalty: 0.3,
	})
	if err != nil {
		log.Fatalf("创建失败: %v", err)
	}

	messages := []*schema.Message{
		schema.SystemMessage("你是一个富有创造力的故事作家"),
		schema.UserMessage("创作一个关于 AI 框架变成超级英雄的有趣故事开头"),
	}

	response, err := chatModel.Generate(ctx, messages)
	if err != nil {
		log.Fatalf("生成失败: %v", err)
	}

	fmt.Printf("AI 响应: %s\\n", response.Content)
	printTokenUsage(response)
}

// 打印 Token 使用情况
func printTokenUsage(response *schema.Message) {
	if response.ResponseMeta != nil && response.ResponseMeta.Usage != nil {
		fmt.Printf("\\nToken 使用统计:\\n")
		fmt.Printf("  输入 Token: %d\\n", response.ResponseMeta.Usage.PromptTokens)
		fmt.Printf("  输出 Token: %d\\n", response.ResponseMeta.Usage.CompletionTokens)
		fmt.Printf("  总计 Token: %d\\n", response.ResponseMeta.Usage.TotalTokens)
		if response.ResponseMeta.Usage.PromptTokenDetails.CachedTokens > 0 {
			fmt.Printf("  缓存 Token: %d\\n", response.ResponseMeta.Usage.PromptTokenDetails.CachedTokens)
		}
	}
}
