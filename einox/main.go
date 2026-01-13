// Package main 是 Eino 框架全方位应用示例
// 整合了 ChatModel、Prompt、Tool、Callback、Lambda、Workflow、RAG 等核心组件
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/NuyoahCh/einotelos/einox/app"
	"github.com/NuyoahCh/einotelos/einox/config"
)

func main() {
	ctx := context.Background()

	// 1. 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 2. 创建应用实例
	application, err := app.New(ctx, cfg)
	if err != nil {
		log.Fatalf("创建应用失败: %v", err)
	}
	defer application.Close()

	// 3. 运行交互式演示
	fmt.Println("========================================")
	fmt.Println("  Eino 框架全方位应用示例 - 豆包版")
	fmt.Println("========================================")
	fmt.Println()

	// 演示各个功能模块
	demos := []struct {
		name string
		fn   func(context.Context) error
	}{
		{"基础对话", application.DemoBasicChat},
		{"流式对话", application.DemoStreamChat},
		{"工具调用", application.DemoToolCall},
		{"Chain 编排", application.DemoChain},
		{"Graph 工作流", application.DemoGraph},
		{"RAG 检索增强", application.DemoRAG},
	}

	for i, demo := range demos {
		fmt.Printf("\n[%d/%d] %s\n", i+1, len(demos), demo.name)
		fmt.Println("----------------------------------------")

		if err := demo.fn(ctx); err != nil {
			log.Printf("演示 %s 失败: %v", demo.name, err)
			continue
		}
	}

	fmt.Println("\n========================================")
	fmt.Println("  所有演示完成！")
	fmt.Println("========================================")

	// 4. 可选：启动交互式命令行
	if len(os.Args) > 1 && os.Args[1] == "-i" {
		application.RunInteractive(ctx)
	}
}
