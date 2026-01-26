package common

import "github.com/cloudwego/eino/schema"

// NewTextPart 创建文本消息部分
func NewTextPart(text string) schema.MessageInputPart {
	return schema.MessageInputPart{
		Type: schema.ChatMessagePartTypeText,
		Text: text,
	}
}

// NewImageURLPart 创建图像 URL 消息部分
func NewImageURLPart(imageURL string) schema.MessageInputPart {
	return schema.MessageInputPart{
		Type: schema.ChatMessagePartTypeImageURL,
		Image: &schema.MessageInputImage{
			MessagePartCommon: schema.MessagePartCommon{
				URL: &imageURL,
			},
		},
	}
}

// NewImageBase64Part 创建 Base64 编码的图像消息部分
func NewImageBase64Part(base64Data, mimeType string) schema.MessageInputPart {
	return schema.MessageInputPart{
		Type: schema.ChatMessagePartTypeImageURL,
		Image: &schema.MessageInputImage{
			MessagePartCommon: schema.MessagePartCommon{
				Base64Data: &base64Data,
				MIMEType:   mimeType,
			},
		},
	}
}
