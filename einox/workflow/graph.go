// Package workflow 提供 Graph 工作流组件
package workflow

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	"github.com/NuyoahCh/einotelos/einox/model"
)

// GraphBuilder Graph 构建器
type GraphBuilder struct {
	ctx       context.Context
	chatModel *model.ChatModel
	tools     []tool.BaseTool
	template  *prompt.DefaultChatTemplate
}

// NewGraphBuilder 创建 Graph 构建器
func NewGraphBuilder(ctx context.Context, chatModel *model.ChatModel) *GraphBuilder {
	return &GraphBuilder{
		ctx:       ctx,
		chatModel: chatModel,
	}
}

// WithTemplate 设置模板
func (b *GraphBuilder) WithTemplate(tpl *prompt.DefaultChatTemplate) *GraphBuilder {
	b.template = tpl
	return b
}

// WithTools 设置工具
func (b *GraphBuilder) WithTools(tools []tool.BaseTool) *GraphBuilder {
	b.tools = tools
	return b
}

// BuildToolGraph 构建工具调用图
func (b *GraphBuilder) BuildToolGraph() (compose.Runnable[map[string]any, *schema.Message], error) {
	if len(b.tools) == 0 {
		return nil, fmt.Errorf("工具列表为空")
	}

	// 获取工具信息并绑定
	toolInfos := make([]*schema.ToolInfo, 0, len(b.tools))
	for _, t := range b.tools {
		info, err := t.Info(b.ctx)
		if err != nil {
			return nil, err
		}
		toolInfos = append(toolInfos, info)
	}

	if err := b.chatModel.BindTools(toolInfos); err != nil {
		return nil, err
	}

	// 创建工具节点
	toolsNode, err := compose.NewToolNode(b.ctx, &compose.ToolsNodeConfig{
		Tools: b.tools,
	})
	if err != nil {
		return nil, err
	}

	// 创建 Graph
	g := compose.NewGraph[map[string]any, *schema.Message]()

	// 节点名称
	const (
		nodePrompt      = "prompt"
		nodeChat        = "chat"
		nodeTools       = "tools"
		nodeExtract     = "extract"
		nodeBuildPrompt = "build_prompt"
		nodeFinalChat   = "final_chat"
	)

	// Lambda: 提取工具结果
	extractLambda := compose.TransformableLambda[[]*schema.Message, *schema.Message](
		func(ctx context.Context, input *schema.StreamReader[[]*schema.Message]) (*schema.StreamReader[*schema.Message], error) {
			return schema.StreamReaderWithConvert(input, func(msgs []*schema.Message) (*schema.Message, error) {
				var content string
				for _, m := range msgs {
					if m != nil && m.Content != "" {
						content += m.Content + "\n"
					}
				}
				if content == "" {
					content = "工具未返回结果"
				}
				return schema.UserMessage("工具执行结果：\n" + content), nil
			}), nil
		},
	)

	// Lambda: 构建最终 prompt
	buildPromptLambda := compose.TransformableLambda[*schema.Message, []*schema.Message](
		func(ctx context.Context, input *schema.StreamReader[*schema.Message]) (*schema.StreamReader[[]*schema.Message], error) {
			return schema.StreamReaderWithConvert(input, func(m *schema.Message) ([]*schema.Message, error) {
				return []*schema.Message{
					schema.SystemMessage("请根据工具返回的结果，用自然语言总结并回答用户的问题。"),
					m,
				}, nil
			}), nil
		},
	)

	// 添加节点
	if b.template != nil {
		_ = g.AddChatTemplateNode(nodePrompt, b.template)
	}
	_ = g.AddChatModelNode(nodeChat, b.chatModel.Inner())
	_ = g.AddToolsNode(nodeTools, toolsNode)
	_ = g.AddLambdaNode(nodeExtract, extractLambda)
	_ = g.AddLambdaNode(nodeBuildPrompt, buildPromptLambda)
	_ = g.AddChatModelNode(nodeFinalChat, b.chatModel.Inner())

	// 添加边
	if b.template != nil {
		_ = g.AddEdge(compose.START, nodePrompt)
		_ = g.AddEdge(nodePrompt, nodeChat)
	} else {
		_ = g.AddEdge(compose.START, nodeChat)
	}
	_ = g.AddEdge(nodeChat, nodeTools)
	_ = g.AddEdge(nodeTools, nodeExtract)
	_ = g.AddEdge(nodeExtract, nodeBuildPrompt)
	_ = g.AddEdge(nodeBuildPrompt, nodeFinalChat)
	_ = g.AddEdge(nodeFinalChat, compose.END)

	return g.Compile(b.ctx)
}

// BuildRAGGraph 构建 RAG 图
func (b *GraphBuilder) BuildRAGGraph(retriever func(context.Context, string) (string, error)) (compose.Runnable[map[string]any, *schema.Message], error) {
	g := compose.NewGraph[map[string]any, *schema.Message]()

	const (
		nodeRetrieve = "retrieve"
		nodePrompt   = "prompt"
		nodeChat     = "chat"
	)

	// Lambda: 检索
	retrieveLambda := compose.InvokableLambda(
		func(ctx context.Context, input map[string]any) (map[string]any, error) {
			query, _ := input["query"].(string)
			context, err := retriever(ctx, query)
			if err != nil {
				return nil, err
			}
			input["context"] = context
			return input, nil
		},
	)

	// RAG 模板
	ragTemplate := prompt.FromMessages(schema.FString,
		schema.SystemMessage(`你是一个基于知识库的问答助手。请根据提供的参考资料回答用户问题。

参考资料：
{context}

---
回答原则：
1. 优先使用参考资料中的信息
2. 如果参考资料不足，请明确说明
3. 不要编造信息`),
		schema.MessagesPlaceholder("history", true),
		schema.UserMessage("{query}"),
	)

	// 添加节点
	_ = g.AddLambdaNode(nodeRetrieve, retrieveLambda)
	_ = g.AddChatTemplateNode(nodePrompt, ragTemplate)
	_ = g.AddChatModelNode(nodeChat, b.chatModel.Inner())

	// 添加边
	_ = g.AddEdge(compose.START, nodeRetrieve)
	_ = g.AddEdge(nodeRetrieve, nodePrompt)
	_ = g.AddEdge(nodePrompt, nodeChat)
	_ = g.AddEdge(nodeChat, compose.END)

	return g.Compile(b.ctx)
}

// RunGraph 运行图
func RunGraph(ctx context.Context, graph compose.Runnable[map[string]any, *schema.Message], input map[string]any) (*schema.Message, error) {
	return graph.Invoke(ctx, input)
}
