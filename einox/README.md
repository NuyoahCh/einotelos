# Einox - Eino 框架全方位应用示例

基于字节跳动 CloudWeGo Eino 框架的大模型应用示例，使用火山引擎豆包 API。

## 功能特性

本项目整合了 Eino 框架的核心组件：

- **ChatModel** - 对话模型，支持普通对话和流式对话
- **Prompt** - 提示词模板，支持变量替换和消息占位符
- **Tool** - 工具调用，让 LLM 能够调用外部功能
- **Callback** - 回调机制，用于日志记录和指标收集
- **Lambda** - 自定义函数节点，用于数据转换
- **Chain** - 链式编排，适合简单顺序流程
- **Graph** - 图式编排，支持复杂分支和并行
- **RAG** - 检索增强生成，基于知识库问答

## 项目结构

```
einox/
├── main.go           # 应用入口
├── config/           # 配置管理
│   └── config.go
├── model/            # 模型组件
│   ├── chat.go       # ChatModel 封装
│   └── ark.go        # ARK API 实现
├── tools/            # 工具组件
│   └── tools.go      # 计算器、天气、时间等工具
├── callback/         # 回调组件
│   └── callback.go   # 日志和指标回调
├── prompt/           # 提示词模板
│   └── prompt.go     # 预定义模板
├── rag/              # RAG 组件
│   ├── document.go   # 文档处理
│   ├── vectorstore.go # 向量存储
│   └── retriever.go  # 检索器
├── workflow/         # 工作流编排
│   ├── chain.go      # Chain 编排
│   └── graph.go      # Graph 编排
└── app/              # 应用主体
    └── app.go        # 演示功能
```

## 环境配置

### 必需环境变量

```bash
# 火山引擎 ARK API Key
export ARK_API_KEY="your_ark_api_key"

# 对话模型名称（可选，默认 doubao-pro-32k）
export ARK_MODEL_NAME="your_model_endpoint_id"

# 向量模型名称（可选，用于 RAG）
export ARK_EMBEDDING_MODEL="your_embedding_model_endpoint_id"
```

### 可选环境变量

```bash
# ARK API 地址（默认北京区域）
export ARK_BASE_URL="https://ark.cn-beijing.volces.com/api/v3"

# 向量数据库类型（默认 memory）
export VECTOR_DB_TYPE="memory"
```

## 运行方式

### 运行演示

```bash
# 进入项目目录
cd einox

# 运行所有演示
go run main.go

# 运行交互模式
go run main.go -i
```

### 编译运行

```bash
# 编译
go build -o einox-app ./einox

# 运行
./einox-app
```

## 演示内容

运行后会依次演示以下功能：

1. **基础对话** - 简单的问答交互
2. **流式对话** - 打字机效果的流式输出
3. **工具调用** - 计算器、天气查询、时间获取
4. **Chain 编排** - 链式调用代码助手
5. **Graph 工作流** - 图式编排工具调用
6. **RAG 检索增强** - 基于知识库的问答

## 组件说明

### 工具列表

| 工具名 | 功能 | 参数 |
|--------|------|------|
| calculator | 数学计算 | expression: 表达式 |
| get_weather | 天气查询 | city: 城市名 |
| get_current_time | 获取时间 | format: 格式, timezone: 时区 |
| web_search | 网络搜索 | query: 关键词, limit: 数量 |

### 提示词模板

- `GeneralAssistant` - 通用助手
- `CodeAssistant` - 代码助手
- `TranslationAssistant` - 翻译助手
- `RAGAssistant` - RAG 问答助手
- `ToolAssistant` - 工具调用助手

## 扩展开发

### 添加新工具

```go
// 在 tools/tools.go 中添加
type MyTool struct{}

func (t *MyTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
    return &schema.ToolInfo{
        Name: "my_tool",
        Desc: "工具描述",
        ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
            "param1": {Type: "string", Desc: "参数描述", Required: true},
        }),
    }, nil
}

func (t *MyTool) InvokableRun(ctx context.Context, args string, opts ...tool.Option) (string, error) {
    // 实现逻辑
    return "result", nil
}
```

### 自定义提示词

```go
// 在 prompt/prompt.go 中添加
Templates.MyTemplate = prompt.FromMessages(schema.FString,
    schema.SystemMessage("你的系统提示词"),
    schema.MessagesPlaceholder("history", true),
    schema.UserMessage("{query}"),
)
```

### 构建自定义工作流

```go
// 使用 ChainBuilder
builder := workflow.NewChainBuilder(ctx, chatModel).
    WithTemplate(myTemplate).
    WithTools(myTools)

chain, _ := builder.BuildToolChain()
result, _ := workflow.RunChain(ctx, chain, "用户问题", history)
```

## 注意事项

1. 确保已正确配置火山引擎 ARK API Key
2. 模型名称需要使用火山引擎控制台中的 Endpoint ID
3. RAG 功能默认使用内存向量存储，生产环境建议使用 Redis 或 VikingDB
4. 工具调用需要模型支持 Function Calling 功能

## 参考资料

- [CloudWeGo Eino 官方文档](https://rcn3ahrrdvjj.feishu.cn/wiki/space/7582137522705140933)
- [火山引擎 ARK 文档](https://www.volcengine.com/docs/82379)
- [豆包大模型](https://www.volcengine.com/product/doubao)
