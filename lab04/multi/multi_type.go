package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

func main() {
	ctx := context.Background()

	// 支持多种角色的消息
	template := prompt.FromMessages(
		schema.FString,
		// System: 系统提示，定义 AI 的角色和行为
		schema.SystemMessage("你是{role}，你的专长是{expertise}"),

		// User: 用户消息
		schema.UserMessage("我的问题是：{question}"),

		// Assistant: AI 的历史回复（用于多轮对话）
		schema.AssistantMessage("我理解了，让我思考一下...", []schema.ToolCall{}),

		// User: 继续对话
		schema.UserMessage("请详细说明"),
	)

	variables := map[string]any{
		"role":      "一位了解哲学的程序员",
		"expertise": "分析和取舍确定性问题",
		"question":  "在不确定的环境中，如何做出最佳决策？",
	}

	messages, err := template.Format(ctx, variables)
	if err != nil {
		log.Fatalf("格式化失败: %v", err)
	}

	for i, msg := range messages {
		fmt.Printf("%d. [%s]\n   %s\n\n", i+1, msg.Role, msg.Content)
	}
}
