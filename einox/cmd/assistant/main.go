// AI 学习助手 - Eino 框架教学示例
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/NuyoahCh/einotelos/einox/config"
	"github.com/NuyoahCh/einotelos/einox/internal/assistant"
)

func main() {
	ctx := context.Background()

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("配置错误: %v\n", err)
		fmt.Println("请确保已设置环境变量: ARK_API_KEY, ARK_MODEL_NAME")
		os.Exit(1)
	}

	// 创建助手
	ast, err := assistant.New(ctx, cfg)
	if err != nil {
		fmt.Printf("初始化失败: %v\n", err)
		os.Exit(1)
	}
	defer ast.Close()

	// 显示欢迎界面
	printWelcome()

	// 主循环
	scanner := bufio.NewScanner(os.Stdin)
	for {
		printMenu()
		fmt.Print("请选择 [1-6]: ")

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
		case "6", "q", "quit", "exit":
			fmt.Println("👋 再见，祝学习愉快！")
			return
		default:
			fmt.Println("❌ 无效选择，请重试")
		}
	}
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
	fmt.Println("│  1. 💬 智能对话    - 自由问答              │")
	fmt.Println("│  2. 📚 知识库问答  - RAG 检索增强          │")
	fmt.Println("│  3. 🔧 学习工具    - 计算/天气/时间        │")
	fmt.Println("│  4. 💻 代码助手    - 代码解释与生成        │")
	fmt.Println("│  5. 🌐 翻译助手    - 中英互译              │")
	fmt.Println("│  6. 🚪 退出                                │")
	fmt.Println("└────────────────────────────────────────────┘")
}

func runChat(ctx context.Context, ast *assistant.Assistant, scanner *bufio.Scanner) {
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
		if err := ast.ChatStream(ctx, input); err != nil {
			fmt.Printf("\n错误: %v\n", err)
		}
		fmt.Println()
	}
}

func runKnowledgeQA(ctx context.Context, ast *assistant.Assistant, scanner *bufio.Scanner) {
	fmt.Println("📚 知识库问答模式 (输入 'back' 返回菜单)")
	fmt.Println("────────────────────────────────────────────")
	fmt.Println()

	// 检查知识库是否为空
	if ast.KnowledgeCount() == 0 {
		fmt.Println("📝 知识库为空，请先添加知识")
		fmt.Println()
		fmt.Println("示例知识（直接回车使用默认）:")
		fmt.Print("> ")

		if scanner.Scan() {
			input := strings.TrimSpace(scanner.Text())
			if input == "" {
				// 使用默认知识
				defaultKnowledge := []string{
					`Eino 是字节跳动开源的 Go 语言 LLM 应用开发框架。主要特点：
1. 组件化设计：ChatModel、Prompt、Tool、Retriever 等组件可灵活组合
2. 工作流编排：支持 Chain（链式）和 Graph（图式）两种编排方式
3. 流式支持：原生支持流式输出，提升用户体验
4. 回调机制：支持日志、监控、追踪等回调处理`,

					`RAG（检索增强生成）是一种结合检索和生成的技术：
1. 文档加载：支持本地文件、URL、S3 等多种来源
2. 文档分割：将长文档切分成小块，便于检索
3. 向量化：使用 Embedding 模型将文本转换为向量
4. 检索：根据用户问题检索相关文档片段
5. 生成：将检索结果作为上下文，让 LLM 生成回答`,

					`Go 语言（Golang）是 Google 开发的编程语言，特点：
1. 简洁：语法简单，易于学习
2. 高效：编译速度快，运行性能好
3. 并发：goroutine 和 channel 原生支持并发
4. 工具链：内置格式化、测试、文档等工具`,
				}
				for _, k := range defaultKnowledge {
					ast.AddKnowledge(ctx, k)
				}
				fmt.Printf("✅ 已添加 %d 条默认知识\n", len(defaultKnowledge))
			} else if input != "back" {
				ast.AddKnowledge(ctx, input)
				fmt.Println("✅ 知识已添加")
			}
		}
	} else {
		fmt.Printf("📖 当前知识库: %d 条记录\n", ast.KnowledgeCount())
	}

	fmt.Println()
	fmt.Println("命令: 'add' 添加知识, 'clear' 清空知识库, 'back' 返回")
	fmt.Println()

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
			fmt.Print("输入知识内容: ")
			if scanner.Scan() {
				knowledge := strings.TrimSpace(scanner.Text())
				if knowledge != "" {
					ast.AddKnowledge(ctx, knowledge)
					fmt.Println("✅ 知识已添加")
				}
			}
			continue
		}
		if input == "clear" {
			ast.ClearKnowledge(ctx)
			fmt.Println("✅ 知识库已清空")
			continue
		}

		fmt.Print("回答: ")
		if err := ast.KnowledgeQA(ctx, input); err != nil {
			fmt.Printf("\n错误: %v\n", err)
		}
		fmt.Println()
	}
}

func runTools(ctx context.Context, ast *assistant.Assistant, scanner *bufio.Scanner) {
	fmt.Println("🔧 学习工具模式 (输入 'back' 返回菜单)")
	fmt.Println("────────────────────────────────────────────")
	fmt.Println()
	fmt.Println("可用工具:")
	fmt.Println("  • 计算器 - 例: 计算 123 * 456")
	fmt.Println("  • 天气   - 例: 北京天气怎么样")
	fmt.Println("  • 时间   - 例: 现在几点了")
	fmt.Println()

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

		result, err := ast.UseTool(ctx, input)
		if err != nil {
			fmt.Printf("错误: %v\n", err)
		} else {
			fmt.Printf("结果: %s\n", result)
		}
		fmt.Println()
	}
}

func runCodeAssistant(ctx context.Context, ast *assistant.Assistant, scanner *bufio.Scanner) {
	fmt.Println("💻 代码助手模式 (输入 'back' 返回菜单)")
	fmt.Println("────────────────────────────────────────────")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  • 用 Go 写一个快速排序")
	fmt.Println("  • 解释这段代码: for i := range arr { ... }")
	fmt.Println("  • Python 和 Go 的区别是什么")
	fmt.Println()

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

		fmt.Println()
		fmt.Print("AI: ")
		if err := ast.CodeAssist(ctx, input); err != nil {
			fmt.Printf("\n错误: %v\n", err)
		}
		fmt.Println()
	}
}

func runTranslator(ctx context.Context, ast *assistant.Assistant, scanner *bufio.Scanner) {
	fmt.Println("🌐 翻译助手模式 (输入 'back' 返回菜单)")
	fmt.Println("────────────────────────────────────────────")
	fmt.Println()
	fmt.Println("自动识别语言，中文翻译成英文，英文翻译成中文")
	fmt.Println()

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
		if err := ast.Translate(ctx, input); err != nil {
			fmt.Printf("\n错误: %v\n", err)
		}
		fmt.Println()
	}
}
