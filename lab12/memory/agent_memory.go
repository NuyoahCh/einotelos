package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// AgentMemory 演示 Agent 的记忆管理
// Agent 可以记住之前的对话和交互，提供更连贯的体验
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

	// 示例 1：短期记忆（对话历史）
	fmt.Println("=== 示例 1: 短期记忆（对话历史）===")
	demonstrateShortTermMemory(ctx, chatModel)

	// 示例 2：长期记忆（持久化存储）
	fmt.Println("\n=== 示例 2: 长期记忆（持久化存储）===")
	demonstrateLongTermMemory(ctx, chatModel)

	// 示例 3：语义记忆（向量检索）
	fmt.Println("\n=== 示例 3: 语义记忆（向量检索）===")
	demonstrateSemanticMemory(ctx, chatModel)

	// 示例 4：工作记忆（任务上下文）
	fmt.Println("\n=== 示例 4: 工作记忆（任务上下文）===")
	demonstrateWorkingMemory(ctx, chatModel)
}

// demonstrateShortTermMemory 演示短期记忆
func demonstrateShortTermMemory(ctx context.Context, chatModel model.ChatModel) {
	// 创建带记忆的 Agent
	agent := NewMemoryAgent(chatModel, 10) // 保留最近 10 条消息

	// 多轮对话
	conversations := []string{
		"你好，我叫张三。",
		"我喜欢打篮球。",
		"我的生日是 5 月 20 日。",
		"我最喜欢的球队是湖人队。",
		"请告诉我，我叫什么名字？",
		"我喜欢什么运动？",
		"我的生日是什么时候？",
	}

	for i, userInput := range conversations {
		fmt.Printf("\n轮次 %d:\n", i+1)
		fmt.Printf("用户: %s\n", userInput)

		response, err := agent.Chat(ctx, userInput)
		if err != nil {
			log.Printf("对话失败: %v", err)
			continue
		}

		fmt.Printf("Agent: %s\n", response)
	}
}

// demonstrateLongTermMemory 演示长期记忆
func demonstrateLongTermMemory(ctx context.Context, chatModel model.ChatModel) {
	// 创建带持久化记忆的 Agent
	storage := NewMemoryStorage()
	agent := NewPersistentMemoryAgent(chatModel, storage)

	// 第一次会话
	fmt.Println("\n--- 第一次会话 ---")
	agent.Chat(ctx, "我的项目名称是 EinoTelos，这是一个 LLM 教学项目。")
	agent.Chat(ctx, "项目使用 Go 语言开发。")
	agent.Chat(ctx, "项目的主要目标是帮助开发者学习 Eino 框架。")

	// 保存记忆
	agent.SaveMemory("user_001")

	// 模拟新会话（加载之前的记忆）
	fmt.Println("\n--- 第二次会话（加载记忆）---")
	newAgent := NewPersistentMemoryAgent(chatModel, storage)
	newAgent.LoadMemory("user_001")

	response, _ := newAgent.Chat(ctx, "我的项目叫什么名字？使用什么语言开发？")
	fmt.Printf("Agent: %s\n", response)
}

// demonstrateSemanticMemory 演示语义记忆
func demonstrateSemanticMemory(ctx context.Context, chatModel model.ChatModel) {
	// 创建语义记忆 Agent
	agent := NewSemanticMemoryAgent(chatModel)

	// 添加知识到语义记忆
	knowledge := []string{
		"Eino 是字节跳动开源的 LLM 应用开发框架。",
		"Go 是一种静态类型的编译型编程语言。",
		"RAG 是检索增强生成技术，结合了检索和生成。",
		"Agent 是能够自主决策和行动的智能体。",
		"向量数据库用于存储和检索高维向量数据。",
	}

	for _, k := range knowledge {
		agent.AddKnowledge(ctx, k)
	}

	// 基于语义相似度检索相关知识
	queries := []string{
		"什么是 Eino？",
		"告诉我关于 RAG 的信息。",
		"Go 语言有什么特点？",
	}

	for _, query := range queries {
		fmt.Printf("\n查询: %s\n", query)
		response, _ := agent.QueryWithMemory(ctx, query)
		fmt.Printf("回答: %s\n", response)
	}
}

// demonstrateWorkingMemory 演示工作记忆
func demonstrateWorkingMemory(ctx context.Context, chatModel model.ChatModel) {
	// 创建带工作记忆的 Agent
	agent := NewWorkingMemoryAgent(chatModel)

	// 执行多步骤任务
	fmt.Println("\n任务: 计算 (10 + 5) * 3 - 8")

	agent.Chat(ctx, "第一步：计算 10 + 5")
	agent.Chat(ctx, "第二步：将结果乘以 3")
	agent.Chat(ctx, "第三步：减去 8")
	agent.Chat(ctx, "最终结果是多少？")
}

// MemoryAgent 带记忆的 Agent
type MemoryAgent struct {
	chatModel           model.ChatModel
	conversationHistory []*schema.Message
	maxHistory          int
}

// NewMemoryAgent 创建带记忆的 Agent
func NewMemoryAgent(chatModel model.ChatModel, maxHistory int) *MemoryAgent {
	return &MemoryAgent{
		chatModel: chatModel,
		conversationHistory: []*schema.Message{
			schema.SystemMessage("你是一个有记忆的助手，能够记住之前的对话内容。"),
		},
		maxHistory: maxHistory,
	}
}

// Chat 进行对话
func (a *MemoryAgent) Chat(ctx context.Context, userInput string) (string, error) {
	// 添加用户消息到历史
	a.conversationHistory = append(a.conversationHistory, schema.UserMessage(userInput))

	// 限制历史长度（保留系统消息 + 最近的 N 条消息）
	if len(a.conversationHistory) > a.maxHistory+1 {
		// 保留系统消息和最近的消息
		a.conversationHistory = append(
			a.conversationHistory[:1],
			a.conversationHistory[len(a.conversationHistory)-a.maxHistory:]...,
		)
	}

	// 生成响应
	response, err := a.chatModel.Generate(ctx, a.conversationHistory)
	if err != nil {
		return "", err
	}

	// 添加 AI 响应到历史
	a.conversationHistory = append(a.conversationHistory, response)

	return response.Content, nil
}

// PersistentMemoryAgent 带持久化记忆的 Agent
type PersistentMemoryAgent struct {
	chatModel model.ChatModel
	storage   *MemoryStorage
	userID    string
	messages  []*schema.Message
}

// NewPersistentMemoryAgent 创建带持久化记忆的 Agent
func NewPersistentMemoryAgent(chatModel model.ChatModel, storage *MemoryStorage) *PersistentMemoryAgent {
	return &PersistentMemoryAgent{
		chatModel: chatModel,
		storage:   storage,
		messages: []*schema.Message{
			schema.SystemMessage("你是一个助手，能够记住用户的信息和偏好。"),
		},
	}
}

// Chat 进行对话
func (a *PersistentMemoryAgent) Chat(ctx context.Context, userInput string) (string, error) {
	a.messages = append(a.messages, schema.UserMessage(userInput))

	response, err := a.chatModel.Generate(ctx, a.messages)
	if err != nil {
		return "", err
	}

	a.messages = append(a.messages, response)
	return response.Content, nil
}

// SaveMemory 保存记忆
func (a *PersistentMemoryAgent) SaveMemory(userID string) {
	a.userID = userID
	a.storage.Save(userID, a.messages)
	fmt.Printf("\n记忆已保存（用户: %s）\n", userID)
}

// LoadMemory 加载记忆
func (a *PersistentMemoryAgent) LoadMemory(userID string) {
	a.userID = userID
	if messages, exists := a.storage.Load(userID); exists {
		a.messages = messages
		fmt.Printf("\n记忆已加载（用户: %s，共 %d 条消息）\n", userID, len(messages))
	}
}

// MemoryStorage 记忆存储（简化实现）
type MemoryStorage struct {
	data map[string][]*schema.Message
}

// NewMemoryStorage 创建记忆存储
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string][]*schema.Message),
	}
}

// Save 保存记忆
func (s *MemoryStorage) Save(userID string, messages []*schema.Message) {
	s.data[userID] = messages
}

// Load 加载记忆
func (s *MemoryStorage) Load(userID string) ([]*schema.Message, bool) {
	messages, exists := s.data[userID]
	return messages, exists
}

// SemanticMemoryAgent 语义记忆 Agent
type SemanticMemoryAgent struct {
	chatModel     model.ChatModel
	knowledgeBase []string
}

// NewSemanticMemoryAgent 创建语义记忆 Agent
func NewSemanticMemoryAgent(chatModel model.ChatModel) *SemanticMemoryAgent {
	return &SemanticMemoryAgent{
		chatModel:     chatModel,
		knowledgeBase: []string{},
	}
}

// AddKnowledge 添加知识
func (a *SemanticMemoryAgent) AddKnowledge(ctx context.Context, knowledge string) {
	a.knowledgeBase = append(a.knowledgeBase, knowledge)
	fmt.Printf("已添加知识: %s\n", knowledge)
}

// QueryWithMemory 基于记忆查询
func (a *SemanticMemoryAgent) QueryWithMemory(ctx context.Context, query string) (string, error) {
	// 简化实现：将所有知识作为上下文
	// 实际应用中应该使用向量检索找到最相关的知识
	context := "相关知识:\n"
	for _, k := range a.knowledgeBase {
		context += fmt.Sprintf("- %s\n", k)
	}

	messages := []*schema.Message{
		schema.SystemMessage("你是一个知识助手，请基于提供的知识回答问题。"),
		schema.UserMessage(fmt.Sprintf("%s\n\n问题: %s", context, query)),
	}

	response, err := a.chatModel.Generate(ctx, messages)
	if err != nil {
		return "", err
	}

	return response.Content, nil
}

// WorkingMemoryAgent 工作记忆 Agent
type WorkingMemoryAgent struct {
	chatModel     model.ChatModel
	workingMemory map[string]interface{}
	messages      []*schema.Message
}

// NewWorkingMemoryAgent 创建工作记忆 Agent
func NewWorkingMemoryAgent(chatModel model.ChatModel) *WorkingMemoryAgent {
	return &WorkingMemoryAgent{
		chatModel:     chatModel,
		workingMemory: make(map[string]interface{}),
		messages: []*schema.Message{
			schema.SystemMessage("你是一个助手，能够记住任务执行过程中的中间结果。"),
		},
	}
}

// Chat 进行对话
func (a *WorkingMemoryAgent) Chat(ctx context.Context, userInput string) (string, error) {
	// 构建包含工作记忆的提示
	memoryContext := "\n当前工作记忆:\n"
	for key, value := range a.workingMemory {
		memoryContext += fmt.Sprintf("- %s: %v\n", key, value)
	}

	a.messages = append(a.messages, schema.UserMessage(userInput+memoryContext))

	response, err := a.chatModel.Generate(ctx, a.messages)
	if err != nil {
		return "", err
	}

	a.messages = append(a.messages, response)

	// 更新工作记忆（简化实现）
	a.workingMemory[fmt.Sprintf("step_%d", len(a.workingMemory)+1)] = response.Content

	fmt.Printf("Agent: %s\n", response.Content)
	return response.Content, nil
}

// MemoryEntry 记忆条目
type MemoryEntry struct {
	Timestamp time.Time
	Content   string
	Type      string // "conversation", "fact", "task"
	Metadata  map[string]interface{}
}
