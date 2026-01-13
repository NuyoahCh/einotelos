// Package model 提供 ARK 模型实现
package model

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"github.com/NuyoahCh/einotelos/einox/config"
)

// ARKChatModel 火山引擎 ARK 对话模型
type ARKChatModel struct {
	apiKey    string
	model     string
	baseURL   string
	client    *http.Client
	toolInfos []*schema.ToolInfo
}

// NewARKChatModel 创建 ARK 对话模型
func NewARKChatModel(ctx context.Context, cfg *config.ARKConfig) (*ARKChatModel, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("ARK API Key 不能为空")
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://ark.cn-beijing.volces.com/api/v3"
	}

	return &ARKChatModel{
		apiKey:  cfg.APIKey,
		model:   cfg.ChatModel,
		baseURL: baseURL,
		client:  &http.Client{},
	}, nil
}

// arkMessage ARK API 消息格式
type arkMessage struct {
	Role       string        `json:"role"`
	Content    string        `json:"content,omitempty"`
	ToolCalls  []arkToolCall `json:"tool_calls,omitempty"`
	ToolCallID string        `json:"tool_call_id,omitempty"`
}

type arkToolCall struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"`
	Function arkFunctionCall `json:"function"`
}

type arkFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type arkTool struct {
	Type     string      `json:"type"`
	Function arkFunction `json:"function"`
}

type arkFunction struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

type arkRequest struct {
	Model    string       `json:"model"`
	Messages []arkMessage `json:"messages"`
	Tools    []arkTool    `json:"tools,omitempty"`
	Stream   bool         `json:"stream"`
}

type arkResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role      string        `json:"role"`
			Content   string        `json:"content"`
			ToolCalls []arkToolCall `json:"tool_calls,omitempty"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type arkStreamResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role      string        `json:"role"`
			Content   string        `json:"content"`
			ToolCalls []arkToolCall `json:"tool_calls,omitempty"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// Generate 生成响应
func (m *ARKChatModel) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	// 转换消息格式
	arkMsgs := make([]arkMessage, 0, len(messages))
	for _, msg := range messages {
		arkMsg := arkMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		}
		if msg.ToolCallID != "" {
			arkMsg.ToolCallID = msg.ToolCallID
		}
		arkMsgs = append(arkMsgs, arkMsg)
	}

	// 构建请求
	req := arkRequest{
		Model:    m.model,
		Messages: arkMsgs,
		Stream:   false,
	}

	// 添加工具
	if len(m.toolInfos) > 0 {
		req.Tools = m.convertTools()
	}

	// 发送请求
	resp, err := m.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New("ARK 返回空响应")
	}

	choice := resp.Choices[0]
	result := &schema.Message{
		Role:    schema.RoleType(choice.Message.Role),
		Content: choice.Message.Content,
		ResponseMeta: &schema.ResponseMeta{
			Usage: &schema.TokenUsage{
				PromptTokens:     resp.Usage.PromptTokens,
				CompletionTokens: resp.Usage.CompletionTokens,
				TotalTokens:      resp.Usage.TotalTokens,
			},
		},
	}

	// 处理工具调用
	if len(choice.Message.ToolCalls) > 0 {
		result.ToolCalls = make([]schema.ToolCall, 0, len(choice.Message.ToolCalls))
		for _, tc := range choice.Message.ToolCalls {
			result.ToolCalls = append(result.ToolCalls, schema.ToolCall{
				ID:   tc.ID,
				Type: tc.Type,
				Function: schema.FunctionCall{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			})
		}
	}

	return result, nil
}

// Stream 流式生成
func (m *ARKChatModel) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	// 转换消息格式
	arkMsgs := make([]arkMessage, 0, len(messages))
	for _, msg := range messages {
		arkMsgs = append(arkMsgs, arkMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		})
	}

	// 构建请求
	req := arkRequest{
		Model:    m.model,
		Messages: arkMsgs,
		Stream:   true,
	}

	// 创建流
	sr, sw := schema.Pipe[*schema.Message](10)

	go func() {
		defer sw.Close()
		if err := m.doStreamRequest(ctx, req, sw); err != nil {
			sw.Send(nil, err)
		}
	}()

	return sr, nil
}

// BindTools 绑定工具
func (m *ARKChatModel) BindTools(tools []*schema.ToolInfo) error {
	m.toolInfos = tools
	return nil
}

func (m *ARKChatModel) convertTools() []arkTool {
	tools := make([]arkTool, 0, len(m.toolInfos))
	for _, info := range m.toolInfos {
		// 构建参数 schema
		params := map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
			"required":   []string{},
		}

		tools = append(tools, arkTool{
			Type: "function",
			Function: arkFunction{
				Name:        info.Name,
				Description: info.Desc,
				Parameters:  params,
			},
		})
	}
	return tools
}

func (m *ARKChatModel) doRequest(ctx context.Context, req arkRequest) (*arkResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", m.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+m.apiKey)

	httpResp, err := m.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("ARK API 错误 [%d]: %s", httpResp.StatusCode, string(respBody))
	}

	var resp arkResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}

func (m *ARKChatModel) doStreamRequest(ctx context.Context, req arkRequest, sw *schema.StreamWriter[*schema.Message]) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", m.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+m.apiKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	httpResp, err := m.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(httpResp.Body)
		return fmt.Errorf("ARK API 错误 [%d]: %s", httpResp.StatusCode, string(respBody))
	}

	// 读取 SSE 流
	reader := httpResp.Body
	buf := make([]byte, 4096)
	var leftover string

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		data := leftover + string(buf[:n])
		lines := strings.Split(data, "\n")

		// 保留最后一个不完整的行
		leftover = lines[len(lines)-1]
		lines = lines[:len(lines)-1]

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || line == "data: [DONE]" {
				continue
			}

			if strings.HasPrefix(line, "data: ") {
				jsonData := strings.TrimPrefix(line, "data: ")
				var streamResp arkStreamResponse
				if err := json.Unmarshal([]byte(jsonData), &streamResp); err != nil {
					continue
				}

				if len(streamResp.Choices) > 0 {
					delta := streamResp.Choices[0].Delta
					msg := &schema.Message{
						Role:    schema.RoleType(delta.Role),
						Content: delta.Content,
					}
					if sw.Send(msg, nil) {
						return nil
					}
				}
			}
		}
	}
}
