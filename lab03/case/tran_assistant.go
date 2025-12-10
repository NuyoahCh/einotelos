package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/schema"
)

// Translator 基于 Deepseek 的翻译助手
type Translator struct {
	chatModel *deepseek.ChatModel
}

// TranslatorConfig 翻译器配置
type TranslatorConfig struct {
	APIKey   string
	Model    string
	BaseURL  string
	Timeout  time.Duration
	Retries  int
}

// NewTranslator 创建一个新的翻译器实例
func NewTranslator(cfg TranslatorConfig) (*Translator, error) {
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, errors.New("missing api key")
	}
	if cfg.Model == "" {
		cfg.Model = "deepseek-chat"
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.deepseek.com"
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.Retries < 0 {
		cfg.Retries = 0
	}

	// 建模时可用 Background；真正超时控制在 Translate 时做
	chatModel, err := deepseek.NewChatModel(context.Background(), &deepseek.ChatModelConfig{
		APIKey:  cfg.APIKey,
		Model:   cfg.Model,
		BaseURL: cfg.BaseURL,
	})
	if err != nil {
		return nil, err
	}

	return &Translator{chatModel: chatModel}, nil
}

// Translate 翻译文本到目标语言
func (t *Translator) Translate(ctx context.Context, text, targetLang string) (string, error) {
	text = strings.TrimSpace(text)
	targetLang = strings.TrimSpace(targetLang)
	if text == "" {
		return "", errors.New("empty text")
	}
	if targetLang == "" {
		return "", errors.New("empty target language")
	}

	// 超时控制
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 更严格的提示词：只输出译文；保留格式；不添加引号/解释
	system := fmt.Sprintf(
		"你是一个专业翻译引擎。将用户输入翻译成%s。"+
			"要求：只输出译文，不要解释；保留原有换行与列表格式；不要添加引号；不要输出多余内容。",
		targetLang,
	)

	messages := []*schema.Message{
		schema.SystemMessage(system),
		schema.UserMessage(text),
	}

	// 简单重试（网络抖动/429 等情况你可再细化错误判断）
	var lastErr error
	for attempt := 0; attempt <= 2; attempt++ { // 默认 3 次（你可配置）
		resp, err := t.chatModel.Generate(ctx, messages)
		if err == nil {
			return strings.TrimSpace(resp.Content), nil
		}
		lastErr = err

		// ctx 超时/取消就别重试
		if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(ctx.Err(), context.Canceled) {
			break
		}

		// 退避
		time.Sleep(time.Duration(attempt+1) * 300 * time.Millisecond)
	}
	return "", lastErr
}

func main() {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	translator, err := NewTranslator(TranslatorConfig{
		APIKey:  apiKey,
		Model:   "deepseek-chat",
		BaseURL: "https://api.deepseek.com",
		Timeout: 30 * time.Second,
		Retries: 2,
	})
	if err != nil {
		log.Fatalf("创建翻译器失败: %v", err)
	}

	tests := []struct {
		content string
		target  string
	}{
		{"Hello, how are you?", "中文"},
		{"Eino 是一个强大的 AI 开发框架", "English"},
		{"Les roses sont rouges", "中文"},
		{"- item1\n- item2\n", "中文"},
	}

	for _, item := range tests {
		result, err := translator.Translate(context.Background(), item.content, item.target)
		if err != nil {
			log.Printf("翻译失败: %v", err)
			continue
		}
		fmt.Printf("原文: %s\n翻译: %s\n\n", item.content, result)
	}
}
