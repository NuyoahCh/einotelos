# AI 学习助手 - Eino 框架教学示例

基于字节跳动 Eino 框架 + 豆包大模型的 Web 应用。

## 快速开始

### 1. 配置环境变量

```bash
export ARK_API_KEY="your_api_key"
export ARK_MODEL_NAME="ep-20250113xxxxxx-xxxxx"  # Endpoint ID
```

### 2. 运行

```bash
# 从项目根目录
go run ./einox/main.go

# 或指定端口
go run ./einox/main.go -port 3000
```

### 3. 访问

打开浏览器访问 http://localhost:8080

## 功能

| 功能 | 说明 | Eino 组件 |
|------|------|-----------|
| 💬 智能对话 | 流式问答，多轮对话 | ChatModel + Stream |
| 📚 知识库问答 | RAG 检索增强 | RAG + Retriever |
| 🔧 学习工具 | 计算/天气/时间 | Tool + Function Calling |
| 💻 代码助手 | 代码解释与生成 | Prompt Template |
| 🌐 翻译助手 | 中英互译 | Prompt Template |

## 项目结构

```
einox/
├── main.go              # 入口
├── config/              # 配置
├── model/               # 模型组件
├── server/              # HTTP 服务
│   ├── server.go        # API 路由
│   └── static/          # 前端文件
│       ├── index.html
│       ├── style.css
│       └── app.js
├── tools/               # 工具组件
├── rag/                 # RAG 组件
├── callback/            # 回调组件
├── prompt/              # 提示词模板
└── workflow/            # 工作流编排
```

## API 接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/chat` | POST | 普通对话 |
| `/api/chat/stream` | POST | 流式对话 (SSE) |
| `/api/knowledge/query` | POST | 知识库问答 |
| `/api/knowledge/add` | POST | 添加知识 |
| `/api/knowledge/clear` | POST | 清空知识库 |
| `/api/tools` | POST | 工具调用 |
| `/api/code` | POST | 代码助手 |
| `/api/translate` | POST | 翻译 |

## 技术栈

- 后端: Go + Eino 框架
- 前端: 原生 HTML/CSS/JS
- 模型: 火山引擎豆包
- 通信: REST API + SSE 流式
