package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"

	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// 工具入参
type playerInfoRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// 工具出参
type playerInfoResponse struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Role        string `json:"role"`         // 后卫/锋线/中锋/教练/爱好者
	HeightCM    int    `json:"height_cm"`    // 身高
	WeightKG    int    `json:"weight_kg"`    // 体重
	PlayStyle   string `json:"play_style"`   // 风格
	WeeklyHours int    `json:"weekly_hours"` // 每周训练/打球时长
}

func main() {
	ctx := context.Background()
	g := compose.NewGraph[map[string]any, *schema.Message]()

	// 1) ChatTemplate 节点（篮球主题）
	systemTpl := `你是一名篮球教练与比赛分析师。你需要结合用户的基本信息与训练习惯，
使用 player_info API，为其补全信息，然后给出适合他的训练计划、位置建议与一套简单战术建议。
注意：邮箱必须出现，用于查询信息。`

	chatTpl := prompt.FromMessages(schema.FString,
		schema.SystemMessage(systemTpl),
		schema.MessagesPlaceholder("histories", true),
		schema.UserMessage("{user_query}"),
	)

	// 2) 推荐模板
	recommendTpl := `
你是一名篮球教练与比赛分析师。请结合工具返回的用户信息，为用户输出建议，要求具体、可执行。

--- 训练资源（可选方案库）---

### A. 训练方向库（按位置/风格）
**1. 后卫（控运与节奏）**
- 核心：运球对抗、挡拆阅读、急停跳投、突破分球
- 训练：左右手变向组合、弱侧手终结、1v1 变速

**2. 锋线（持球终结与防守）**
- 核心：三威胁、低位脚步、协防轮转、错位单打
- 训练：三分接投+一运、背身转身、closeout 防守

**3. 内线（篮下统治与护框）**
- 核心：卡位、顺下吃饼、护框、二次进攻
- 训练：对抗上篮、掩护质量、篮板站位

### B. 一套简单战术（适合大多数业余队）
- **高位挡拆（P&R）**：持球人借掩护突破/投篮/分球，弱侧埋伏投手
- **Spain P&R（简化版）**：挡拆后再给顺下人做背掩护，制造错位/空切
- **5-out（五外）**：拉开空间，强弱侧转移球，靠突破分球创造空位三分

### C. 输出规则
1) 先总结用户画像（身高体重、风格、每周训练时长）
2) 给出建议位置与核心技能树（3-5个技能）
3) 输出一周训练计划（按天、每次45-90分钟）
4) 给一套战术建议 + 业余局实战注意事项（3条）
`

	// 3) DeepSeek ChatModel（环境变量取 key）
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		log.Fatal("缺少环境变量 DEEPSEEK_API_KEY")
	}

	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  os.Getenv("DEEPSEEK_API_KEY"),
		Model:   "deepseek-chat",
		BaseURL: "https://api.deepseek.com",
	})
	if err != nil {
		log.Fatalf("创建 DeepSeek ChatModel 失败: %v", err)
	}

	// 4) 工具：player_info（mock 示例）
	playerInfoTool := utils.NewTool(
		&schema.ToolInfo{
			Name: "player_info",
			Desc: "根据用户的姓名和邮箱，查询用户的篮球相关信息（位置倾向、身体数据、打球习惯等）",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"name":  {Type: "string", Desc: "用户的姓名"},
				"email": {Type: "string", Desc: "用户的邮箱"},
			}),
		},
		func(ctx context.Context, input *playerInfoRequest) (*playerInfoResponse, error) {
			// demo：随便 mock 一份
			return &playerInfoResponse{
				Name:        input.Name,
				Email:       input.Email,
				Role:        "锋线",
				HeightCM:    182,
				WeightKG:    78,
				PlayStyle:   "偏投射+无球空切，偶尔持球突破",
				WeeklyHours: 4,
			}, nil
		},
	)

	// 5) 绑定工具到模型（让模型能产生 tool_calls）
	info, err := playerInfoTool.Info(ctx)
	if err != nil {
		panic(err)
	}
	if err := chatModel.BindTools([]*schema.ToolInfo{info}); err != nil {
		panic(err)
	}

	// 6) ToolsNode
	toolsNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
		Tools: []tool.BaseTool{playerInfoTool},
	})
	if err != nil {
		panic(err)
	}

	// 7) Lambda：从 toolsNode 输出 messages 中提取工具结果 -> 转成普通 user 文本
	// 这样第二次调用 ChatModel 时不携带 role=tool，避免 400。
	extractToolOps := func(
		ctx context.Context,
		input *schema.StreamReader[[]*schema.Message],
	) (*schema.StreamReader[*schema.Message], error) {
		return schema.StreamReaderWithConvert(input, func(msgs []*schema.Message) (*schema.Message, error) {
			if len(msgs) == 0 {
				return nil, errors.New("no messages from tools node")
			}

			// 这里用一个更“保守”的策略：把 toolsNode 输出的所有 message 内容拼起来，
			// 作为纯文本喂给下一轮（不依赖 Role 字段是否叫 Role）
			type msgLite struct {
				Role    string `json:"role,omitempty"`
				Content string `json:"content,omitempty"`
			}
			lites := make([]msgLite, 0, len(msgs))
			for _, m := range msgs {
				if m == nil {
					continue
				}
				lites = append(lites, msgLite{
					Role:    string(m.Role), // schema.Message 的 Role 字段
					Content: m.Content,      // Content 一般都有
				})
			}
			b, _ := json.MarshalIndent(lites, "", "  ")

			text := "工具执行完成，返回信息如下（结构化摘要）：\n" + string(b)
			return schema.UserMessage(text), nil
		}), nil
	}
	extractToolLambda := compose.TransformableLambda[[]*schema.Message, *schema.Message](extractToolOps)

	// 8) Lambda：构造第二次模型输入：system(recommendTpl) + user(工具结果文本)
	buildPromptOps := func(
		ctx context.Context,
		input *schema.StreamReader[*schema.Message],
	) (*schema.StreamReader[[]*schema.Message], error) {
		return schema.StreamReaderWithConvert(input, func(m *schema.Message) ([]*schema.Message, error) {
			if m == nil {
				return nil, errors.New("nil message")
			}
			out := make([]*schema.Message, 0, 2)
			out = append(out, schema.SystemMessage(recommendTpl))
			out = append(out, m)
			return out, nil
		}), nil
	}
	buildPromptLambda := compose.TransformableLambda[*schema.Message, []*schema.Message](buildPromptOps)

	// 9) Graph 编排
	const (
		promptNodeKey        = "prompt"
		chatNodeKey          = "chat"
		toolsNodeKey         = "tools"
		extractNodeKey       = "extract_tool_result"
		lambdaPromptNodeKey  = "build_recommend_prompt"
		recommendChatNodeKey = "chat_recommend"
	)

	_ = g.AddChatTemplateNode(promptNodeKey, chatTpl)
	_ = g.AddChatModelNode(chatNodeKey, chatModel)
	_ = g.AddToolsNode(toolsNodeKey, toolsNode)
	_ = g.AddLambdaNode(extractNodeKey, extractToolLambda)
	_ = g.AddLambdaNode(lambdaPromptNodeKey, buildPromptLambda)
	_ = g.AddChatModelNode(recommendChatNodeKey, chatModel)

	_ = g.AddEdge(compose.START, promptNodeKey)
	_ = g.AddEdge(promptNodeKey, chatNodeKey)
	_ = g.AddEdge(chatNodeKey, toolsNodeKey)
	_ = g.AddEdge(toolsNodeKey, extractNodeKey)
	_ = g.AddEdge(extractNodeKey, lambdaPromptNodeKey)
	_ = g.AddEdge(lambdaPromptNodeKey, recommendChatNodeKey)
	_ = g.AddEdge(recommendChatNodeKey, compose.END)

	// 10) 编译运行
	runnable, err := g.Compile(ctx)
	if err != nil {
		panic(err)
	}

	output, err := runnable.Invoke(ctx, map[string]any{
		"histories":  []*schema.Message{},
		"user_query": "我叫 morning, 邮箱是 lumworn@gmail.com。我的目标是提升实战表现，帮我制定训练计划并推荐适合的位置和打法。",
	})
	if err != nil {
		panic(err)
	}

	println("=====================思考内容====================")
	if output.ReasoningContent != "" {
		println(output.ReasoningContent)
	}
	println("=========================================")
	if output.Content != "" {
		println(output.Content)
	}
}
