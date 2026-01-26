# Lab 11 - 多模态应用开发

## 实验目标

- 学习如何使用 Eino 框架处理多模态输入（文本 + 图像）
- 掌握图像理解和视觉问答（VQA）的实现方法
- 理解图文混合对话的应用场景
- 实践多模态工作流的设计和实现

## 核心概念

### 1. 多模态 AI

多模态 AI 是指能够处理和理解多种类型数据（如文本、图像、音频、视频）的人工智能系统。在本实验中，我们主要关注文本和图像的结合。

### 2. 视觉问答（VQA）

视觉问答是一种多模态任务，系统需要根据给定的图像回答用户提出的问题。这要求模型同时理解图像内容和文本问题。

### 3. 图像理解

图像理解是指 AI 模型能够识别图像中的对象、场景、文字、颜色等信息，并用自然语言描述图像内容。

### 4. ChatMessagePart

在 Eino 框架中，`ChatMessagePart` 用于构建包含多种类型内容的消息：
- `TextChatMessagePart`: 文本内容
- `ImageChatMessagePart`: 图像内容（URL 或 Base64）

## 代码说明

### 示例 1：图像理解（image_understanding.go）

这个示例展示了如何让 AI 理解和描述图像内容。

```go
// 创建包含图像的消息
messages := []*schema.Message{
    schema.SystemMessage("你是一个专业的图像分析助手。"),
    {
        Role: schema.User,
        Content: []schema.ChatMessagePart{
            schema.NewTextChatMessagePart("请详细描述这张图片的内容。"),
            schema.NewImageChatMessagePart(imageURL),
        },
    },
}

// 调用模型进行图像理解
response, err := chatModel.Generate(ctx, messages)
```

**关键点：**
- 使用 `schema.NewImageChatMessagePart()` 添加图像
- 图像可以通过 URL 或 Base64 编码提供
- 支持在多轮对话中引用图像

### 示例 2：图文混合对话（multimodal_chat.go）

展示了如何在一次对话中同时包含多个文本和图像部分。

```go
messages := []*schema.Message{
    {
        Role: schema.User,
        Content: []schema.ChatMessagePart{
            schema.NewTextChatMessagePart("我想了解这款产品："),
            schema.NewImageChatMessagePart(productImageURL),
            schema.NewTextChatMessagePart("请告诉我配置和价格。"),
        },
    },
}
```

**应用场景：**
- 产品咨询：上传产品图片并询问详情
- 文档分析：分析包含多张图片的文档
- 设计评审：对比不同版本的设计稿
- 图像编辑建议：获取照片改进建议

### 示例 3：视觉问答（vision_qa.go）

针对图像提出具体问题并获得答案。

```go
// 针对同一张图片提出多个问题
questions := []string{
    "这是什么产品？",
    "产品的颜色是什么？",
    "这个产品适合什么场景使用？",
}

for _, question := range questions {
    messages := []*schema.Message{
        {
            Role: schema.User,
            Content: []schema.ChatMessagePart{
                schema.NewTextChatMessagePart(question),
                schema.NewImageChatMessagePart(imageURL),
            },
        },
    }
    response, _ := chatModel.Generate(ctx, messages)
}
```

**典型应用：**
- 产品识别和描述
- 场景理解和分析
- 图像中的文字识别（OCR）
- 多图比较和对比

## 运行步骤

### 1. 准备图像资源

```bash
# 方式一：使用在线图像 URL
imageURL := "https://example.com/image.jpg"

# 方式二：使用本地图像（需转换为 Base64）
imageData, _ := os.ReadFile("local-image.jpg")
base64Image := fmt.Sprintf("data:image/jpeg;base64,%s", imageData)
```

### 2. 运行图像理解示例

```bash
cd lab11/image
go run image_understanding.go
```

### 3. 运行图文混合对话示例

```bash
cd lab11/multimodal
go run multimodal_chat.go
```

### 4. 运行视觉问答示例

```bash
cd lab11/vision
go run vision_qa.go
```

## 常见问题

### Q1: 支持哪些图像格式？

A: 常见的图像格式都支持，包括 JPEG、PNG、GIF、WebP 等。建议使用 JPEG 或 PNG 格式以获得最佳效果。

### Q2: 图像大小有限制吗？

A: 是的，不同的模型提供商有不同的限制。一般建议：
- 图像文件大小：< 20MB
- 图像分辨率：建议不超过 4096x4096
- 如果图像过大，建议先压缩或调整尺寸

### Q3: 如何使用本地图像？

A: 需要将本地图像转换为 Base64 编码：

```go
imageData, err := os.ReadFile("image.jpg")
if err != nil {
    log.Fatal(err)
}
// 添加 MIME 类型前缀
base64Image := fmt.Sprintf("data:image/jpeg;base64,%s", 
    base64.StdEncoding.EncodeToString(imageData))
```

### Q4: 可以在一条消息中包含多张图片吗？

A: 可以！只需添加多个 `ImageChatMessagePart`：

```go
Content: []schema.ChatMessagePart{
    schema.NewTextChatMessagePart("请比较这些图片："),
    schema.NewImageChatMessagePart(image1URL),
    schema.NewImageChatMessagePart(image2URL),
    schema.NewImageChatMessagePart(image3URL),
}
```

### Q5: 多模态对话的 Token 消耗如何计算？

A: 图像会被转换为 Token 进行处理，消耗量取决于：
- 图像的分辨率
- 图像的复杂度
- 模型的处理方式

一般来说，一张图像可能消耗几百到几千个 Token。

### Q6: 如何优化多模态应用的性能？

A: 几个优化建议：
1. 压缩图像：在保证质量的前提下减小文件大小
2. 调整分辨率：根据实际需求调整图像分辨率
3. 批量处理：对多张图片进行批量分析
4. 缓存结果：对相同图像的分析结果进行缓存

## 进阶学习

### 1. 流式多模态对话

使用 `Stream()` 方法实现实时响应：

```go
stream, err := chatModel.Stream(ctx, messages)
defer stream.Close()

for {
    chunk, err := stream.Recv()
    if err != nil {
        break
    }
    fmt.Print(chunk.Content)
}
```

### 2. 复杂的多模态工作流

设计包含多个步骤的工作流：
1. 图像上传和初步分析
2. 基于分析结果提出问题
3. 深入探讨特定细节
4. 生成总结报告

### 3. 多模态 RAG

结合检索增强生成（RAG）技术：
- 存储图像和相关文本到向量数据库
- 根据查询检索相关图像
- 基于检索结果生成回答

### 4. 实际应用场景

- **电商客服**：用户上传商品图片，AI 识别并提供信息
- **医疗辅助**：分析医学影像并提供初步诊断建议
- **教育辅助**：解答学生上传的题目图片
- **内容审核**：自动识别和过滤不当图像内容

## 最佳实践

1. **图像预处理**：在上传前对图像进行适当的压缩和优化
2. **错误处理**：妥善处理图像加载失败、格式不支持等错误
3. **提示词优化**：为不同的视觉任务设计专门的系统提示词
4. **成本控制**：监控 Token 使用量，避免不必要的重复分析
5. **隐私保护**：处理敏感图像时注意数据安全和隐私保护

## 参考资源

- [Eino 官方文档](https://rcn3ahrrdvjj.feishu.cn/wiki/space/7582137522705140933)
- [DeepSeek 多模态 API 文档](https://platform.deepseek.com/docs)
- [视觉问答研究论文](https://arxiv.org/abs/1505.00468)

---

**下一步学习**：完成本实验后，可以继续学习 Lab 12（Agent 智能体开发），了解如何构建更复杂的 AI 应用。
