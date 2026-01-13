// AI 学习助手 - Eino 框架教学示例
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"github.com/NuyoahCh/einotelos/einox/callback"
	"github.com/NuyoahCh/einotelos/einox/config"
	"github.com/NuyoahCh/einotelos/einox/model"
	"github.com/NuyoahCh/einotelos/einox/rag"
	"github.com/NuyoahCh/einotelos/einox/tools"
	"github.com/NuyoahCh/einotelos/einox/workflow"
)

// Assistant AI 学习助手
type Assistant struct {
	config    *config.Config
	chatModel *model.ChatModel
	rag       *rag.SimpleRAG
	history   []*schema.Message
}

func main() {
	ctx := context.Background()

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("❌ 配置错误: %v\n", err)
		fmt.Println("请确保已设置环境变量: ARK_API_KEY, ARK_MODEL_NAME")
		os.Exit(1)
	}

	// 创建助手
	ast, err := newAssistant(ctx, cfg)
	if err != nil {
		fmt.Printf("❌ 初始化失败: %v\n", err)
		os.Exit(1)
	}

	// 显示欢迎界面
	printWelcome()

	// 主循环
	scanner := bufio.NewScanner(os.Stdin)
	for {
		printMenu()
		fmt.Print("请选择 [1-7]: ")

		if !scanner.Scan() {
			break
		}

		choice := strings.TrimSpace(scanner.Text())
		fmt.Println()

		switch choice {
		case "1":
			runChat(ctx, ast, scanner)
		case "2":
			runKnowledgeQA(ctx, ast, scanner)
		case "3":
			runTools(ctx, ast, scanner)
		case "4":
			runCodeAssistant(ctx, ast, scanner)
		case "5":
			runTranslator(ctx, ast, scanner)
		case "6":
			runWorkflowDemo(ctx, ast)
		case "7", "q", "quit", "exit":
			fmt.Println("👋 再见，祝学习愉快！")
			return
		default:
			fmt.Println("❌ 无效选择，请重试")
		}
	}
}

func newAssistant(ctx context.Context, cfg *config.Config) (*Assistant, error) {
	chatModel, err := model.NewChatModel(ctx, &cfg.ARK)
	if err != nil {
		return nil, fmt.Errorf("创建模型失败: %w", err)
	}

	mockEmbedder := rag.NewMockEmbedder(128)
	simpleRAG := rag.NewSimpleRAG(mockEmbedder, cfg.RAG.TopK)

	return &Assistant{
		config:    cfg,
		chatModel: chatModel,
		rag:       simpleRAG,
		history:   make([]*schema.Message, 0),
	}, nil
}

func printWelcome() {
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════╗")
	fmt.Println("║                                            ║")
	fmt.Println("║      🎓 AI 学习助手 - Eino 教学示例        ║")
	fmt.Println("║                                            ║")
	fmt.Println("║   基于字节跳动 Eino 框架 + 豆包大模型      ║")
	fmt.Println("║                                            ║")
	fmt.Println("╚════════════════════════════════════════════╝")
	fmt.Println()
}

func printMenu() {
	fmt.Println()
	fmt.Println("┌────────────────────────────────────────────┐")
	fmt.Println("│  功能菜单                                  │")
	fmt.Println("├────────────────────────────────────────────┤")
	fmt.Println("│  1. 💬 智能对话    - 流式问答              │")
	fmt.Println("│  2. 📚 知识库问答  - RAG 检索增强          │")
	fmt.Println("│  3. 🔧 学习工具    - 计算/天气/时间        │")
	fmt.Println("│  4. 💻 代码助手    - 代码解释与生成        │")
	fmt.Println("│  5. 🌐 翻译助手    - 中英互译              │")
	fmt.Println("│  6. ⚙️  工作流演示  - Chain/Graph 编排     │")
	fmt.Println("│  7. 🚪 退出                                │")
	fmt.Println("└────────────────────────────────────────────┘")
}

// ==================== 1. 智能对话 ====================

func runChat(ctx context.Context, ast *Assistant, scanner *bufio.Scanner) {
	fmt.Println("💬 智能对话模式 (输入 'back' 返回菜单)")
	fmt.Println("────────────────────────────────────────────")

	for {
		fmt.Print("\n你: ")
		if !scanner.Scan() {
			return
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if input == "back" {
			return
		}

		fmt.Print("AI: ")
		if err := ast.chatStream(ctx, input); err != nil {
			fmt.Printf("\n错误: %v\n", err)
		}
		fmt.Println()
	}
}

func (a *Assistant) chatStream(ctx context.Context, query string) error {
	messages := []*schema.Message{
		schema.SystemMessage("你是一个友好的 AI 学习助手，帮助用户解答各种问题。回答要简洁清晰。"),
	}
	messages = append(messages, a.history...)
	messages = append(messages, schema.UserMessage(query))

	stream, err := a.chatModel.Stream(ctx, messages)
	if err != nil {
		return err
	}
	defer stream.Close()

	fullContent, _ := callback.CollectStream(stream, &callback.StreamCallback{
		OnChunk: func(content string) {
			fmt.Print(content)
		},
	})

	a.addHistory(query, fullContent)
	return nil
}

func (a *Assistant) addHistory(query, response string) {
	a.history = append(a.history, schema.UserMessage(query))
	a.history = append(a.history, schema.AssistantMessage(response, nil))
	if len(a.history) > 10 {
		a.history = a.history[len(a.history)-10:]
	}
}

// ==================== 2. 知识库问答 ====================

func runKnowledgeQA(ctx context.Context, ast *Assistant, scanner *bufio.Scanner) {
	fmt.Println("📚 知识库问答模式 (输入 'back' 返回菜单)")
	fmt.Println("────────────────────────────────────────────")
	fmt.Println()

	if ast.knowledgeCount() == 0 {
		fmt.Println("📝 知识库为空，直接回车添加默认知识:")
		fmt.Print("> ")

		if scanner.Scan() {
			input := strings.TrimSpace(scanner.Text())
			if input == "" {
				ast.addDefaultKnowledge(ctx)
			} else if input != "back" {
				ast.rag.AddText(ctx, input, nil)
				fmt.Println("✅ 知识已添加")
			}
		}
	} else {
		fmt.Printf("📖 当前知识库: %d 条记录\n", ast.knowledgeCount())
	}

	fmt.Println("\n命令: 'add' 添加, 'clear' 清空, 'back' 返回\n")

	for {
		fmt.Print("问题: ")
		if !scanner.Scan() {
			return
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if input == "back" {
			return
		}
		if input == "add" {
			fmt.Print("输入知识: ")
			if scanner.Scan() {
				if k := strings.TrimSpace(scanner.Text()); k != "" {
					ast.rag.AddText(ctx, k, nil)
					fmt.Println("✅ 已添加")
				}
			}
			continue
		}
		if input == "clear" {
			ast.rag = rag.NewSimpleRAG(rag.NewMockEmbedder(128), ast.config.RAG.TopK)
			fmt.Println("✅ 已清空")
			continue
		}

		fmt.Print("回答: ")
		ast.knowledgeQA(ctx, input)
		fmt.Println()
	}
}

func (a *Assistant) knowledgeCount() int {
	if store, ok := a.rag.GetStore().(*rag.MemoryVectorStore); ok {
		return store.Count()
	}
	return 0
}

func (a *Assistant) addDefaultKnowledge(ctx context.Context) {
	knowledge := []string{
		`Eino 是字节跳动开源的 Go 语言 LLM 应用开发框架。主要特点：
1. 组件化设计：ChatModel、Prompt、Tool、Retriever 等组件可灵活组合
2. 工作流编排：支持 Chain（链式）和 Graph（图式）两种编排方式
3. 流式支持：原生支持流式输出，提升用户体验`,

		`RAG（检索增强生成）是一种结合检索和生成的技术：
1. 文档分割：将长文档切分成小块
2. 向量化：使用 Embedding 模型将文本转换为向量
3. 检索：根据用户问题检索相关文档片段
4. 生成：将检索结果作为上下文，让 LLM 生成回答`,

		`Go 语言是 Google 开发的编程语言，特点：简洁、高效、原生支持并发（goroutine）`,
	}
	for _, k := range knowledge {
		a.rag.AddText(ctx, k, nil)
	}
	fmt.Printf("✅ 已添加 %d 条默认知识\n", len(knowledge))
}

func (a *Assistant) knowledgeQA(ctx context.Context, query string) {
	ragContext, docs, err := a.rag.Query(ctx, query)
	if err != nil {
		fmt.Printf("检索错误: %v", err)
		return
	}

	messages := []*schema.Message{
		schema.SystemMessage(fmt.Sprintf(`基于以下资料回答问题，资料不足请说明：

%s`, ragContext)),
		schema.UserMessage(query),
	}

	stream, err := a.chatModel.Stream(ctx, messages)
	if err != nil {
		fmt.Printf("错误: %v", err)
		return
	}
	defer stream.Close()

	callback.CollectStream(stream, &callback.StreamCallback{
		OnChunk: func(c string) { fmt.Print(c) },
	})

	if len(docs) > 0 {
		fmt.Printf("\n\n📎 参考了 %d 条知识", len(docs))
	}
}

// ==================== 3. 学习工具 ====================

func runTools(ctx context.Context, ast *Assistant, scanner *bufio.Scanner) {
	fmt.Println("🔧 学习工具模式 (输入 'back' 返回菜单)")
	fmt.Println("────────────────────────────────────────────")
	fmt.Println("\n可用工具:")
	fmt.Println("  • 计算器 - 例: 计算 123 * 456")
	fmt.Println("  • 天气   - 例: 北京天气怎么样")
	fmt.Println("  • 时间   - 例: 现在几点了\n")

	allTools := tools.GetAllTools()
	toolInfos, _ := tools.GetToolInfos(ctx, allTools)
	ast.chatModel.BindTools(toolInfos)

	for {
		fmt.Print("请求: ")
		if !scanner.Scan() {
			return
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if input == "back" {
			return
		}

		messages := []*schema.Message{
			schema.SystemMessage("根据用户需求使用工具：calculator(计算)、get_weather(天气)、get_current_time(时间)"),
			schema.UserMessage(input),
		}

		response, err := ast.chatModel.Generate(ctx, messages)
		if err != nil {
			fmt.Printf("错误: %v\n\n", err)
			continue
		}

		if len(response.ToolCalls) > 0 {
			for _, tc := range response.ToolCalls {
				result := executeTool(ctx, allTools, tc.Function.Name, tc.Function.Arguments)
				fmt.Printf("结果: [%s] %s\n\n", tc.Function.Name, result)
			}
		} else {
			fmt.Printf("结果: %s\n\n", response.Content)
		}
	}
}

func executeTool(ctx context.Context, allTools []tool.BaseTool, name, args string) string {
	for _, t := range allTools {
		info, _ := t.Info(ctx)
		if info.Name == name {
			if invokable, ok := t.(interface {
				InvokableRun(context.Context, string, ...tool.Option) (string, error)
			}); ok {
				result, err := invokable.InvokableRun(ctx, args)
				if err != nil {
					return fmt.Sprintf("错误: %v", err)
				}
				return result
			}
		}
	}
	return "工具未找到"
}

// ==================== 4. 代码助手 ====================

func runCodeAssistant(ctx context.Context, ast *Assistant, scanner *bufio.Scanner) {
	fmt.Println("💻 代码助手模式 (输入 'back' 返回菜单)")
	fmt.Println("────────────────────────────────────────────")
	fmt.Println("\n示例: 用Go写快速排序 / 解释for range / Go和Python区别\n")

	for {
		fmt.Print("问题: ")
		if !scanner.Scan() {
			return
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if input == "back" {
			return
		}

		fmt.Print("\nAI: ")
		messages := []*schema.Message{
			schema.SystemMessage("你是编程助手，代码用markdown代码块，解释简洁。"),
			schema.UserMessage(input),
		}

		stream, err := ast.chatModel.Stream(ctx, messages)
		if err != nil {
			fmt.Printf("错误: %v\n", err)
			continue
		}
		callback.CollectStream(stream, &callback.StreamCallback{
			OnChunk: func(c string) { fmt.Print(c) },
		})
		stream.Close()
		fmt.Println("\n")
	}
}

// ==================== 5. 翻译助手 ====================

func runTranslator(ctx context.Context, ast *Assistant, scanner *bufio.Scanner) {
	fmt.Println("🌐 翻译助手模式 (输入 'back' 返回菜单)")
	fmt.Println("────────────────────────────────────────────")
	fmt.Println("\n自动识别：中文→英文，英文→中文\n")

	for {
		fmt.Print("原文: ")
		if !scanner.Scan() {
			return
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if input == "back" {
			return
		}

		fmt.Print("译文: ")
		messages := []*schema.Message{
			schema.SystemMessage("翻译助手：中文翻英文，英文翻中文，只输出译文。"),
			schema.UserMessage(input),
		}

		stream, err := ast.chatModel.Stream(ctx, messages)
		if err != nil {
			fmt.Printf("错误: %v\n", err)
			continue
		}
		callback.CollectStream(stream, &callback.StreamCallback{
			OnChunk: func(c string) { fmt.Print(c) },
		})
		stream.Close()
		fmt.Println("\n")
	}
}

// ==================== 6. 工作流演示 ====================

func runWorkflowDemo(ctx context.Context, ast *Assistant) {
	fmt.Println("⚙️  工作流演示 - 展示 Eino Chain/Graph 编排能力")
	fmt.Println("────────────────────────────────────────────")

	// Chain 演示
	fmt.Println("\n📌 Chain 编排演示:")
	fmt.Println("   构建: Template → ChatModel")

	builder := workflow.NewChainBuilder(ctx, ast.chatModel)
	chain, err := builder.BuildSimpleChain()
	if err != nil {
		fmt.Printf("   ❌ 构建失败: %v\n", err)
	} else {
		fmt.Println("   ✅ Chain 构建成功")

		result, err := workflow.RunChain(ctx, chain, "用一句话介绍Go语言", nil)
		if err != nil {
			fmt.Printf("   ❌ 运行失败: %v\n", err)
		} else {
			fmt.Printf("   💬 结果: %s\n", truncate(result.Content, 100))
		}
	}

	// Graph 演示
	fmt.Println("\n📌 Graph 编排演示:")
	fmt.Println("   构建: START → Chat → Tools → Extract → FinalChat → END")

	allTools := tools.GetAllTools()
	graphBuilder := workflow.NewGraphBuilder(ctx, ast.chatModel).WithTools(allTools)
	graph, err := graphBuilder.BuildToolGraph()
	if err != nil {
		fmt.Printf("   ❌ 构建失败: %v\n", err)
	} else {
		fmt.Println("   ✅ Graph 构建成功")

		result, err := workflow.RunGraph(ctx, graph, map[string]any{
			"query":   "现在几点了",
			"history": []*schema.Message{},
		})
		if err != nil {
			fmt.Printf("   ❌ 运行失败: %v\n", err)
		} else {
			fmt.Printf("   💬 结果: %s\n", truncate(result.Content, 100))
		}
	}

	fmt.Println("\n✅ 工作流演示完成")
	fmt.Println()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
