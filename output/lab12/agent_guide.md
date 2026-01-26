# Lab 12 - Agent 智能体开发

## 实验目标

- 理解 Agent（智能体）的核心概念和工作原理
- 掌握 ReAct（Reasoning + Acting）模式的实现
- 学习 Agent 记忆管理的不同策略
- 实践多 Agent 协作系统的设计和开发

## 核心概念

### 1. Agent（智能体）

Agent 是能够感知环境、自主决策并采取行动以实现目标的智能系统。在 LLM 应用中，Agent 通常具备以下能力：
- **感知**：理解用户输入和环境状态
- **推理**：分析问题并制定计划
- **行动**：调用工具或执行操作
- **学习**：从经验中改进表现

### 2. ReAct 模式

ReAct（Reasoning + Acting）是一种让 LLM 交替进行推理和行动的方法：
- **Thought（思考）**：分析当前情况，决定下一步行动
- **Action（行动）**：执行具体操作（如调用工具）
- **Observation（观察）**：获取行动结果
- 循环上述过程直到找到答案

### 3. Agent 记忆

Agent 的记忆系统包括：
- **短期记忆**：对话历史，保存在上下文中
- **长期记忆**：持久化存储的信息和经验
- **语义记忆**：基于向量检索的知识库
- **工作记忆**：任务执行过程中的临时信息

### 4. 多 Agent 系统

多个专门化的 Agent 协同工作，每个 Agent 负责特定领域：
- **协调者**：分配任务和协调工作流
- **专家 Agent**：处理特定类型的任务
- **通信机制**：Agent 之间的信息传递

## 代码说明

### 示例 1：ReAct Agent（react_agent.go）

实现了 ReAct 模式的智能体，能够使用工具解决复杂问题。

```go
// 创建 ReAct Agent
agent := NewReActAgent(chatModel)

// 注册工具
agent.RegisterTool(&CalculatorTool{})
agent.RegisterTool(&SearchTool{})
agent.RegisterTool(&KnowledgeTool{})

// 解决问题
answer, err := agent.Solve(ctx, "问题描述", maxSteps)
```

**工作流程：**
1. Agent 分析问题并思考（Thought）
2. 决定使用哪个工具（Action）
3. 执行工具并获取结果（Observation）
4. 基于结果继续思考
5. 重复直到找到最终答案（Final Answer）

**关键组件：**
- `Tool` 接口：定义工具的标准接口
- `ReActAgent`：实现 ReAct 循环逻辑
- `parseResponse`：解析模型输出，提取 Action 和 Input

### 示例 2：Agent 记忆（agent_memory.go）

展示了 Agent 的多种记忆管理策略。

#### 短期记忆（对话历史）

```go
agent := NewMemoryAgent(chatModel, 10) // 保留最近 10 条消息

// 多轮对话
agent.Chat(ctx, "我叫张三")
agent.Chat(ctx, "我喜欢打篮球")
agent.Chat(ctx, "我叫什么名字？") // Agent 能记住之前的信息
```

#### 长期记忆（持久化存储）

```go
storage := NewMemoryStorage()
agent := NewPersistentMemoryAgent(chatModel, storage)

// 第一次会话
agent.Chat(ctx, "我的项目叫 EinoTelos")
agent.SaveMemory("user_001")

// 新会话加载记忆
newAgent := NewPersistentMemoryAgent(chatModel, storage)
newAgent.LoadMemory("user_001")
newAgent.Chat(ctx, "我的项目叫什么？") // 能记住之前保存的信息
```

#### 语义记忆（向量检索）

```go
agent := NewSemanticMemoryAgent(chatModel)

// 添加知识
agent.AddKnowledge(ctx, "Eino 是字节跳动开源的 LLM 框架")
agent.AddKnowledge(ctx, "RAG 是检索增强生成技术")

// 基于语义相似度查询
response, _ := agent.QueryWithMemory(ctx, "什么是 Eino？")
```

### 示例 3：多 Agent 系统（multi_agent_system.go）

实现了多个专门化 Agent 的协作系统。

```go
// 创建多 Agent 系统
system := NewMultiAgentSystem(chatModel)

// 系统包含多个专门化的 Agent：
// - Planner: 任务规划
// - Researcher: 信息研究
// - Writer: 内容创作
// - Reviewer: 质量审查
// - Developer: 代码开发

// 执行复杂任务
result, err := system.ExecuteTask(ctx, "开发一个待办事项应用")
```

**工作流程：**
1. **协调者分析任务**：理解需求，制定执行计划
2. **分配给专家 Agent**：根据任务类型选择合适的 Agent
3. **顺序执行步骤**：每个 Agent 完成自己的部分
4. **结果传递**：后续步骤可以使用前面的结果
5. **汇总输出**：协调者整合所有结果

**Agent 角色：**
- **Planner**：分析需求，制定详细计划
- **Researcher**：收集信息，进行研究分析
- **Writer**：创作内容，编写文档
- **Reviewer**：审查质量，提供改进建议
- **Developer**：设计架构，实现代码

## 运行步骤

### 1. 运行 ReAct Agent

```bash
cd lab12/react
go run react_agent.go
```

**预期输出：**
- Agent 的思考过程
- 工具调用和结果
- 最终答案

### 2. 运行记忆管理示例

```bash
cd lab12/memory
go run agent_memory.go
```

**预期输出：**
- 短期记忆：多轮对话中的信息保持
- 长期记忆：跨会话的信息持久化
- 语义记忆：基于相似度的知识检索

### 3. 运行多 Agent 系统

```bash
cd lab12/multi_agent
go run multi_agent_system.go
```

**预期输出：**
- 任务分解和执行计划
- 各个 Agent 的工作过程
- 最终的综合结果

## 常见问题

### Q1: ReAct 模式的优势是什么？

A: ReAct 模式的主要优势：
1. **可解释性**：能看到 Agent 的思考过程
2. **灵活性**：可以动态调整策略
3. **工具使用**：能够有效利用外部工具
4. **错误恢复**：可以从失败中学习并重试

### Q2: 如何设计有效的工具？

A: 设计工具的最佳实践：
1. **清晰的接口**：明确的输入输出定义
2. **详细的描述**：让 Agent 理解工具的用途
3. **错误处理**：妥善处理异常情况
4. **幂等性**：相同输入产生相同输出
5. **性能优化**：避免耗时过长的操作

### Q3: Agent 记忆应该保存多少信息？

A: 记忆管理的权衡：
- **短期记忆**：保留最近 5-20 条消息，避免上下文过长
- **长期记忆**：选择性保存重要信息，定期清理过时数据
- **语义记忆**：使用向量检索，只加载相关知识
- **成本考虑**：更多记忆意味着更多 Token 消耗

### Q4: 多 Agent 系统如何避免混乱？

A: 协调策略：
1. **明确角色**：每个 Agent 有清晰的职责
2. **协调者**：中心化的任务分配和结果汇总
3. **依赖管理**：明确步骤之间的依赖关系
4. **通信协议**：标准化的信息传递格式
5. **冲突解决**：处理 Agent 之间的意见分歧

### Q5: 如何评估 Agent 的性能？

A: 评估指标：
1. **任务完成率**：成功解决问题的比例
2. **效率**：完成任务所需的步骤数和时间
3. **准确性**：答案的正确性和质量
4. **成本**：Token 使用量和 API 调用次数
5. **用户满意度**：实际使用中的反馈

## 进阶学习

### 1. 自主 Agent

实现能够自主设定目标和执行计划的 Agent：
- 目标分解：将大目标拆分为子目标
- 计划生成：自动制定执行计划
- 自我反思：评估执行结果并调整策略

### 2. 工具学习

让 Agent 学习如何更好地使用工具：
- 工具选择：根据任务选择最合适的工具
- 参数优化：学习最佳的工具参数
- 组合使用：将多个工具组合解决复杂问题

### 3. Agent 通信协议

设计标准化的 Agent 间通信机制：
- 消息格式：定义统一的消息结构
- 协议规范：制定通信规则和约定
- 异步处理：支持异步消息传递

### 4. 实际应用场景

- **客户服务**：自动处理客户咨询和问题
- **数据分析**：自动收集、分析和报告数据
- **内容生成**：自动创作文章、报告等内容
- **代码助手**：辅助开发者编写和调试代码
- **个人助理**：管理日程、提醒事项等

## 最佳实践

### 1. Agent 设计原则

- **单一职责**：每个 Agent 专注于特定任务
- **模块化**：Agent 和工具应该可以独立开发和测试
- **可观察性**：记录 Agent 的决策过程
- **容错性**：优雅处理错误和异常情况

### 2. 提示词工程

为 Agent 设计有效的系统提示词：
```go
systemPrompt := `你是一个专业的助手，具备以下能力：
1. 分析问题并制定计划
2. 使用工具获取信息
3. 基于观察结果进行推理
4. 给出准确的最终答案

工作流程：
- 思考当前情况
- 决定下一步行动
- 执行并观察结果
- 重复直到解决问题`
```

### 3. 性能优化

- **缓存结果**：避免重复的工具调用
- **并行执行**：独立任务可以并行处理
- **早停策略**：达到目标后立即停止
- **资源限制**：设置最大步骤数和超时时间

### 4. 安全考虑

- **输入验证**：检查用户输入的合法性
- **权限控制**：限制 Agent 可以执行的操作
- **审计日志**：记录所有重要操作
- **沙箱环境**：在隔离环境中执行危险操作

## 参考资源

- [ReAct 论文](https://arxiv.org/abs/2210.03629)
- [LangChain Agent 文档](https://python.langchain.com/docs/modules/agents/)
- [AutoGPT 项目](https://github.com/Significant-Gravitas/AutoGPT)
- [Eino 官方文档](https://rcn3ahrrdvjj.feishu.cn/wiki/space/7582137522705140933)

---

**下一步学习**：完成本实验后，可以继续学习 Lab 13（生产环境部署），了解如何将 Agent 应用部署到生产环境。
