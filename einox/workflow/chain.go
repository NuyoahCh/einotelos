// Package workflow 提供工作流编排组件
package workflow

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	"github.com/NuyoahCh/einotelos/einox/model"
)

// ChainBuilder Chain 构建器
type ChainBuilder struct {
	ctx       context.Context
	chatModel *model.ChatModel
	tools     []tool.BaseTool
	template  *prompt.DefaultChatTemplate
}

// NewChainBuilder 创建 Chain 构建器
func NewChainBuilder(ctx context.Context, chatModel *model.ChatModel) *ChainBuilder {
	return &ChainBuilder{
		ctx:       ctx,
		chatModel: chatModel,
	}
}

// WithTemplate 设置模板
func (b *ChainBuilder) WithTemplate(tpl *prompt.DefaultChatTemplate) *ChainBuilder {
	b.template = tpl
	return b
}

// WithTools 设置工具
func (b *ChainBuilder) WithTools(tools []tool.BaseTool) *ChainBuilder {
	b.tools = tools
	return b
}

// BuildSimpleChain 构建简单对话链
func (b *ChainBuilder) BuildSimpleChain() (compose.Runnable[map[string]any, *schema.Message], error) {
	chain := compose.NewChain[map[string]any, *schema.Message]()

	// 添加模板节点
	if b.template != nil {
		chain.AppendChatTemplate(b.template)
	}

	// 添加模型节点
	chain.AppendChatModel(b.chatModel.Inner())

	return chain.Compile(b.ctx)
}

// BuildToolChain 构建工具调用链
func (b *ChainBuilder) BuildToolChain() (compose.Runnable[map[string]any, *schema.Message], error) {
	if len(b.tools) == 0 {
		return nil, fmt.Errorf("工具列表为空")
	}

	// 获取工具信息并绑定到模型
	toolInfos := make([]*schema.ToolInfo, 0, len(b.tools))
	for _, t := range b.tools {
		info, err := t.Info(b.ctx)
		if err != nil {
			return nil, fmt.Errorf("获取工具信息失败: %w", err)
		}
		toolInfos = append(toolInfos, info)
	}

	if err := b.chatModel.BindTools(toolInfos); err != nil {
		return nil, fmt.Errorf("绑定工具失败: %w", err)
	}

	// 创建工具节点
	toolsNode, err := compose.NewToolNode(b.ctx, &compose.ToolsNodeConfig{
		Tools: b.tools,
	})
	if err != nil {
		return nil, fmt.Errorf("创建工具节点失败: %w", err)
	}

	// Lambda: 提取工具结果转为用户消息
	extractToolResult := compose.TransformableLambda[[]*schema.Message, *schema.Message](
		func(ctx context.Context, input *schema.StreamReader[[]*schema.Message]) (*schema.StreamReader[*schema.Message], error) {
			return schema.StreamReaderWithConvert(input, func(msgs []*schema.Message) (*schema.Message, error) {
				if len(msgs) == 0 {
					return schema.UserMessage("工具未返回结果"), nil
				}

				// 收集工具结果
				var results []string
				for _, m := range msgs {
					if m != nil && m.Content != "" {
						results = append(results, m.Content)
					}
				}

				if len(results) == 0 {
					return schema.UserMessage("工具未返回有效结果"), nil
				}

				return schema.UserMessage("工具执行结果：\n" + results[0]), nil
			}), nil
		},
	)

	// Lambda: 构建最终回复的 prompt
	buildFinalPrompt := compose.TransformableLambda[*schema.Message, []*schema.Message](
		func(ctx context.Context, input *schema.StreamReader[*schema.Message]) (*schema.StreamReader[[]*schema.Message], error) {
			return schema.StreamReaderWithConvert(input, func(m *schema.Message) ([]*schema.Message, error) {
				return []*schema.Message{
					schema.SystemMessage("请根据工具返回的结果，用自然语言回答用户的问题。"),
					m,
				}, nil
			}), nil
		},
	)

	// 构建链
	chain := compose.NewChain[map[string]any, *schema.Message]()

	if b.template != nil {
		chain.AppendChatTemplate(b.template)
	}

	chain.
		AppendChatModel(b.chatModel.Inner()).
		AppendToolsNode(toolsNode).
		AppendLambda(extractToolResult).
		AppendLambda(buildFinalPrompt).
		AppendChatModel(b.chatModel.Inner())

	return chain.Compile(b.ctx)
}

// RunChain 运行链
func RunChain(ctx context.Context, chain compose.Runnable[map[string]any, *schema.Message], query string, history []*schema.Message) (*schema.Message, error) {
	return chain.Invoke(ctx, map[string]any{
		"query":   query,
		"history": history,
	})
}

// ToolResultExtractor 工具结果提取器
type ToolResultExtractor struct{}

func (e *ToolResultExtractor) Extract(msgs []*schema.Message) string {
	var results []map[string]any

	for _, m := range msgs {
		if m == nil || m.Content == "" {
			continue
		}

		var data map[string]any
		if err := json.Unmarshal([]byte(m.Content), &data); err == nil {
			results = append(results, data)
		} else {
			results = append(results, map[string]any{"result": m.Content})
		}
	}

	if len(results) == 0 {
		return "无结果"
	}

	b, _ := json.MarshalIndent(results, "", "  ")
	return string(b)
}
