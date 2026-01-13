// Package model 提供大模型相关组件
package model

import (
	"context"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"github.com/NuyoahCh/einotelos/einox/config"
)

// ChatModel 封装的对话模型
type ChatModel struct {
	inner  model.ChatModel
	config *config.ARKConfig
}

// NewChatModel 创建豆包对话模型
func NewChatModel(ctx context.Context, cfg *config.ARKConfig) (*ChatModel, error) {
	// 使用 ARK ChatModel
	chatModel, err := NewARKChatModel(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &ChatModel{
		inner:  chatModel,
		config: cfg,
	}, nil
}

// Generate 生成响应
func (m *ChatModel) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	return m.inner.Generate(ctx, messages, opts...)
}

// Stream 流式生成
func (m *ChatModel) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return m.inner.Stream(ctx, messages, opts...)
}

// BindTools 绑定工具
func (m *ChatModel) BindTools(tools []*schema.ToolInfo) error {
	return m.inner.BindTools(tools)
}

// Inner 获取内部模型实例
func (m *ChatModel) Inner() model.ChatModel {
	return m.inner
}

// Embedder 向量化模型
type Embedder struct {
	inner embedding.Embedder
}

// NewEmbedder 创建向量化模型
func NewEmbedder(ctx context.Context, cfg *config.ARKConfig) (*Embedder, error) {
	embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		APIKey: cfg.APIKey,
		Model:  cfg.EmbeddingModel,
	})
	if err != nil {
		return nil, err
	}

	return &Embedder{inner: embedder}, nil
}

// EmbedStrings 向量化文本
func (e *Embedder) EmbedStrings(ctx context.Context, texts []string) ([][]float64, error) {
	return e.inner.EmbedStrings(ctx, texts)
}

// Inner 获取内部实例
func (e *Embedder) Inner() embedding.Embedder {
	return e.inner
}
