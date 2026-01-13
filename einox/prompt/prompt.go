// Package prompt 提供提示词模板组件
package prompt

import (
	"context"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

// Templates 预定义的提示词模板
var Templates = struct {
	// 通用助手
	GeneralAssistant *prompt.DefaultChatTemplate
	// 代码助手
	CodeAssistant *prompt.DefaultChatTemplate
	// 翻译助手
	TranslationAssistant *prompt.DefaultChatTemplate
	// RAG 问答
	RAGAssistant *prompt.DefaultChatTemplate
	// 工具调用
	ToolAssistant *prompt.DefaultChatTemplate
}{}

func init() {
	// 通用助手模板
	Templates.GeneralAssistant = prompt.FromMessages(schema.FString,
		schema.SystemMessage(`你是一个智能助手，由字节跳动的豆包大模型驱动。
你的特点：
- 知识渊博，能够回答各种问题
- 逻辑清晰，善于分析问题
- 语言流畅，表达准确
- 乐于助人，态度友好

请根据用户的问题提供有帮助的回答。`),
		schema.MessagesPlaceholder("history", true),
		schema.UserMessage("{query}"),
	)

	// 代码助手模板
	Templates.CodeAssistant = prompt.FromMessages(schema.FString,
		schema.SystemMessage(`你是一个专业的编程助手，精通多种编程语言和技术栈。
你的能力：
- 代码编写：能够编写高质量、可维护的代码
- 代码审查：能够发现代码中的问题并提出改进建议
- 问题调试：能够帮助定位和解决代码中的 bug
- 技术解答：能够解释技术概念和最佳实践

请用专业但易懂的方式回答编程相关问题。代码示例请使用 markdown 代码块格式。`),
		schema.MessagesPlaceholder("history", true),
		schema.UserMessage("{query}"),
	)

	// 翻译助手模板
	Templates.TranslationAssistant = prompt.FromMessages(schema.FString,
		schema.SystemMessage(`你是一个专业的翻译助手，精通中英文互译。
翻译原则：
- 准确传达原文含义
- 符合目标语言的表达习惯
- 保持原文的语气和风格
- 专业术语翻译准确

请将用户提供的文本翻译成目标语言。如果用户没有指定目标语言，中文翻译成英文，英文翻译成中文。`),
		schema.MessagesPlaceholder("history", true),
		schema.UserMessage("{query}"),
	)

	// RAG 问答模板
	Templates.RAGAssistant = prompt.FromMessages(schema.FString,
		schema.SystemMessage(`你是一个基于知识库的问答助手。请根据提供的参考资料回答用户问题。

回答原则：
1. 优先使用参考资料中的信息
2. 如果参考资料不足以回答问题，请明确说明
3. 不要编造参考资料中没有的信息
4. 可以适当整合和总结多个来源的信息

参考资料：
{context}

---`),
		schema.MessagesPlaceholder("history", true),
		schema.UserMessage("{query}"),
	)

	// 工具调用模板
	Templates.ToolAssistant = prompt.FromMessages(schema.FString,
		schema.SystemMessage(`你是一个智能助手，可以使用各种工具来帮助用户完成任务。

可用工具：
- calculator: 执行数学计算
- get_weather: 查询天气信息
- get_current_time: 获取当前时间
- web_search: 搜索网络信息

使用原则：
1. 根据用户需求选择合适的工具
2. 可以组合使用多个工具
3. 工具返回结果后，用自然语言总结给用户
4. 如果不需要工具，直接回答即可`),
		schema.MessagesPlaceholder("history", true),
		schema.UserMessage("{query}"),
	)
}

// NewCustomTemplate 创建自定义模板
func NewCustomTemplate(systemPrompt string) *prompt.DefaultChatTemplate {
	return prompt.FromMessages(schema.FString,
		schema.SystemMessage(systemPrompt),
		schema.MessagesPlaceholder("history", true),
		schema.UserMessage("{query}"),
	)
}

// FormatMessages 格式化消息
func FormatMessages(ctx context.Context, tpl *prompt.DefaultChatTemplate, query string, history []*schema.Message) ([]*schema.Message, error) {
	return tpl.Format(ctx, map[string]any{
		"query":   query,
		"history": history,
	})
}

// FormatRAGMessages 格式化 RAG 消息
func FormatRAGMessages(ctx context.Context, query string, context string, history []*schema.Message) ([]*schema.Message, error) {
	return Templates.RAGAssistant.Format(ctx, map[string]any{
		"query":   query,
		"context": context,
		"history": history,
	})
}
