package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

// PromptTemplate 提示词模板管理
type PromptTemplates struct{}

// 翻译助手模板
func (p *PromptTemplates) Translator(sourceLang, targetLang string) prompt.ChatTemplate {
	return prompt.FromMessages(
		schema.FString,
		schema.SystemMessage(fmt.Sprintf(
			"你是一个专业的翻译助手。请将%s翻译成%s。\n"+
				"要求：\n"+
				"1. 保持原文的语气和风格\n"+
				"2. 确保翻译准确、流畅\n"+
				"3. 只返回翻译结果，不要添加解释",
			sourceLang, targetLang,
		)),
		schema.UserMessage("{text}"),
	)
}

// 代码审查模板
func (p *PromptTemplates) CodeReviewer(language string) prompt.ChatTemplate {
	return prompt.FromMessages(
		schema.FString,
		schema.SystemMessage(fmt.Sprintf(
			"你是一个资深的%s开发专家。请审查以下代码，并提供：\n"+
				"1. 潜在的bug或问题\n"+
				"2. 性能优化建议\n"+
				"3. 代码风格改进建议\n"+
				"4. 安全性评估",
			language,
		)),
		schema.UserMessage("请审查以下代码：\n\n```{language}\n{code}\n```"),
	)
}

// 技术面试官模板
func (p *PromptTemplates) TechInterviewer(position, level string) prompt.ChatTemplate {
	return prompt.FromMessages(
		schema.FString,
		schema.SystemMessage(fmt.Sprintf(
			"你是一位%s职位的面试官，针对%s级别的候选人。\n"+
				"请根据候选人的回答：\n"+
				"1. 评估答案的准确性和深度\n"+
				"2. 提出有针对性的追问\n"+
				"3. 给出建设性的反馈",
			position, level,
		)),
		schema.UserMessage("候选人回答：{answer}\n\n请评估并追问。"),
	)
}

func main() {
	ctx := context.Background()
	templates := &PromptTemplates{}

	// 创建 ChatModel
	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey:  os.Getenv("DEEPSEEK_API_KEY"),
		Model:   "deepseek-chat",
		BaseURL: "https://api.deepseek.com",
	})
	if err != nil {
		log.Fatalf("创建模型失败: %v", err)
	}

	// 示例1: 使用翻译模板
	fmt.Println("===== 翻译示例 =====")
	translatorTemplate := templates.Translator("中文", "英文")
	messages, _ := translatorTemplate.Format(ctx, map[string]any{
		"text": "Eino 是一个强大的 AI 开发框架",
	})
	response, _ := chatModel.Generate(ctx, messages)
	fmt.Printf("翻译结果: %s\n\n", response.Content)

	// 示例2: 使用代码审查模板
	fmt.Println("===== 代码审查示例 =====")
	reviewerTemplate := templates.CodeReviewer("Go")
	messages, _ = reviewerTemplate.Format(ctx, map[string]any{
		"language": "go",
		"code": `func add(a, b int) int {
	return a + b
}`,
	})
	response, _ = chatModel.Generate(ctx, messages)
	fmt.Printf("审查结果:\n%s\n\n", response.Content)

	// 示例3: 使用面试官模板
	fmt.Println("===== 面试官示例 =====")
	interviewerTemplate := templates.TechInterviewer("Go后端开发", "中级")
	messages, _ = interviewerTemplate.Format(ctx, map[string]any{
		"answer": "goroutine 是 Go 语言的轻量级线程，由 Go 运行时管理",
	})
	response, _ = chatModel.Generate(ctx, messages)
	fmt.Printf("面试官反馈:\n%s\n\n", response.Content)
}
