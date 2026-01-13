// Package app 提供应用主体
package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cloudwego/eino/schema"

	"github.com/NuyoahCh/einotelos/einox/callback"
	"github.com/NuyoahCh/einotelos/einox/config"
	"github.com/NuyoahCh/einotelos/einox/model"
	"github.com/NuyoahCh/einotelos/einox/prompt"
	"github.com/NuyoahCh/einotelos/einox/rag"
	"github.com/NuyoahCh/einotelos/einox/tools"
	"github.com/NuyoahCh/einotelos/einox/workflow"
)

// App 应用主体
type App struct {
	config    *config.Config
	chatModel *model.ChatModel
	embedder  *model.Embedder
	rag       *rag.SimpleRAG
	tools     []interface {
		Info(context.Context) (*schema.ToolInfo, error)
	}
	history []*schema.Message
}

// New 创建应用实例
func New(ctx context.Context, cfg *config.Config) (*App, error) {
	// 创建对话模型
	chatModel, err := model.NewChatModel(ctx, &cfg.ARK)
	if err != nil {
		return nil, fmt.Errorf("创建对话模型失败: %w", err)
	}

	app := &App{
		config:    cfg,
		chatModel: chatModel,
		history:   make([]*schema.Message, 0),
	}

	// 尝试创建向量模型（可选）
	embedder, err := model.NewEmbedder(ctx, &cfg.ARK)
	if err != nil {
		fmt.Printf("警告: 创建向量模型失败，RAG 功能将使用模拟向量: %v\n", err)
		// 使用模拟向量器
		mockEmbedder := rag.NewMockEmbedder(128)
		app.rag = rag.NewSimpleRAG(mockEmbedder, cfg.RAG.TopK)
	} else {
		app.embedder = embedder
		app.rag = rag.NewSimpleRAG(&embedderAdapter{embedder}, cfg.RAG.TopK)
	}

	return app, nil
}

// embedderAdapter 适配器
type embedderAdapter struct {
	inner *model.Embedder
}

func (a *embedderAdapter) EmbedStrings(ctx context.Context, texts []string) ([][]float64, error) {
	return a.inner.EmbedStrings(ctx, texts)
}

// Close 关闭应用
func (a *App) Close() {
	// 清理资源
}

// DemoBasicChat 演示基础对话
func (a *App) DemoBasicChat(ctx context.Context) error {
	fmt.Println("演示基础对话功能...")

	messages := []*schema.Message{
		schema.SystemMessage("你是一个友好的助手，由字节跳动豆包大模型驱动。"),
		schema.UserMessage("你好！请用一句话介绍一下你自己。"),
	}

	response, err := a.chatModel.Generate(ctx, messages)
	if err != nil {
		return fmt.Errorf("生成响应失败: %w", err)
	}

	fmt.Printf("AI: %s\n", response.Content)

	if response.ResponseMeta != nil && response.ResponseMeta.Usage != nil {
		fmt.Printf("Token 使用: prompt=%d, completion=%d, total=%d\n",
			response.ResponseMeta.Usage.PromptTokens,
			response.ResponseMeta.Usage.CompletionTokens,
			response.ResponseMeta.Usage.TotalTokens)
	}

	return nil
}

// DemoStreamChat 演示流式对话
func (a *App) DemoStreamChat(ctx context.Context) error {
	fmt.Println("演示流式对话功能...")

	messages := []*schema.Message{
		schema.SystemMessage("你是一个知识渊博的助手。"),
		schema.UserMessage("请简要介绍一下 Go 语言的三个主要特点。"),
	}

	stream, err := a.chatModel.Stream(ctx, messages)
	if err != nil {
		return fmt.Errorf("创建流失败: %w", err)
	}
	defer stream.Close()

	fmt.Print("AI: ")
	fullContent, err := callback.CollectStream(stream, &callback.StreamCallback{
		OnChunk: func(content string) {
			fmt.Print(content)
		},
	})
	if err != nil {
		return err
	}
	fmt.Println()
	fmt.Printf("(完整响应长度: %d 字符)\n", len(fullContent))

	return nil
}

// DemoToolCall 演示工具调用
func (a *App) DemoToolCall(ctx context.Context) error {
	fmt.Println("演示工具调用功能...")

	// 获取所有工具
	allTools := tools.GetAllTools()

	// 获取工具信息
	toolInfos, err := tools.GetToolInfos(ctx, allTools)
	if err != nil {
		return fmt.Errorf("获取工具信息失败: %w", err)
	}

	// 绑定工具到模型
	if err := a.chatModel.BindTools(toolInfos); err != nil {
		return fmt.Errorf("绑定工具失败: %w", err)
	}

	// 测试用例
	testCases := []string{
		"现在几点了？",
		"帮我计算 123 * 456",
		"北京今天天气怎么样？",
	}

	for _, query := range testCases {
		fmt.Printf("\n用户: %s\n", query)

		messages := []*schema.Message{
			schema.SystemMessage("你是一个智能助手，可以使用工具来帮助用户。"),
			schema.UserMessage(query),
		}

		response, err := a.chatModel.Generate(ctx, messages)
		if err != nil {
			fmt.Printf("错误: %v\n", err)
			continue
		}

		if len(response.ToolCalls) > 0 {
			fmt.Printf("AI 决定调用工具:\n")
			for _, tc := range response.ToolCalls {
				fmt.Printf("  - %s(%s)\n", tc.Function.Name, tc.Function.Arguments)

				// 执行工具
				for _, t := range allTools {
					info, _ := t.Info(ctx)
					if info.Name == tc.Function.Name {
						if invokable, ok := t.(interface {
							InvokableRun(context.Context, string, ...interface{}) (string, error)
						}); ok {
							result, err := invokable.InvokableRun(ctx, tc.Function.Arguments)
							if err != nil {
								fmt.Printf("  工具执行错误: %v\n", err)
							} else {
								fmt.Printf("  工具结果: %s\n", result)
							}
						}
						break
					}
				}
			}
		} else {
			fmt.Printf("AI: %s\n", response.Content)
		}
	}

	return nil
}

// DemoChain 演示 Chain 编排
func (a *App) DemoChain(ctx context.Context) error {
	fmt.Println("演示 Chain 编排功能...")

	// 构建简单对话链
	builder := workflow.NewChainBuilder(ctx, a.chatModel).
		WithTemplate(prompt.Templates.CodeAssistant)

	chain, err := builder.BuildSimpleChain()
	if err != nil {
		return fmt.Errorf("构建 Chain 失败: %w", err)
	}

	// 运行链
	response, err := workflow.RunChain(ctx, chain, "请用 Go 写一个简单的 Hello World 程序", nil)
	if err != nil {
		return fmt.Errorf("运行 Chain 失败: %w", err)
	}

	fmt.Printf("AI: %s\n", response.Content)
	return nil
}

// DemoGraph 演示 Graph 工作流
func (a *App) DemoGraph(ctx context.Context) error {
	fmt.Println("演示 Graph 工作流功能...")

	// 构建工具调用图
	allTools := tools.GetAllTools()
	toolList := make([]interface {
		Info(context.Context) (*schema.ToolInfo, error)
		InvokableRun(context.Context, string, ...interface{}) (string, error)
	}, 0)

	// 类型转换
	for _, t := range allTools {
		if bt, ok := t.(interface {
			Info(context.Context) (*schema.ToolInfo, error)
			InvokableRun(context.Context, string, ...interface{}) (string, error)
		}); ok {
			toolList = append(toolList, bt)
		}
	}

	builder := workflow.NewGraphBuilder(ctx, a.chatModel).
		WithTemplate(prompt.Templates.ToolAssistant).
		WithTools(allTools)

	graph, err := builder.BuildToolGraph()
	if err != nil {
		return fmt.Errorf("构建 Graph 失败: %w", err)
	}

	// 运行图
	response, err := workflow.RunGraph(ctx, graph, map[string]any{
		"query":   "帮我查一下北京的天气，然后计算一下 25 + 30 等于多少",
		"history": []*schema.Message{},
	})
	if err != nil {
		return fmt.Errorf("运行 Graph 失败: %w", err)
	}

	fmt.Printf("AI: %s\n", response.Content)
	return nil
}

// DemoRAG 演示 RAG 检索增强
func (a *App) DemoRAG(ctx context.Context) error {
	fmt.Println("演示 RAG 检索增强功能...")

	// 添加知识库内容
	knowledgeBase := []string{
		`Eino 是字节跳动开源的 Go 语言 LLM 应用开发框架。它提供了丰富的组件，包括：
		1. ChatModel - 对话模型组件，支持多种 LLM 服务
		2. Prompt - 提示词模板组件，支持变量替换和消息占位符
		3. Tool - 工具组件，让 LLM 能够调用外部功能
		4. Retriever - 检索组件，用于 RAG 应用
		5. Indexer - 索引组件，用于文档向量化存储`,

		`Eino 的工作流编排支持三种方式：
		1. Chain - 链式调用，适合简单的顺序流程
		2. Graph - 图式编排，支持复杂的分支和并行
		3. Workflow - 工作流编排，提供更灵活的节点管理`,

		`使用 Eino 开发 RAG 应用的步骤：
		1. 加载文档 - 使用 Loader 组件加载各种格式的文档
		2. 分割文档 - 使用 Splitter 将长文档切分成小块
		3. 向量化 - 使用 Embedding 组件将文本转换为向量
		4. 存储索引 - 使用 Indexer 将向量存入向量数据库
		5. 检索查询 - 使用 Retriever 根据用户问题检索相关文档
		6. 生成回答 - 将检索结果作为上下文，让 LLM 生成回答`,
	}

	// 索引知识库
	fmt.Println("正在索引知识库...")
	for i, text := range knowledgeBase {
		if err := a.rag.AddText(ctx, text, map[string]any{"index": i}); err != nil {
			return fmt.Errorf("索引文档失败: %w", err)
		}
	}
	fmt.Printf("已索引 %d 个文档\n", len(knowledgeBase))

	// 测试查询
	queries := []string{
		"Eino 框架有哪些主要组件？",
		"如何使用 Eino 开发 RAG 应用？",
	}

	for _, query := range queries {
		fmt.Printf("\n用户: %s\n", query)

		// 检索相关文档
		context, docs, err := a.rag.Query(ctx, query)
		if err != nil {
			fmt.Printf("检索失败: %v\n", err)
			continue
		}

		fmt.Printf("检索到 %d 个相关文档\n", len(docs))

		// 使用 RAG 模板生成回答
		messages, err := prompt.FormatRAGMessages(ctx, query, context, nil)
		if err != nil {
			fmt.Printf("格式化消息失败: %v\n", err)
			continue
		}

		response, err := a.chatModel.Generate(ctx, messages)
		if err != nil {
			fmt.Printf("生成回答失败: %v\n", err)
			continue
		}

		fmt.Printf("AI: %s\n", response.Content)
	}

	return nil
}

// RunInteractive 运行交互式命令行
func (a *App) RunInteractive(ctx context.Context) {
	fmt.Println("\n========================================")
	fmt.Println("  进入交互模式 (输入 'quit' 退出)")
	fmt.Println("========================================")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("\n你: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if input == "quit" || input == "exit" {
			fmt.Println("再见！")
			break
		}

		// 构建消息
		messages := append([]*schema.Message{
			schema.SystemMessage("你是一个友好的助手，由字节跳动豆包大模型驱动。"),
		}, a.history...)
		messages = append(messages, schema.UserMessage(input))

		// 生成响应
		response, err := a.chatModel.Generate(ctx, messages)
		if err != nil {
			fmt.Printf("错误: %v\n", err)
			continue
		}

		fmt.Printf("AI: %s\n", response.Content)

		// 更新历史
		a.history = append(a.history, schema.UserMessage(input))
		a.history = append(a.history, response)

		// 限制历史长度
		if len(a.history) > 20 {
			a.history = a.history[len(a.history)-20:]
		}
	}
}
