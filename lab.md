# 实验环境配置指南

## 实验环境

本项目是基于 CloudWeGo Eino 框架的大语言模型应用开发教学项目，包含多个实验模块，涵盖从基础聊天到高级 RAG 应用的完整学习路径。

### 开发环境

#### 1. 访问项目地址

- GitHub 仓库: [https://github.com/NuyoahCh/einotelos](https://github.com/NuyoahCh/einotelos)
- CloudWeGo Eino 官方文档: [https://rcn3ahrrdvjj.feishu.cn/wiki/space/7582137522705140933](https://rcn3ahrrdvjj.feishu.cn/wiki/space/7582137522705140933)

#### 2. 安装 Golang SDK

- **版本要求**: Go 1.18+（推荐使用 Go 1.24.10）
- **下载地址**: [https://go.dev/dl/](https://go.dev/dl/)
- **安装验证**:
  ```bash
  go version
  # 输出示例: go version go1.24.10 darwin/amd64
  ```

#### 3. 推荐使用 IDE

- **Visual Studio Code** (推荐)
  - 安装 Go 扩展: `Go` by Google
  - 支持代码高亮、自动补全、调试等功能
- **GoLand** by JetBrains
- 其他支持 Go 的代码编辑器

#### 4. 推荐使用 OS

- **Linux** (推荐)
- **macOS** 
- **Windows** (需安装 Git Bash 或 WSL)

#### 5. Fork 代码仓库

- 访问项目主页，点击右上角 Fork 按钮
- 将项目克隆到自己的账户下，方便自由完成构建和实验

#### 6. Clone 代码到本地（**位置务必为**）

```bash
# 克隆到指定路径
cd $GOPATH/src/github.com/{your_username}/
git clone https://github.com/{your_username}/einotelos.git
cd einotelos
```

如果没有设置 `$GOPATH`，可以使用默认路径：
```bash
cd ~/go/src/github.com/{your_username}/
```

#### 7. 安装项目依赖

```bash
# 下载并安装项目依赖包
go mod download
go mod tidy
```

#### 8. 配置环境变量

本项目需要配置 API Key 才能正常运行。支持多种 LLM 服务商：

##### DeepSeek（推荐入门使用）
```bash
export DEEPSEEK_API_KEY="your_deepseek_api_key"
```
- 注册地址: [https://platform.deepseek.com/](https://platform.deepseek.com/)
- 新用户赠送免费额度

##### 火山引擎 ARK（推荐使用）
```bash
export ARK_API_KEY="your_ark_api_key"
export ARK_MODEL_NAME="your_model_name"
```
- 注册地址: 火山引擎控制台

##### Ollama（本地部署，可选）
```bash
# 如果使用本地 Ollama 服务
# 下载安装: https://ollama.ai/
ollama serve
```

**环境变量持久化（推荐）**

为避免每次重启终端都要重新设置，建议将环境变量添加到配置文件：

```bash
# macOS/Linux (使用 zsh)
echo 'export DEEPSEEK_API_KEY="your_deepseek_api_key"' >> ~/.zshrc
source ~/.zshrc

# macOS/Linux (使用 bash)
echo 'export DEEPSEEK_API_KEY="your_deepseek_api_key"' >> ~/.bashrc
source ~/.bashrc
```

---

## 目录组织

### 核心模块

- **lab01/** - 聊天快速入门，演示最基础的对话功能
- **lab02/** - 工作流与链式调用
  - `chain/` - 链式调用模式
  - `graph/` - 图式工作流
  - `workflow/` - 工作流编排
- **lab03/** - 生成配置与错误处理
  - `generate/` - 单次、多次、流式生成
  - `callback/` - 回调函数配置
  - `error/` - 错误处理机制
  - `case/` - 翻译助手实战案例
- **lab04/** - 提示词工程
  - `complex/` - 复杂逻辑提示词
  - `multi/` - 多类型消息
  - `multiplex/` - 模型多路复用
  - `replace/` - 变量替换
- **lab05/** - 文档加载与解析
  - `loader/` - 文档加载器（本地、URL、S3）
  - `parser/` - 文档解析器（PDF、HTML、Text）
  - `transformer/` - 文档切分器（Markdown、递归、语义）
  - `extparser/` - 扩展解析器
  - `case/` - 文档处理案例
- **lab06/** - 向量化与检索
  - `text/` - 文本向量化
  - `document/` - 文档向量化
  - `memory/` - 记忆向量化
  - `case/` - 文档问答案例
- **lab07/** - Lambda 函数
  - `basic/` - 基础 Lambda 功能
  - `option/` - Lambda 配置选项
  - `stream/` - 流式 Lambda
- **lab08/** - 索引器
  - `alone/` - 独立索引器
  - `arrange/` - 编排索引器
- **lab09/** - 检索增强生成（RAG）
  - `retrieval_augment.go` - RAG 完整实现
- **lab10/** - 工具调用
  - `interface/` - 计算器等接口工具
  - `weather/` - 天气查询工具
  - `db/` - 数据库工具
  - `file/` - 文件操作工具
  - `newtool/` - 自定义工具
  - `case/` - 工具集成案例

### 辅助目录

- **einox/** - 实战项目主入口（待实现）
- **output/** - 各实验的输出结果和文档
- **go.mod** - Go 模块依赖配置
- **LICENSE** - 开源许可证
- **README.md** - 项目说明文档

---

## 实验步骤

### 1. 跟着设计文档已完成一遍

建议按照以下顺序学习各个实验模块：

1. **Lab 01 - 快速入门**: 理解基本的 ChatModel 使用方式
2. **Lab 02 - 工作流**: 学习如何编排复杂的对话流程
3. **Lab 03 - 生成配置**: 掌握不同的生成模式和错误处理
4. **Lab 04 - 提示词工程**: 学习高效的提示词设计
5. **Lab 05 - 文档处理**: 了解文档加载、解析、切分的完整流程
6. **Lab 06 - 向量化**: 掌握向量嵌入和相似度检索
7. **Lab 07 - Lambda**: 学习函数式编程在 LLM 中的应用
8. **Lab 08 - 索引器**: 构建高效的文档索引系统
9. **Lab 09 - RAG**: 实现完整的检索增强生成应用
10. **Lab 10 - 工具调用**: 让 LLM 能够调用外部工具

### 2. 运行实验代码

每个实验都可以独立运行。以 Lab 01 为例：

```bash
# 进入项目根目录
cd /path/to/einotelos

# 运行 Lab 01 快速入门
cd lab01
go run chat_quickstart.go
```

**常见运行方式：**

```bash
# 方式一：直接运行单个文件
go run lab01/chat_quickstart.go

# 方式二：进入目录后运行
cd lab01
go run chat_quickstart.go

# 方式三：构建后运行（推荐生产环境）
go build -o bin/chat_quickstart lab01/chat_quickstart.go
./bin/chat_quickstart
```

### 3. 修改参数进行调试

在运行过程中，可以：

- 修改提示词内容，观察输出变化
- 调整模型参数（temperature、max_tokens 等）
- 切换不同的模型进行对比测试
- 添加日志输出，跟踪程序执行流程

### 4. 最后自己完全手写实现一遍

为了深入理解框架原理：

- 不看原有代码，根据需求从零开始编写
- 实现过程中遇到问题，查阅官方文档
- 对比自己的实现与示例代码的差异
- 思考不同实现方式的优劣

### 5. 完成后查看 output 目录的文档

每个实验的 `output/` 目录下都有对应的 Markdown 文档：

- 包含实验的详细说明
- 代码的关键知识点解析
- 常见问题和解决方案
- 进阶学习建议

---

## 常见问题排查

### 依赖安装失败

```bash
# 清理缓存后重新安装
go clean -modcache
go mod download
```

### 运行时提示 API Key 错误

```bash
# 检查环境变量是否设置
echo $DEEPSEEK_API_KEY

# 如果为空，重新设置
export DEEPSEEK_API_KEY="your_api_key"
```

### 网络连接问题

```bash
# 配置 Go 模块代理（中国大陆用户推荐）
go env -w GOPROXY=https://goproxy.cn,direct
```

### 代码提示和补全不生效

在 VS Code 中：
1. 安装 Go 扩展
2. 执行命令：`Go: Install/Update Tools`
3. 重启 VS Code

---

## 进阶学习资源

- **CloudWeGo Eino 官方文档**: [https://rcn3ahrrdvjj.feishu.cn/wiki/space/7582137522705140933](https://rcn3ahrrdvjj.feishu.cn/wiki/space/7582137522705140933)
- **Go 语言官方文档**: [https://go.dev/doc/](https://go.dev/doc/)
- **LangChain 概念参考**: [https://python.langchain.com/](https://python.langchain.com/)

---

## 贡献指南

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

---

## 联系方式

如有问题，欢迎通过以下方式联系：

- 提交 GitHub Issue
- 参考官方文档寻求帮助
- 加入 CloudWeGo 社区讨论

祝你学习愉快！🚀
