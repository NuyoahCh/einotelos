# Einox - Eino 框架教学示例

基于字节跳动 CloudWeGo Eino 框架的大模型应用示例，使用火山引擎豆包 API。

## 快速开始

### 1. 配置环境变量

```bash
# 必需：火山引擎 ARK API Key
export ARK_API_KEY="your_api_key"

# 必需：对话模型 Endpoint ID（在火山引擎控制台创建推理接入点获取）
export ARK_MODEL_NAME="ep-20250113xxxxxx-xxxxx"

# 可选：向量模型 Endpoint ID（RAG 功能需要）
export ARK_EMBEDDING_MODEL="ep-20250113yyyyyy-yyyyy"
```

> ⚠️ 注意：`ARK_MODEL_NAME` 填的是 **Endpoint ID**，不是模型名称！
> 
> 获取方式：火山引擎控制台 → 模型推理 → 推理接入点 → 创建/复制 ID

### 2. 运行 AI 学习助手

```bash
# 进入项目目录
cd einox

# 运行学习助手
go run cmd/assistant/main.go
```

### 3. 运行功能演示

```bash
go run main.go
```

---

## AI 学习助手功能

```
╔════════════════════════════════════════════╗
║      🎓 AI 学习助手 - Eino 教学示例        ║
╚════════════════════════════════════════════╝

┌────────────────────────────────────────────┐
│  1. 💬 智能对话    - 自由问答              │
│  2. 📚 知识库问答  - RAG 检索增强          │
│  3. 🔧 学习工具    - 计算/天气/时间        │
│  4. 💻 代码助手    - 代码解释与生成        │
│  5. 🌐 翻译助手    - 中英互译              │
│  6. 🚪 退出                                │
└────────────────────────────────────────────┘
```

### 功能说明

| 功能 | 说明 | Eino 组件 |
|------|------|-----------|
| 智能对话 | 流式输出的自由问答 | ChatModel + Stream |
| 知识库问答 | 基于知识库的问答 | RAG + Retriever |
| 学习工具 | 计算器、天气、时间 | Tool + Function Calling |
| 代码助手 | 代码解释和生成 | Prompt Template |
| 翻译助手 | 中英文互译 | Prompt Template |

---

## 项目结构

```
einox/
├── cmd/
│   └── assistant/        # AI 学习助手入口
│       └── main.go
├── internal/
│   └── assistant/        # 助手核心逻辑
│       └── assistant.go
├── main.go               # 功能演示入口
├── config/               # 配置管理
├── model/                # 模型组件（ARK API）
├── tools/                # 工具组件
├── callback/             # 回调组件
├── prompt/               # 提示词模板
├── rag/                  # RAG 组件
└── workflow/             # 工作流编排
```

---

## Eino 核心组件演示

本项目覆盖了 Eino 框架的核心组件：

### 1. ChatModel - 对话模型
```go
// 普通对话
response, _ := chatModel.Generate(ctx, messages)

// 流式对话
stream, _ := chatModel.Stream(ctx, messages)
```

### 2. Prompt - 提示词模板
```go
template := prompt.FromMessages(schema.FString,
    schema.SystemMessage("你是一个助手"),
    schema.MessagesPlaceholder("history", true),
    schema.UserMessage("{query}"),
)
```

### 3. Tool - 工具调用
```go
// 定义工具
tool := utils.NewTool(&schema.ToolInfo{
    Name: "calculator",
    Desc: "数学计算",
    ParamsOneOf: schema.NewParamsOneOfByParams(...),
}, handler)

// 绑定到模型
chatModel.BindTools(toolInfos)
```

### 4. RAG - 检索增强
```go
// 添加知识
rag.AddText(ctx, "知识内容", metadata)

// 检索问答
context, docs, _ := rag.Query(ctx, "问题")
```

### 5. Chain/Graph - 工作流编排
```go
// Chain 链式编排
chain := compose.NewChain[I, O]()
chain.AppendChatTemplate(tpl).AppendChatModel(model)

// Graph 图式编排
graph := compose.NewGraph[I, O]()
graph.AddChatModelNode("chat", model)
graph.AddEdge(compose.START, "chat")
```

---

## 教学建议

1. **入门**：先运行 `cmd/assistant/main.go`，体验各功能
2. **理解**：阅读 `internal/assistant/assistant.go`，了解实现
3. **实践**：修改提示词、添加新工具、扩展功能
4. **进阶**：学习 `workflow/` 下的 Chain 和 Graph 编排

---

## 参考资料

- [CloudWeGo Eino 官方文档](https://rcn3ahrrdvjj.feishu.cn/wiki/space/7582137522705140933)
- [火山引擎 ARK 文档](https://www.volcengine.com/docs/82379)
