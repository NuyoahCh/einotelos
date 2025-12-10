package main

import (
	"context"
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
	Name  string `json:"name"`
	Email string `json:"email"`
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

	// 1) 篮球主题：ChatTemplate
	systemTpl := `你是一名篮球教练与比赛分析师。你需要结合用户的基本信息与训练习惯，
使用 player_info 工具补全信息，然后给出适合他的训练计划/位置建议/一套简单战术建议。
注意：邮箱必须出现，用于查询信息。`

	chatTpl := prompt.FromMessages(schema.FString,
		schema.SystemMessage(systemTpl),
		schema.MessagesPlaceholder("histories", true),
		schema.UserMessage("{user_query}"),
	)

	// 2) 推荐模板（工具回来的信息 + 固定“篮球知识库/规则”）
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

	// 3) 创建 DeepSeek ChatModel（从环境变量取 key）
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		log.Fatal("缺少环境变量 DEEPSEEK_API_KEY")
	}

	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  os.Getenv("DEEPSEEK_API_KEY"), // 获取环境变量中的 APIKey 配置
		Model:   "deepseek-chat",               // 指定使用的模型名称
		BaseURL: "https://api.deepseek.com",    // 自选的 API 服务器地址
	})
	if err != nil {
		log.Fatalf("创建 DeepSeek ChatModel 失败: %v", err)
	}

	// 4) 创建工具：player_info（示例 mock：实际接你们的用户系统即可）
	playerInfoTool := utils.NewTool(
		&schema.ToolInfo{
			Name: "player_info",
			Desc: "根据用户的姓名和邮箱，查询用户的篮球相关信息（位置倾向、身体数据、打球习惯等）",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"name": {
					Type: "string",
					Desc: "用户姓名",
				},
				"email": {
					Type: "string",
					Desc: "用户邮箱",
				},
			}),
		},
		// mock 实现方式
		func(ctx context.Context, input *playerInfoRequest) (output *playerInfoResponse, err error) {
			// demo：按邮箱域名/名字随便 mock，一般你会改成查数据库/服务
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

	// 6) ToolsNode
	toolsNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
		Tools: []tool.BaseTool{playerInfoTool},
	})
	if err != nil {
		panic(err)
	}
	
	// 7) lambda：从 toolsNode 输出中提取“工具返回内容”，转成普通 user 文本（避免 role=tool）
	toolToTextOps := func(
		ctx context.Context,
		input *schema.StreamReader[[]*schema.Message],
	) (output *schema.StreamReader[*schema.Message], err error) {
		return schema.StreamReaderWithConvert(input, func(msgs []*schema.Message) (*schema.Message, error) {
			// 从 toolsNode 输出里找 role=tool 的消息，把它的 Content 摘出来
			// 注意：不同版本 schema.Message 里 role 字段名可能不同（Role/MessageType），下面按常见 Role 写
			var toolContents []string
			for _, m := range msgs {
				if m == nil {
					continue
				}
				if m.Role == "tool" { // <- 如果你编译报错，说明字段名不对，见下方“字段不一致怎么办”
					toolContents = append(toolContents, m.Content)
				}
			}

			// 没找到 tool 内容也没关系，给一个兜底文本，避免空
			text := "工具未返回有效信息。"
			if len(toolContents) > 0 {
				// 多个工具结果就拼起来
				text = "工具返回的用户信息如下：\n" + "- " + toolContents[0]
				for i := 1; i < len(toolContents); i++ {
					text += "\n- " + toolContents[i]
				}
			}

			// 这里返回一个“普通 user 消息”，后面再拼 recommendTpl
			return schema.UserMessage(text), nil
		}), nil
	}
	lambdaToolToText := compose.TransformableLambda[[]*schema.Message, *schema.Message](toolToTextOps)

	// 8) lambda：构造第二次模型的输入：system(recommendTpl) + user(工具结果文本)
	promptTransformOps := func(
		ctx context.Context,
		input *schema.StreamReader[*schema.Message],
	) (output *schema.StreamReader[[]*schema.Message], err error) {
		return schema.StreamReaderWithConvert(input, func(m *schema.Message) ([]*schema.Message, error) {
			out := make([]*schema.Message, 0, 2)
			out = append(out, schema.SystemMessage(recommendTpl))
			out = append(out, m)
			return out, nil
		}), nil
	}
	lambdaPrompt := compose.TransformableLambda[*schema.Message, []*schema.Message](promptTransformOps)

	// 9) Chain 编排：template -> chat -> tools -> lambda -> lambdaPrompt -> chat
	chain := compose.NewChain[map[string]any, *schema.Message]()
	chain.
		AppendChatTemplate(chatTpl).
		AppendChatModel(chatModel).
		AppendToolsNode(toolsNode).
		AppendLambda(lambdaToolToText).
		AppendLambda(lambdaPrompt).
		AppendChatModel(chatModel)

	// 10) 编译运行
	runnable, err := chain.Compile(ctx)
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

	// 可选：token 用量（如果 deepseek 的 response meta 有 usage）
	if output.ResponseMeta != nil && output.ResponseMeta.Usage != nil {
		println("\nToken 使用统计:")
		println("  输入 Token:", int(output.ResponseMeta.Usage.PromptTokens))
		println("  输出 Token:", int(output.ResponseMeta.Usage.CompletionTokens))
		println("  总计 Token:", int(output.ResponseMeta.Usage.TotalTokens))
	}
}
