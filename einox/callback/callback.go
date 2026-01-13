// Package callback 提供回调处理组件
package callback

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cloudwego/eino/schema"
)

// CallbackInput 回调输入
type CallbackInput struct {
	Messages []*schema.Message
}

// CallbackOutput 回调输出
type CallbackOutput struct {
	Message    *schema.Message
	TokenUsage *schema.TokenUsage
}

// RunInfo 运行信息
type RunInfo struct {
	Name string
}

// LoggingCallback 日志回调
type LoggingCallback struct {
	Verbose bool
}

// OnStart 开始时回调
func (c *LoggingCallback) OnStart(ctx context.Context, info *RunInfo, input *CallbackInput) context.Context {
	if c.Verbose {
		log.Printf("[Callback] 开始调用模型")
		log.Printf("  消息数量: %d", len(input.Messages))
		for i, msg := range input.Messages {
			log.Printf("  [%d] %s: %s", i, msg.Role, truncate(msg.Content, 50))
		}
	}
	return context.WithValue(ctx, "start_time", time.Now())
}

// OnEnd 结束时回调
func (c *LoggingCallback) OnEnd(ctx context.Context, info *RunInfo, output *CallbackOutput) context.Context {
	if c.Verbose {
		startTime, _ := ctx.Value("start_time").(time.Time)
		duration := time.Since(startTime)

		log.Printf("[Callback] 模型调用完成")
		log.Printf("  耗时: %v", duration)
		if output.Message != nil {
			log.Printf("  响应: %s", truncate(output.Message.Content, 100))
		}
		if output.TokenUsage != nil {
			log.Printf("  Token: prompt=%d, completion=%d, total=%d",
				output.TokenUsage.PromptTokens,
				output.TokenUsage.CompletionTokens,
				output.TokenUsage.TotalTokens)
		}
	}
	return ctx
}

// OnError 错误时回调
func (c *LoggingCallback) OnError(ctx context.Context, info *RunInfo, err error) context.Context {
	log.Printf("[Callback] 模型调用错误: %v", err)
	return ctx
}

// MetricsCallback 指标收集回调
type MetricsCallback struct {
	TotalCalls     int
	TotalTokens    int
	TotalLatencyMs int64
	SuccessCount   int
	ErrorCount     int
}

func (c *MetricsCallback) OnStart(ctx context.Context, info *RunInfo, input *CallbackInput) context.Context {
	c.TotalCalls++
	return context.WithValue(ctx, "metrics_start", time.Now())
}

func (c *MetricsCallback) OnEnd(ctx context.Context, info *RunInfo, output *CallbackOutput) context.Context {
	c.SuccessCount++
	if startTime, ok := ctx.Value("metrics_start").(time.Time); ok {
		c.TotalLatencyMs += time.Since(startTime).Milliseconds()
	}
	if output.TokenUsage != nil {
		c.TotalTokens += output.TokenUsage.TotalTokens
	}
	return ctx
}

func (c *MetricsCallback) OnError(ctx context.Context, info *RunInfo, err error) context.Context {
	c.ErrorCount++
	return ctx
}

func (c *MetricsCallback) Report() string {
	avgLatency := float64(0)
	if c.TotalCalls > 0 {
		avgLatency = float64(c.TotalLatencyMs) / float64(c.TotalCalls)
	}
	return fmt.Sprintf(
		"调用统计:\n  总调用: %d\n  成功: %d\n  失败: %d\n  总Token: %d\n  平均延迟: %.2fms",
		c.TotalCalls, c.SuccessCount, c.ErrorCount, c.TotalTokens, avgLatency,
	)
}

// NewModelCallbackHandler 创建模型回调处理器
func NewModelCallbackHandler(verbose bool) *LoggingCallback {
	return &LoggingCallback{Verbose: verbose}
}

// StreamCallback 流式回调
type StreamCallback struct {
	OnChunk func(content string)
	OnDone  func(fullContent string)
}

// CollectStream 收集流式响应
func CollectStream(stream *schema.StreamReader[*schema.Message], cb *StreamCallback) (string, error) {
	var fullContent string

	for {
		chunk, err := stream.Recv()
		if err != nil {
			break
		}
		if chunk != nil && chunk.Content != "" {
			fullContent += chunk.Content
			if cb != nil && cb.OnChunk != nil {
				cb.OnChunk(chunk.Content)
			}
		}
	}

	if cb != nil && cb.OnDone != nil {
		cb.OnDone(fullContent)
	}

	return fullContent, nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
