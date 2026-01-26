package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// ReActAgent 实现 ReAct（Reasoning + Acting）模式的智能体
// ReAct 是一种让 LLM 交替进行推理（Reasoning）和行动（Acting）的方法
// 通过这种方式，Agent 可以更好地解决复杂问题
func main() {
	ctx := context.Background()

	// 创建 ChatModel 实例
	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  os.Getenv("DEEPSEEK_API_KEY"),
		Model:   "deepseek-chat",
		BaseURL: "https://api.deepseek.com",
	})
	if err != nil {
		log.Fatalf("创建 ChatModel 实例失败: %v", err)
	}

	// 创建 ReAct Agent
	agent := NewReActAgent(chatModel)

	// 示例 1：数学问题求解
	fmt.Println("=== 示例 1: 数学问题求解 ===")
	question1 := "如果一个商店有 23 个苹果，卖出了 17 个，又进货 45 个，现在有多少个苹果？"
	answer1, err := agent.Solve(ctx, question1, 5)
	if err != nil {
		log.Printf("求解失败: %v", err)
	} else {
		fmt.Printf("\n最终答案: %s\n", answer1)
	}

	// 示例 2：信息查询和推理
	fmt.Println("\n=== 示例 2: 信息查询和推理 ===")
	question2 := "Python 是什么时候发布的？它的创始人是谁？这个人现在多大年纪？"
	answer2, err := agent.Solve(ctx, question2, 5)
	if err != nil {
		log.Printf("求解失败: %v", err)
	} else {
		fmt.Printf("\n最终答案: %s\n", answer2)
	}

	// 示例 3：复杂的多步骤任务
	fmt.Println("\n=== 示例 3: 复杂的多步骤任务 ===")
	question3 := "帮我规划一个周末去北京旅游的行程，包括景点推荐、交通方式和预算估算。"
	answer3, err := agent.Solve(ctx, question3, 8)
	if err != nil {
		log.Printf("求解失败: %v", err)
	} else {
		fmt.Printf("\n最终答案: %s\n", answer3)
	}
}

// ReActAgent ReAct 智能体结构
type ReActAgent struct {
	chatModel model.ChatModel
	tools     map[string]Tool
}

// Tool 工具接口
type Tool interface {
	Name() string
	Description() string
	Execute(ctx context.Context, input string) (string, error)
}

// NewReActAgent 创建新的 ReAct Agent
func NewReActAgent(chatModel model.ChatModel) *ReActAgent {
	agent := &ReActAgent{
		chatModel: chatModel,
		tools:     make(map[string]Tool),
	}

	// 注册可用的工具
	agent.RegisterTool(&CalculatorTool{})
	agent.RegisterTool(&SearchTool{})
	agent.RegisterTool(&KnowledgeTool{})

	return agent
}

// RegisterTool 注册工具
func (a *ReActAgent) RegisterTool(tool Tool) {
	a.tools[tool.Name()] = tool
}

// Solve 使用 ReAct 模式解决问题
func (a *ReActAgent) Solve(ctx context.Context, question string, maxSteps int) (string, error) {
	// 构建系统提示词
	systemPrompt := a.buildSystemPrompt()

	// 初始化对话历史
	messages := []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(fmt.Sprintf("问题: %s", question)),
	}

	// ReAct 循环
	for step := 1; step <= maxSteps; step++ {
		fmt.Printf("\n--- 步骤 %d ---\n", step)

		// 让模型思考下一步行动
		response, err := a.chatModel.Generate(ctx, messages)
		if err != nil {
			return "", fmt.Errorf("生成响应失败: %w", err)
		}

		fmt.Printf("Agent 思考:\n%s\n", response.Content)

		// 解析响应
		action, actionInput, observation := a.parseResponse(response.Content)

		// 如果得到最终答案，返回
		if action == "Final Answer" {
			return actionInput, nil
		}

		// 执行工具调用
		if tool, exists := a.tools[action]; exists {
			result, err := tool.Execute(ctx, actionInput)
			if err != nil {
				observation = fmt.Sprintf("工具执行失败: %v", err)
			} else {
				observation = result
			}
			fmt.Printf("\n观察结果: %s\n", observation)
		} else if action != "" {
			observation = fmt.Sprintf("未知的工具: %s", action)
		}

		// 将观察结果添加到对话历史
		messages = append(messages, response)
		messages = append(messages, schema.UserMessage(fmt.Sprintf("Observation: %s", observation)))
	}

	return "达到最大步骤限制，未能找到答案", nil
}

// buildSystemPrompt 构建系统提示词
func (a *ReActAgent) buildSystemPrompt() string {
	toolDescriptions := ""
	for _, tool := range a.tools {
		toolDescriptions += fmt.Sprintf("- %s: %s\n", tool.Name(), tool.Description())
	}

	return fmt.Sprintf(`你是一个使用 ReAct（Reasoning + Acting）模式的智能助手。

你可以使用以下工具：
%s

请按照以下格式回答问题：

Thought: 我需要思考如何解决这个问题
Action: [工具名称]
Action Input: [工具输入]
Observation: [工具返回的结果]
... (重复 Thought/Action/Action Input/Observation 多次)
Thought: 我现在知道最终答案了
Final Answer: [最终答案]

重要规则：
1. 每次只执行一个 Action
2. 必须等待 Observation 后再继续
3. 充分利用工具来获取信息
4. 当你确定答案时，使用 "Final Answer" 给出结论

现在开始！`, toolDescriptions)
}

// parseResponse 解析模型响应，提取 Action 和 Action Input
func (a *ReActAgent) parseResponse(response string) (action, actionInput, observation string) {
	lines := strings.Split(response, "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "Action:") {
			action = strings.TrimSpace(strings.TrimPrefix(line, "Action:"))
		} else if strings.HasPrefix(line, "Action Input:") {
			actionInput = strings.TrimSpace(strings.TrimPrefix(line, "Action Input:"))
		} else if strings.HasPrefix(line, "Final Answer:") {
			action = "Final Answer"
			actionInput = strings.TrimSpace(strings.TrimPrefix(line, "Final Answer:"))
			// 收集后续所有行作为答案的一部分
			if i+1 < len(lines) {
				actionInput += "\n" + strings.Join(lines[i+1:], "\n")
			}
			break
		}
	}

	return action, actionInput, observation
}

// CalculatorTool 计算器工具
type CalculatorTool struct{}

func (t *CalculatorTool) Name() string {
	return "Calculator"
}

func (t *CalculatorTool) Description() string {
	return "用于执行数学计算。输入应该是一个数学表达式，例如 '23 - 17 + 45'"
}

func (t *CalculatorTool) Execute(ctx context.Context, input string) (string, error) {
	// 简化实现：这里应该使用真正的表达式求值器
	// 为了演示，我们只处理简单的加减法
	input = strings.ReplaceAll(input, " ", "")

	// 这里应该实现真正的计算逻辑
	// 为了简化，返回一个示例结果
	return fmt.Sprintf("计算结果: %s = [计算结果]", input), nil
}

// SearchTool 搜索工具（模拟）
type SearchTool struct{}

func (t *SearchTool) Name() string {
	return "Search"
}

func (t *SearchTool) Description() string {
	return "用于搜索互联网信息。输入应该是搜索查询"
}

func (t *SearchTool) Execute(ctx context.Context, input string) (string, error) {
	// 模拟搜索结果
	// 实际应用中应该调用真实的搜索 API
	return fmt.Sprintf("搜索 '%s' 的结果: [这里是搜索结果的摘要]", input), nil
}

// KnowledgeTool 知识库工具
type KnowledgeTool struct{}

func (t *KnowledgeTool) Name() string {
	return "Knowledge"
}

func (t *KnowledgeTool) Description() string {
	return "用于查询内部知识库。输入应该是要查询的主题或问题"
}

func (t *KnowledgeTool) Execute(ctx context.Context, input string) (string, error) {
	// 模拟知识库查询
	// 实际应用中应该查询向量数据库或知识图谱
	knowledgeBase := map[string]string{
		"Python": "Python 是由 Guido van Rossum 在 1991 年创建的编程语言。",
		"北京":     "北京是中国的首都，有许多著名景点如故宫、长城、天安门等。",
	}

	for key, value := range knowledgeBase {
		if strings.Contains(input, key) {
			return value, nil
		}
	}

	return "知识库中未找到相关信息", nil
}
