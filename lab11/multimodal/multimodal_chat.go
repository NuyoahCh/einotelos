package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/NuyoahCh/einotelos/lab11/common"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// MultimodalChat 演示图文混合的多模态对话
// 在一次对话中可以同时包含文本、图像等多种类型的内容
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

	// 示例 1：图文结合的产品咨询
	fmt.Println("=== 示例 1: 图文结合的产品咨询 ===")
	productConsultation(ctx, chatModel)

	// 示例 2：多图文档分析
	fmt.Println("\n=== 示例 2: 多图文档分析 ===")
	documentAnalysis(ctx, chatModel)

	// 示例 3：图像编辑建议
	fmt.Println("\n=== 示例 3: 图像编辑建议 ===")
	imageEditingSuggestions(ctx, chatModel)

	// 示例 4：复杂的多模态工作流
	fmt.Println("\n=== 示例 4: 复杂的多模态工作流 ===")
	complexMultimodalWorkflow(ctx, chatModel)
}

// productConsultation 产品咨询场景：用户上传产品图片并咨询相关问题
func productConsultation(ctx context.Context, chatModel model.ChatModel) {
	productImageURL := "https://example.com/laptop.jpg"

	// 构建多模态消息
	messages := []*schema.Message{
		schema.SystemMessage("你是一个专业的产品顾问，能够根据产品图片提供详细的咨询服务。"),
		{
			Role: schema.User,
			UserInputMultiContent: []schema.MessageInputPart{
				common.NewTextPart("我想了解这款笔记本电脑："),
				common.NewImageURLPart(productImageURL),
				common.NewTextPart("请告诉我：\n1. 这是什么品牌和型号？\n2. 主要配置如何？\n3. 适合什么用途？\n4. 大概的价格区间？"),
			},
		},
	}

	response, err := chatModel.Generate(ctx, messages)
	if err != nil {
		log.Printf("产品咨询失败: %v", err)
		return
	}

	fmt.Printf("产品顾问回复:\n%s\n", response.Content)

	// 继续追问
	messages = append(messages, response)
	messages = append(messages, schema.UserMessage("这款笔记本和 MacBook Pro 相比有什么优缺点？"))

	response2, err := chatModel.Generate(ctx, messages)
	if err != nil {
		log.Printf("追问失败: %v", err)
		return
	}

	fmt.Printf("\n追问回复:\n%s\n", response2.Content)
}

// documentAnalysis 文档分析场景：分析包含多张图片的文档
func documentAnalysis(ctx context.Context, chatModel model.ChatModel) {
	page1URL := "https://example.com/doc-page1.jpg"
	page2URL := "https://example.com/doc-page2.jpg"
	page3URL := "https://example.com/doc-page3.jpg"

	messages := []*schema.Message{
		schema.SystemMessage("你是一个文档分析专家，能够理解和总结多页文档的内容。"),
		{
			Role: schema.User,
			UserInputMultiContent: []schema.MessageInputPart{
				common.NewTextPart("这是一份技术文档的三页内容，请帮我总结主要内容："),
				common.NewImageURLPart(page1URL),
				common.NewImageURLPart(page2URL),
				common.NewImageURLPart(page3URL),
				common.NewTextPart("请提供：\n1. 文档的主题\n2. 关键技术点\n3. 重要的数据或图表\n4. 结论或建议"),
			},
		},
	}

	response, err := chatModel.Generate(ctx, messages)
	if err != nil {
		log.Printf("文档分析失败: %v", err)
		return
	}

	fmt.Printf("文档分析结果:\n%s\n", response.Content)
}

// imageEditingSuggestions 图像编辑建议场景
func imageEditingSuggestions(ctx context.Context, chatModel model.ChatModel) {
	photoURL := "https://example.com/photo.jpg"

	messages := []*schema.Message{
		schema.SystemMessage("你是一个专业的摄影师和图像编辑专家。"),
		{
			Role: schema.User,
			UserInputMultiContent: []schema.MessageInputPart{
				common.NewTextPart("我拍了这张照片，但感觉不够理想："),
				common.NewImageURLPart(photoURL),
				common.NewTextPart("请给我一些改进建议：\n1. 构图方面\n2. 光线和色彩\n3. 后期处理建议\n4. 如何让照片更有吸引力"),
			},
		},
	}

	response, err := chatModel.Generate(ctx, messages)
	if err != nil {
		log.Printf("获取编辑建议失败: %v", err)
		return
	}

	fmt.Printf("图像编辑建议:\n%s\n", response.Content)
}

// complexMultimodalWorkflow 复杂的多模态工作流示例
func complexMultimodalWorkflow(ctx context.Context, chatModel model.ChatModel) {
	// 场景：设计评审工作流
	// 1. 上传设计稿
	// 2. AI 分析设计
	// 3. 提出修改建议
	// 4. 上传修改后的版本
	// 5. 对比分析

	designV1URL := "https://example.com/design-v1.jpg"
	designV2URL := "https://example.com/design-v2.jpg"

	// 第一步：初始设计评审
	messages := []*schema.Message{
		schema.SystemMessage("你是一个资深的 UI/UX 设计师，擅长评审和改进设计。"),
		{
			Role: schema.User,
			UserInputMultiContent: []schema.MessageInputPart{
				common.NewTextPart("这是我们 App 的首页设计初稿："),
				common.NewImageURLPart(designV1URL),
				common.NewTextPart("请从以下角度评审：\n1. 视觉层次\n2. 用户体验\n3. 色彩搭配\n4. 可访问性"),
			},
		},
	}

	response1, err := chatModel.Generate(ctx, messages)
	if err != nil {
		log.Printf("设计评审失败: %v", err)
		return
	}

	fmt.Printf("初始设计评审:\n%s\n", response1.Content)

	// 第二步：根据反馈修改后的版本对比
	messages = append(messages, response1)
	messages = append(messages, &schema.Message{
		Role: schema.User,
		UserInputMultiContent: []schema.MessageInputPart{
			common.NewTextPart("根据你的建议，我做了修改。这是新版本："),
			common.NewImageURLPart(designV2URL),
			common.NewTextPart("请对比两个版本，说明改进之处和还需要优化的地方。"),
		},
	})

	response2, err := chatModel.Generate(ctx, messages)
	if err != nil {
		log.Printf("版本对比失败: %v", err)
		return
	}

	fmt.Printf("\n版本对比分析:\n%s\n", response2.Content)

	// 第三步：最终建议
	messages = append(messages, response2)
	messages = append(messages, schema.UserMessage("基于当前版本，请给出最终的优化建议和实施优先级。"))

	response3, err := chatModel.Generate(ctx, messages)
	if err != nil {
		log.Printf("获取最终建议失败: %v", err)
		return
	}

	fmt.Printf("\n最终优化建议:\n%s\n", response3.Content)
}

// streamingMultimodalChat 流式多模态对话示例
func streamingMultimodalChat(ctx context.Context, chatModel model.ChatModel) {
	imageURL := "https://example.com/complex-scene.jpg"

	messages := []*schema.Message{
		schema.SystemMessage("你是一个详细的场景描述专家。"),
		{
			Role: schema.User,
			UserInputMultiContent: []schema.MessageInputPart{
				common.NewTextPart("请详细描述这个场景，包括所有细节："),
				common.NewImageURLPart(imageURL),
			},
		},
	}

	// 使用流式生成获取实时响应
	fmt.Println("开始流式生成响应...")

	stream, err := chatModel.Stream(ctx, messages)
	if err != nil {
		log.Printf("启动流式生成失败: %v", err)
		return
	}
	defer stream.Close()

	// 逐块接收和显示响应
	for {
		chunk, err := stream.Recv()
		if err != nil {
			break
		}
		fmt.Print(chunk.Content)
	}
	fmt.Println()
}
