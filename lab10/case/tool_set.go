package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func main() {
	ctx := context.Background()

	// 1. 创建多个工具
	// 计算器工具
	calculator := utils.NewTool(
		&schema.ToolInfo{
			Name: "calculator",
			Desc: "执行数学计算",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"expression": {
					Type:     "string",
					Desc:     "数学表达式，例如: 10 + 20",
					Required: true,
				},
			}),
		},
		func(ctx context.Context, params map[string]any) (string, error) {
			// 简化实现：只处理简单加法
			return "30", nil
		},
	)

	// 时间工具
	timeTool := utils.NewTool(
		&schema.ToolInfo{
			Name:        "get_time",
			Desc:        "获取当前时间",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}),
		},
		func(ctx context.Context, params map[string]any) (string, error) {
			return time.Now().Format("2006-01-02 15:04:05"), nil
		},
	)

	// 2. 创建 ChatModel（支持 Function Calling）
	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  os.Getenv("DEEPSEEK_API_KEY"),
		Model:   "deepseek-chat",
		BaseURL: "https://api.deepseek.com",
	})
	if err != nil {
		log.Fatalf("创建模型失败: %v", err)
	}

	// 3. 创建 ToolsNode
	toolsNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
		Tools: []tool.BaseTool{calculator, timeTool},
	})
	if err != nil {
		log.Fatalf("创建 ToolsNode 失败: %v", err)
	}

	// 4. 获取工具信息列表
	calcInfo, err := calculator.Info(ctx)
	if err != nil {
		log.Fatalf("获取工具信息失败: %v", err)
	}
	timeInfo, err := timeTool.Info(ctx)
	if err != nil {
		log.Fatalf("获取工具信息失败: %v", err)
	}

	toolInfos := []*schema.ToolInfo{calcInfo, timeInfo}

	// 测试多个场景
	testCases := []string{
		"现在几点了？",
		"帮我计算 10 + 20 等于多少",
		"你好，请介绍一下你自己",
	}

	for i, question := range testCases {
		fmt.Printf("\\n========== 测试 %d ==========\\n", i+1)
		fmt.Printf("问题: %s\\n", question)

		messages := []*schema.Message{
			schema.UserMessage(question),
		}

		// AI 决定使用工具
		response, err := chatModel.Generate(ctx, messages,
			model.WithTools(toolInfos),
		)
		if err != nil {
			log.Printf("生成失败: %v", err)
			continue
		}

		fmt.Printf("AI 的决策:\\n")
		if len(response.ToolCalls) > 0 {
			for _, toolCall := range response.ToolCalls {
				fmt.Printf("  ✓ 使用工具: %s\\n", toolCall.Function.Name)
				fmt.Printf("  参数: %s\\n", toolCall.Function.Arguments)

				// 执行工具
				toolResult, err := toolsNode.Invoke(ctx, response)
				if err != nil {
					log.Printf("  工具执行失败: %v", err)
					continue
				}

				fmt.Printf("  结果: %+v\\n", toolResult)
			}
		} else {
			fmt.Printf("  ✓ 直接回答: %s\\n", response.Content)
		}
	}
}
