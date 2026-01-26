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

// VisionQA 演示视觉问答（Visual Question Answering）功能
// 用户可以针对图像提出具体问题，模型会基于图像内容给出答案
func main() {
	ctx := context.Background()

	// 创建支持视觉的 ChatModel
	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  os.Getenv("DEEPSEEK_API_KEY"),
		Model:   "deepseek-chat",
		BaseURL: "https://api.deepseek.com",
	})
	if err != nil {
		log.Fatalf("创建 ChatModel 实例失败: %v", err)
	}

	// 示例图像 URL（请替换为实际图像）
	imageURL := "https://example.com/product-image.jpg"

	// 场景 1：产品识别和描述
	fmt.Println("=== 场景 1: 产品识别 ===")
	productQuestions := []string{
		"这是什么产品？",
		"这个产品的主要特点是什么？",
		"产品的颜色和材质看起来如何？",
		"这个产品适合什么场景使用？",
	}
	askQuestionsAboutImage(ctx, chatModel, imageURL, productQuestions, "你是一个产品分析专家。")

	// 场景 2：场景理解
	fmt.Println("\n=== 场景 2: 场景理解 ===")
	sceneURL := "https://example.com/street-scene.jpg"
	sceneQuestions := []string{
		"这是什么地方？",
		"图片中有多少人？",
		"天气情况如何？",
		"这个场景可能是在什么时间拍摄的？",
	}
	askQuestionsAboutImage(ctx, chatModel, sceneURL, sceneQuestions, "你是一个场景分析专家。")

	// 场景 3：文字识别（OCR）
	fmt.Println("\n=== 场景 3: 图像中的文字识别 ===")
	textImageURL := "https://example.com/text-image.jpg"
	ocrQuestions := []string{
		"图片中有哪些文字？请全部列出。",
		"文字的语言是什么？",
		"文字内容的主题是什么？",
	}
	askQuestionsAboutImage(ctx, chatModel, textImageURL, ocrQuestions, "你是一个 OCR 专家，擅长识别图像中的文字。")

	// 场景 4：比较多张图片
	fmt.Println("\n=== 场景 4: 多图比较 ===")
	compareImages(ctx, chatModel)
}

// askQuestionsAboutImage 针对单张图像提出多个问题
func askQuestionsAboutImage(ctx context.Context, chatModel model.ChatModel, imageURL string, questions []string, systemPrompt string) {
	for i, question := range questions {
		messages := []*schema.Message{
			schema.SystemMessage(systemPrompt),
			{
				Role: schema.User,
				UserInputMultiContent: []schema.MessageInputPart{
					common.NewTextPart(question),
					common.NewImageURLPart(imageURL),
				},
			},
		}

		response, err := chatModel.Generate(ctx, messages)
		if err != nil {
			log.Printf("问题 %d 处理失败: %v", i+1, err)
			continue
		}

		fmt.Printf("\nQ%d: %s\n", i+1, question)
		fmt.Printf("A%d: %s\n", i+1, response.Content)
	}
}

// compareImages 比较多张图片的内容
func compareImages(ctx context.Context, chatModel model.ChatModel) {
	image1URL := "https://example.com/image1.jpg"
	image2URL := "https://example.com/image2.jpg"

	messages := []*schema.Message{
		schema.SystemMessage("你是一个图像比较专家，能够分析和比较多张图片的异同。"),
		{
			Role: schema.User,
			UserInputMultiContent: []schema.MessageInputPart{
				common.NewTextPart("请比较这两张图片，说明它们的相似之处和不同之处："),
				common.NewImageURLPart(image1URL),
				common.NewImageURLPart(image2URL),
			},
		},
	}

	response, err := chatModel.Generate(ctx, messages)
	if err != nil {
		log.Printf("图像比较失败: %v", err)
		return
	}

	fmt.Printf("\n图像比较结果:\n%s\n", response.Content)
}

// 使用本地图像文件的示例（需要转换为 Base64）
func useLocalImage(ctx context.Context, chatModel model.ChatModel, imagePath string) {
	// 读取本地图像文件
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		log.Printf("读取图像文件失败: %v", err)
		return
	}

	// 将图像转换为 Base64 编码
	// 注意：实际使用时需要添加适当的 MIME 类型前缀
	// 例如：data:image/jpeg;base64,<base64_data>
	base64Image := fmt.Sprintf("data:image/jpeg;base64,%s", imageData)

	messages := []*schema.Message{
		schema.SystemMessage("你是一个图像分析助手。"),
		{
			Role: schema.User,
			UserInputMultiContent: []schema.MessageInputPart{
				common.NewTextPart("请描述这张图片。"),
				common.NewImageURLPart(base64Image),
			},
		},
	}

	response, err := chatModel.Generate(ctx, messages)
	if err != nil {
		log.Printf("处理本地图像失败: %v", err)
		return
	}

	fmt.Printf("本地图像分析结果:\n%s\n", response.Content)
}
