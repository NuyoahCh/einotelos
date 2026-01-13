// Package server 提供 HTTP 服务
package server

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"github.com/NuyoahCh/einotelos/einox/config"
	"github.com/NuyoahCh/einotelos/einox/model"
	"github.com/NuyoahCh/einotelos/einox/rag"
	"github.com/NuyoahCh/einotelos/einox/tools"
)

//go:embed static/*
var staticFiles embed.FS

// Server HTTP 服务
type Server struct {
	config    *config.Config
	chatModel *model.ChatModel
	rag       *rag.SimpleRAG
	tools     []tool.BaseTool
}

// New 创建服务实例
func New(cfg *config.Config) (*Server, error) {
	ctx := context.Background()

	chatModel, err := model.NewChatModel(ctx, &cfg.ARK)
	if err != nil {
		return nil, fmt.Errorf("创建模型失败: %w", err)
	}

	mockEmbedder := rag.NewMockEmbedder(128)
	simpleRAG := rag.NewSimpleRAG(mockEmbedder, cfg.RAG.TopK)

	// 添加默认知识
	defaultKnowledge := []string{
		`Eino 是字节跳动开源的 Go 语言 LLM 应用开发框架。主要特点：组件化设计、工作流编排（Chain/Graph）、流式支持、回调机制。`,
		`RAG（检索增强生成）流程：文档分割 → 向量化 → 存储 → 检索 → 生成回答。`,
		`Go 语言特点：简洁、高效、原生支持并发（goroutine 和 channel）。`,
	}
	for _, k := range defaultKnowledge {
		simpleRAG.AddText(ctx, k, nil)
	}

	return &Server{
		config:    cfg,
		chatModel: chatModel,
		rag:       simpleRAG,
		tools:     tools.GetAllTools(),
	}, nil
}

// Run 启动服务
func (s *Server) Run(port int) error {
	mux := http.NewServeMux()

	// API 路由
	mux.HandleFunc("/api/chat", s.handleChat)
	mux.HandleFunc("/api/chat/stream", s.handleChatStream)
	mux.HandleFunc("/api/knowledge/query", s.handleKnowledgeQuery)
	mux.HandleFunc("/api/knowledge/add", s.handleKnowledgeAdd)
	mux.HandleFunc("/api/knowledge/clear", s.handleKnowledgeClear)
	mux.HandleFunc("/api/tools", s.handleTools)
	mux.HandleFunc("/api/code", s.handleCode)
	mux.HandleFunc("/api/translate", s.handleTranslate)

	// 静态文件
	staticFS, _ := fs.Sub(staticFiles, "static")
	mux.Handle("/", http.FileServer(http.FS(staticFS)))

	return http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}

// ChatRequest 对话请求
type ChatRequest struct {
	Message string           `json:"message"`
	History []HistoryMessage `json:"history,omitempty"`
	Mode    string           `json:"mode,omitempty"` // chat, code, translate
}

type HistoryMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse 对话响应
type ChatResponse struct {
	Content string `json:"content"`
	Error   string `json:"error,omitempty"`
}

// handleChat 处理对话请求
func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, ChatResponse{Error: "无效的请求"})
		return
	}

	ctx := r.Context()
	messages := s.buildMessages(req, "你是一个友好的 AI 学习助手，回答简洁清晰。")

	response, err := s.chatModel.Generate(ctx, messages)
	if err != nil {
		writeJSON(w, ChatResponse{Error: err.Error()})
		return
	}

	writeJSON(w, ChatResponse{Content: response.Content})
}

// handleChatStream 处理流式对话
func (s *Server) handleChatStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求", http.StatusBadRequest)
		return
	}

	// 设置 SSE 头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()

	// 根据模式选择系统提示
	systemPrompt := "你是一个友好的 AI 学习助手，回答简洁清晰。"
	if req.Mode == "code" {
		systemPrompt = "你是编程助手，代码用markdown代码块，解释简洁。"
	} else if req.Mode == "translate" {
		systemPrompt = "翻译助手：中文翻英文，英文翻中文，只输出译文。"
	}

	messages := s.buildMessages(req, systemPrompt)

	stream, err := s.chatModel.Stream(ctx, messages)
	if err != nil {
		fmt.Fprintf(w, "data: {\"error\": \"%s\"}\n\n", err.Error())
		flusher.Flush()
		return
	}
	defer stream.Close()

	for {
		chunk, err := stream.Recv()
		if err != nil {
			break
		}
		if chunk != nil && chunk.Content != "" {
			data, _ := json.Marshal(map[string]string{"content": chunk.Content})
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}

	fmt.Fprintf(w, "data: {\"done\": true}\n\n")
	flusher.Flush()
}

// handleKnowledgeQuery 知识库查询
func (s *Server) handleKnowledgeQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, ChatResponse{Error: "无效的请求"})
		return
	}

	ctx := r.Context()

	// 检索知识
	ragContext, docs, err := s.rag.Query(ctx, req.Message)
	if err != nil {
		writeJSON(w, ChatResponse{Error: err.Error()})
		return
	}

	// 构建消息
	systemPrompt := fmt.Sprintf(`基于以下资料回答问题，资料不足请说明：

%s`, ragContext)

	messages := []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(req.Message),
	}

	response, err := s.chatModel.Generate(ctx, messages)
	if err != nil {
		writeJSON(w, ChatResponse{Error: err.Error()})
		return
	}

	content := response.Content
	if len(docs) > 0 {
		content += fmt.Sprintf("\n\n📎 参考了 %d 条知识", len(docs))
	}

	writeJSON(w, ChatResponse{Content: content})
}

// KnowledgeRequest 知识请求
type KnowledgeRequest struct {
	Content string `json:"content"`
}

// handleKnowledgeAdd 添加知识
func (s *Server) handleKnowledgeAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req KnowledgeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, map[string]interface{}{"success": false, "error": "无效的请求"})
		return
	}

	if strings.TrimSpace(req.Content) == "" {
		writeJSON(w, map[string]interface{}{"success": false, "error": "内容不能为空"})
		return
	}

	ctx := r.Context()
	if err := s.rag.AddText(ctx, req.Content, nil); err != nil {
		writeJSON(w, map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	count := 0
	if store, ok := s.rag.GetStore().(*rag.MemoryVectorStore); ok {
		count = store.Count()
	}

	writeJSON(w, map[string]interface{}{"success": true, "count": count})
}

// handleKnowledgeClear 清空知识库
func (s *Server) handleKnowledgeClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.rag = rag.NewSimpleRAG(rag.NewMockEmbedder(128), s.config.RAG.TopK)
	writeJSON(w, map[string]interface{}{"success": true})
}

// ToolRequest 工具请求
type ToolRequest struct {
	Query string `json:"query"`
}

// handleTools 处理工具调用
func (s *Server) handleTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ToolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, ChatResponse{Error: "无效的请求"})
		return
	}

	ctx := r.Context()

	// 绑定工具
	toolInfos, _ := tools.GetToolInfos(ctx, s.tools)
	s.chatModel.BindTools(toolInfos)

	messages := []*schema.Message{
		schema.SystemMessage("根据用户需求使用工具：calculator(计算)、get_weather(天气)、get_current_time(时间)"),
		schema.UserMessage(req.Query),
	}

	response, err := s.chatModel.Generate(ctx, messages)
	if err != nil {
		writeJSON(w, ChatResponse{Error: err.Error()})
		return
	}

	if len(response.ToolCalls) > 0 {
		var results []string
		for _, tc := range response.ToolCalls {
			result := s.executeTool(ctx, tc.Function.Name, tc.Function.Arguments)
			results = append(results, fmt.Sprintf("[%s] %s", tc.Function.Name, result))
		}
		writeJSON(w, ChatResponse{Content: strings.Join(results, "\n")})
	} else {
		writeJSON(w, ChatResponse{Content: response.Content})
	}
}

func (s *Server) executeTool(ctx context.Context, name, args string) string {
	for _, t := range s.tools {
		info, _ := t.Info(ctx)
		if info.Name == name {
			if invokable, ok := t.(interface {
				InvokableRun(context.Context, string, ...tool.Option) (string, error)
			}); ok {
				result, err := invokable.InvokableRun(ctx, args)
				if err != nil {
					return fmt.Sprintf("错误: %v", err)
				}
				return result
			}
		}
	}
	return "工具未找到"
}

// handleCode 代码助手
func (s *Server) handleCode(w http.ResponseWriter, r *http.Request) {
	s.handleChatWithPrompt(w, r, "你是编程助手，代码用markdown代码块，解释简洁。")
}

// handleTranslate 翻译助手
func (s *Server) handleTranslate(w http.ResponseWriter, r *http.Request) {
	s.handleChatWithPrompt(w, r, "翻译助手：中文翻英文，英文翻中文，只输出译文。")
}

func (s *Server) handleChatWithPrompt(w http.ResponseWriter, r *http.Request, systemPrompt string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, ChatResponse{Error: "无效的请求"})
		return
	}

	ctx := r.Context()
	messages := []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(req.Message),
	}

	response, err := s.chatModel.Generate(ctx, messages)
	if err != nil {
		writeJSON(w, ChatResponse{Error: err.Error()})
		return
	}

	writeJSON(w, ChatResponse{Content: response.Content})
}

func (s *Server) buildMessages(req ChatRequest, systemPrompt string) []*schema.Message {
	messages := []*schema.Message{
		schema.SystemMessage(systemPrompt),
	}

	for _, h := range req.History {
		if h.Role == "user" {
			messages = append(messages, schema.UserMessage(h.Content))
		} else {
			messages = append(messages, schema.AssistantMessage(h.Content, nil))
		}
	}

	messages = append(messages, schema.UserMessage(req.Message))
	return messages
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
