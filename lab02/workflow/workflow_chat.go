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

	// 创建 Workflow 编排
	wf := compose.NewWorkflow[map[string]any, *schema.Message]()

	// 1) 系统提示词模板（篮球主题）
	systemTpl := `你是一名篮球教练与比赛分析师。你需要结合用户的基本信息与训练习惯，
使用 player_info API 补全用户画像，然后给出：位置建议、核心技能树、一周训练计划、以及一套简单战术建议。
注意：邮箱必须出现，用于查询信息。`

	chatTpl := prompt.FromMessages(schema.FString,
		schema.SystemMessage(systemTpl),
		schema.MessagesPlaceholder("histories", true),
		schema.UserMessage("{user_query}"),
	)

	// 2) 推荐模板
	recommendTpl := `
你是一名篮球教练与比赛分析师。请结合“工具返回的用户信息”，为用户输出建议，要求具体、可执行。

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

### C. 输出格式要求
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

	// 4) 创建工具：player_info（mock 示例）
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
			// demo：你可以替换成真实查询
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

	// 5) 绑定工具到模型（让模型能发 tool_calls）
	info, err := playerInfoTool.Info(ctx)
	if err != nil {
		panic(err)
	}
	if err := chatModel.BindTools([]*schema.ToolInfo{info}); err != nil {
		panic(err)
	}

	// 6) 工具节点
	toolsNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
		Tools: []tool.BaseTool{playerInfoTool},
	})
	if err != nil {
		panic(err)
	}

	// 7) 转换：把 toolsNode 输出的 []*schema.Message -> 提炼成一个普通 user message（避免 role=tool 进第二次模型）
	toolToTextOps := func(
		ctx context.Context,
		input *schema.StreamReader[[]*schema.Message],
	) (*schema.StreamReader[*schema.Message], error) {
		return schema.StreamReaderWithConvert(input, func(msgs []*schema.Message) (*schema.Message, error) {
			if len(msgs) == 0 {
				return nil, errors.New("no message")
			}

			// 不依赖 Role 字段：把内容统一汇总成 JSON 文本
			type lite struct {
				Content string `json:"content,omitempty"`
				// 如果你确认 schema.Message 有 Role 字段，可加上 Role: m.Role
			}
			lites := make([]lite, 0, len(msgs))
			for _, m := range msgs {
				if m == nil {
					continue
				}
				lites = append(lites, lite{Content: m.Content})
			}

			b, _ := json.MarshalIndent(lites, "", "  ")
			text := "工具返回的用户信息（汇总）：\n" + string(b)

			// 关键：返回普通 user 消息，而不是 tool 消息
			return schema.UserMessage(text), nil
		}), nil
	}
	lambdaToolToText := compose.TransformableLambda[[]*schema.Message, *schema.Message](toolToTextOps)

	// 8) 再转换：把 *schema.Message -> []*schema.Message（system recommendTpl + user(工具结果文本)）
	promptTransformOps := func(
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
	lambdaPrompt := compose.TransformableLambda[*schema.Message, []*schema.Message](promptTransformOps)

	// 9) 添加节点到 Workflow
	wf.AddChatTemplateNode("prompt", chatTpl).AddInput(compose.START)
	wf.AddChatModelNode("chat", chatModel).AddInput("prompt")
	wf.AddToolsNode("tools", toolsNode).AddInput("chat")

	// 注意：这里用新的 lambdaToolToText，替代你原来的“取第一条 message”
	wf.AddLambdaNode("tool_to_text", lambdaToolToText).AddInput("tools")

	wf.AddLambdaNode("prompt_transform", lambdaPrompt).AddInput("tool_to_text")
	wf.AddChatModelNode("chat_recommend", chatModel).AddInput("prompt_transform")
	wf.End().AddInput("chat_recommend")

	// 10) 编译 & 运行
	runnable, err := wf.Compile(ctx)
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
