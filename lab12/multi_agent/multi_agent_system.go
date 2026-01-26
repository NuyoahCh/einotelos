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

// MultiAgentSystem 演示多 Agent 协作系统
// 多个专门化的 Agent 协同工作，完成复杂任务
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

	// 创建多 Agent 系统
	system := NewMultiAgentSystem(chatModel)

	// 示例 1：软件开发任务
	fmt.Println("=== 示例 1: 软件开发任务 ===")
	task1 := "开发一个简单的待办事项管理应用"
	result1, err := system.ExecuteTask(ctx, task1)
	if err != nil {
		log.Printf("任务执行失败: %v", err)
	} else {
		fmt.Printf("\n任务完成！\n结果:\n%s\n", result1)
	}

	// 示例 2：内容创作任务
	fmt.Println("\n=== 示例 2: 内容创作任务 ===")
	task2 := "为一家咖啡店写一篇营销文案"
	result2, err := system.ExecuteTask(ctx, task2)
	if err != nil {
		log.Printf("任务执行失败: %v", err)
	} else {
		fmt.Printf("\n任务完成！\n结果:\n%s\n", result2)
	}

	// 示例 3：研究分析任务
	fmt.Println("\n=== 示例 3: 研究分析任务 ===")
	task3 := "分析人工智能在医疗领域的应用前景"
	result3, err := system.ExecuteTask(ctx, task3)
	if err != nil {
		log.Printf("任务执行失败: %v", err)
	} else {
		fmt.Printf("\n任务完成！\n结果:\n%s\n", result3)
	}
}

// Agent 智能体接口
type Agent interface {
	Name() string
	Role() string
	Process(ctx context.Context, input string) (string, error)
}

// MultiAgentSystem 多 Agent 系统
type MultiAgentSystem struct {
	chatModel   model.ChatModel
	agents      map[string]Agent
	coordinator *CoordinatorAgent
}

// NewMultiAgentSystem 创建多 Agent 系统
func NewMultiAgentSystem(chatModel model.ChatModel) *MultiAgentSystem {
	system := &MultiAgentSystem{
		chatModel: chatModel,
		agents:    make(map[string]Agent),
	}

	// 创建专门化的 Agent
	system.agents["planner"] = NewPlannerAgent(chatModel)
	system.agents["researcher"] = NewResearcherAgent(chatModel)
	system.agents["writer"] = NewWriterAgent(chatModel)
	system.agents["reviewer"] = NewReviewerAgent(chatModel)
	system.agents["developer"] = NewDeveloperAgent(chatModel)

	// 创建协调者 Agent
	system.coordinator = NewCoordinatorAgent(chatModel, system.agents)

	return system
}

// ExecuteTask 执行任务
func (s *MultiAgentSystem) ExecuteTask(ctx context.Context, task string) (string, error) {
	fmt.Printf("\n收到任务: %s\n", task)
	fmt.Println("协调者正在分析任务...")

	// 协调者分析任务并制定执行计划
	plan, err := s.coordinator.CreatePlan(ctx, task)
	if err != nil {
		return "", fmt.Errorf("创建计划失败: %w", err)
	}

	fmt.Printf("\n执行计划:\n%s\n", plan)

	// 执行计划中的每个步骤
	results := make(map[string]string)
	for _, step := range plan.Steps {
		fmt.Printf("\n--- 执行步骤: %s (负责人: %s) ---\n", step.Description, step.AgentName)

		agent, exists := s.agents[step.AgentName]
		if !exists {
			return "", fmt.Errorf("未找到 Agent: %s", step.AgentName)
		}

		// 准备输入（包含之前步骤的结果）
		input := step.Description
		if step.DependsOn != "" {
			if prevResult, ok := results[step.DependsOn]; ok {
				input = fmt.Sprintf("%s\n\n前置步骤结果:\n%s", input, prevResult)
			}
		}

		// Agent 处理任务
		result, err := agent.Process(ctx, input)
		if err != nil {
			return "", fmt.Errorf("Agent %s 处理失败: %w", agent.Name(), err)
		}

		results[step.ID] = result
		fmt.Printf("\n%s 完成工作:\n%s\n", agent.Name(), result)
	}

	// 汇总最终结果
	finalResult := s.coordinator.SummarizeResults(ctx, task, results)
	return finalResult, nil
}

// ExecutionPlan 执行计划
type ExecutionPlan struct {
	Steps []ExecutionStep
}

// ExecutionStep 执行步骤
type ExecutionStep struct {
	ID          string
	Description string
	AgentName   string
	DependsOn   string
}

// CoordinatorAgent 协调者 Agent
type CoordinatorAgent struct {
	chatModel model.ChatModel
	agents    map[string]Agent
}

// NewCoordinatorAgent 创建协调者 Agent
func NewCoordinatorAgent(chatModel model.ChatModel, agents map[string]Agent) *CoordinatorAgent {
	return &CoordinatorAgent{
		chatModel: chatModel,
		agents:    agents,
	}
}

// CreatePlan 创建执行计划
func (a *CoordinatorAgent) CreatePlan(ctx context.Context, task string) (*ExecutionPlan, error) {
	// 构建可用 Agent 列表
	agentList := ""
	for name, agent := range a.agents {
		agentList += fmt.Sprintf("- %s: %s\n", name, agent.Role())
	}

	prompt := fmt.Sprintf(`作为协调者，请为以下任务制定执行计划：

任务: %s

可用的 Agent:
%s

请按照以下格式输出执行计划：
Step 1: [步骤描述] - Agent: [agent_name]
Step 2: [步骤描述] - Agent: [agent_name] - Depends on: Step 1
...

确保步骤之间的依赖关系清晰，每个步骤分配给最合适的 Agent。`, task, agentList)

	messages := []*schema.Message{
		schema.SystemMessage("你是一个任务协调专家，擅长分解任务并分配给合适的团队成员。"),
		schema.UserMessage(prompt),
	}

	response, err := a.chatModel.Generate(ctx, messages)
	if err != nil {
		return nil, err
	}

	// 解析执行计划
	plan := a.parsePlan(response.Content)
	return plan, nil
}

// parsePlan 解析执行计划
func (a *CoordinatorAgent) parsePlan(planText string) *ExecutionPlan {
	plan := &ExecutionPlan{
		Steps: []ExecutionStep{},
	}

	lines := strings.Split(planText, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Step") {
			// 简化的解析逻辑
			// 实际应用中需要更健壮的解析
			step := ExecutionStep{
				ID:          fmt.Sprintf("step_%d", len(plan.Steps)+1),
				Description: line,
				AgentName:   "planner", // 默认值
			}

			// 提取 Agent 名称
			if strings.Contains(line, "Agent:") {
				parts := strings.Split(line, "Agent:")
				if len(parts) > 1 {
					agentPart := strings.TrimSpace(parts[1])
					agentName := strings.Split(agentPart, " ")[0]
					agentName = strings.Trim(agentName, "[]")
					step.AgentName = agentName
				}
			}

			plan.Steps = append(plan.Steps, step)
		}
	}

	return plan
}

// SummarizeResults 汇总结果
func (a *CoordinatorAgent) SummarizeResults(ctx context.Context, task string, results map[string]string) string {
	summary := fmt.Sprintf("任务: %s\n\n", task)
	for id, result := range results {
		summary += fmt.Sprintf("%s:\n%s\n\n", id, result)
	}
	return summary
}

// PlannerAgent 规划者 Agent
type PlannerAgent struct {
	chatModel model.ChatModel
}

func NewPlannerAgent(chatModel model.ChatModel) *PlannerAgent {
	return &PlannerAgent{chatModel: chatModel}
}

func (a *PlannerAgent) Name() string { return "Planner" }
func (a *PlannerAgent) Role() string { return "负责任务规划和需求分析" }

func (a *PlannerAgent) Process(ctx context.Context, input string) (string, error) {
	messages := []*schema.Message{
		schema.SystemMessage("你是一个专业的项目规划师，擅长分析需求和制定详细计划。"),
		schema.UserMessage(input),
	}

	response, err := a.chatModel.Generate(ctx, messages)
	if err != nil {
		return "", err
	}

	return response.Content, nil
}

// ResearcherAgent 研究者 Agent
type ResearcherAgent struct {
	chatModel model.ChatModel
}

func NewResearcherAgent(chatModel model.ChatModel) *ResearcherAgent {
	return &ResearcherAgent{chatModel: chatModel}
}

func (a *ResearcherAgent) Name() string { return "Researcher" }
func (a *ResearcherAgent) Role() string { return "负责信息收集和研究分析" }

func (a *ResearcherAgent) Process(ctx context.Context, input string) (string, error) {
	messages := []*schema.Message{
		schema.SystemMessage("你是一个专业的研究员，擅长收集信息、分析数据和提供深入见解。"),
		schema.UserMessage(input),
	}

	response, err := a.chatModel.Generate(ctx, messages)
	if err != nil {
		return "", err
	}

	return response.Content, nil
}

// WriterAgent 写作者 Agent
type WriterAgent struct {
	chatModel model.ChatModel
}

func NewWriterAgent(chatModel model.ChatModel) *WriterAgent {
	return &WriterAgent{chatModel: chatModel}
}

func (a *WriterAgent) Name() string { return "Writer" }
func (a *WriterAgent) Role() string { return "负责内容创作和文档编写" }

func (a *WriterAgent) Process(ctx context.Context, input string) (string, error) {
	messages := []*schema.Message{
		schema.SystemMessage("你是一个专业的内容创作者，擅长撰写清晰、吸引人的文案和文档。"),
		schema.UserMessage(input),
	}

	response, err := a.chatModel.Generate(ctx, messages)
	if err != nil {
		return "", err
	}

	return response.Content, nil
}

// ReviewerAgent 审查者 Agent
type ReviewerAgent struct {
	chatModel model.ChatModel
}

func NewReviewerAgent(chatModel model.ChatModel) *ReviewerAgent {
	return &ReviewerAgent{chatModel: chatModel}
}

func (a *ReviewerAgent) Name() string { return "Reviewer" }
func (a *ReviewerAgent) Role() string { return "负责质量审查和改进建议" }

func (a *ReviewerAgent) Process(ctx context.Context, input string) (string, error) {
	messages := []*schema.Message{
		schema.SystemMessage("你是一个严格的质量审查员，擅长发现问题并提供改进建议。"),
		schema.UserMessage(input),
	}

	response, err := a.chatModel.Generate(ctx, messages)
	if err != nil {
		return "", err
	}

	return response.Content, nil
}

// DeveloperAgent 开发者 Agent
type DeveloperAgent struct {
	chatModel model.ChatModel
}

func NewDeveloperAgent(chatModel model.ChatModel) *DeveloperAgent {
	return &DeveloperAgent{chatModel: chatModel}
}

func (a *DeveloperAgent) Name() string { return "Developer" }
func (a *DeveloperAgent) Role() string { return "负责代码开发和技术实现" }

func (a *DeveloperAgent) Process(ctx context.Context, input string) (string, error) {
	messages := []*schema.Message{
		schema.SystemMessage("你是一个经验丰富的软件开发工程师，擅长设计和实现高质量的代码。"),
		schema.UserMessage(input),
	}

	response, err := a.chatModel.Generate(ctx, messages)
	if err != nil {
		return "", err
	}

	return response.Content, nil
}
