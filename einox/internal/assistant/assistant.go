// Package assistant 提供 AI 学习助手核心功能
package assistant

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"github.com/NuyoahCh/einotelos/einox/callback"
	"github.com/NuyoahCh/einotelos/einox/config"
	"github.com/NuyoahCh/einotelos/einox/model"
	"github.com/NuyoahCh/einotelos/einox/rag"
	"github.com/NuyoahCh/einotelos/einox/tools"
)

// Assistant AI 学习助手
type Assistant struct {
	config    *config.Config
	chatModel *model.ChatModel
	rag       *rag.SimpleRAG
	tools     []tool.BaseTool
	history   []*schema.Message
}

// New 创建助手实例
func New(ctx context.Context, cfg *config.Config) (*Assistant, error) {
	// 创建对话模型
	chatModel, err := model.NewChatModel(ctx, &cfg.ARK)
	if err != nil {
		return nil, fmt.Errorf("创建模型失败: %w", err)
	}

	// 创建 RAG（使用模拟向量器，简化演示）
	mockEmbedder := rag.NewMockEmbedder(128)
	simpleRAG := rag.NewSimpleRAG(mockEmbedder, cfg.RAG.TopK)

	// 获取工具
	allTools := tools.GetAllTools()

	return &Assistant{
		config:    cfg,
		chatModel: chatModel,
		rag:       simpleRAG,
		tools:     allTools,
		history:   make([]*schema.Message, 0),
	}, nil
}

// Close 关闭助手
func (a *Assistant) Close() {
	// 清理资源
}

// ChatStream 流式对话
func (a *Assistant) ChatStream(ctx context.Context, query string) error {
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

	fullContent, err := callback.CollectStream(stream, &callback.StreamCallback{
		OnChunk: func(content string) {
			fmt.Print(content)
		},
	})
	if err != nil {
		return err
	}

	// 更新历史
	a.addHistory(query, fullContent)
	return nil
}

// KnowledgeQA 知识库问答
func (a *Assistant) KnowledgeQA(ctx context.Context, query string) error {
	// 检索相关知识
	ragContext, docs, err := a.rag.Query(ctx, query)
	if err != nil {
		return err
	}

	// 构建消息
	systemPrompt := `你是一个基于知识库的问答助手。请根据提供的参考资料回答问题。

参考资料：
%s

回答原则：
1. 优先使用参考资料中的信息
2. 如果参考资料不足以回答，请明确说明
3. 回答要简洁准确`

	messages := []*schema.Message{
		schema.SystemMessage(fmt.Sprintf(systemPrompt, ragContext)),
		schema.UserMessage(query),
	}

	// 流式输出
	stream, err := a.chatModel.Stream(ctx, messages)
	if err != nil {
		return err
	}
	defer stream.Close()

	_, err = callback.CollectStream(stream, &callback.StreamCallback{
		OnChunk: func(content string) {
			fmt.Print(content)
		},
	})

	// 显示引用来源
	if len(docs) > 0 {
		fmt.Printf("\n\n📎 参考了 %d 条知识", len(docs))
	}

	return err
}

// AddKnowledge 添加知识
func (a *Assistant) AddKnowledge(ctx context.Context, content string) error {
	return a.rag.AddText(ctx, content, map[string]any{"source": "user"})
}

// ClearKnowledge 清空知识库
func (a *Assistant) ClearKnowledge(ctx context.Context) {
	// 重新创建 RAG
	mockEmbedder := rag.NewMockEmbedder(128)
	a.rag = rag.NewSimpleRAG(mockEmbedder, a.config.RAG.TopK)
}

// KnowledgeCount 获取知识数量
func (a *Assistant) KnowledgeCount() int {
	if store, ok := a.rag.GetStore().(*rag.MemoryVectorStore); ok {
		return store.Count()
	}
	return 0
}

// UseTool 使用工具
func (a *Assistant) UseTool(ctx context.Context, query string) (string, error) {
	// 获取工具信息
	toolInfos, err := tools.GetToolInfos(ctx, a.tools)
	if err != nil {
		return "", err
	}

	// 绑定工具
	if err := a.chatModel.BindTools(toolInfos); err != nil {
		return "", err
	}

	// 让模型决定使用哪个工具
	messages := []*schema.Message{
		schema.SystemMessage(`你是一个智能助手，可以使用以下工具：
- calculator: 数学计算
- get_weather: 查询天气
- get_current_time: 获取当前时间

根据用户需求选择合适的工具。`),
		schema.UserMessage(query),
	}

	response, err := a.chatModel.Generate(ctx, messages)
	if err != nil {
		return "", err
	}

	// 如果有工具调用
	if len(response.ToolCalls) > 0 {
		var results []string
		for _, tc := range response.ToolCalls {
			// 执行工具
			result := a.executeTool(ctx, tc.Function.Name, tc.Function.Arguments)
			results = append(results, fmt.Sprintf("%s: %s", tc.Function.Name, result))
		}

		// 让模型总结结果
		summaryMessages := []*schema.Message{
			schema.SystemMessage("请用自然语言简洁地总结工具返回的结果。"),
			schema.UserMessage(fmt.Sprintf("用户问题: %s\n工具结果: %v", query, results)),
		}

		summary, err := a.chatModel.Generate(ctx, summaryMessages)
		if err != nil {
			return results[0], nil
		}
		return summary.Content, nil
	}

	return response.Content, nil
}

func (a *Assistant) executeTool(ctx context.Context, name, args string) string {
	for _, t := range a.tools {
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

// CodeAssist 代码助手
func (a *Assistant) CodeAssist(ctx context.Context, query string) error {
	messages := []*schema.Message{
		schema.SystemMessage(`你是一个专业的编程助手，精通多种编程语言。
回答要求：
1. 代码示例使用 markdown 代码块
2. 解释要简洁清晰
3. 注意代码的正确性和最佳实践`),
		schema.UserMessage(query),
	}

	stream, err := a.chatModel.Stream(ctx, messages)
	if err != nil {
		return err
	}
	defer stream.Close()

	_, err = callback.CollectStream(stream, &callback.StreamCallback{
		OnChunk: func(content string) {
			fmt.Print(content)
		},
	})
	return err
}

// Translate 翻译
func (a *Assistant) Translate(ctx context.Context, text string) error {
	messages := []*schema.Message{
		schema.SystemMessage(`你是一个专业的翻译助手。
规则：
1. 如果输入是中文，翻译成英文
2. 如果输入是英文，翻译成中文
3. 只输出翻译结果，不要解释`),
		schema.UserMessage(text),
	}

	stream, err := a.chatModel.Stream(ctx, messages)
	if err != nil {
		return err
	}
	defer stream.Close()

	_, err = callback.CollectStream(stream, &callback.StreamCallback{
		OnChunk: func(content string) {
			fmt.Print(content)
		},
	})
	return err
}

func (a *Assistant) addHistory(query, response string) {
	a.history = append(a.history, schema.UserMessage(query))
	a.history = append(a.history, schema.AssistantMessage(response, nil))

	// 限制历史长度
	if len(a.history) > 10 {
		a.history = a.history[len(a.history)-10:]
	}
}
