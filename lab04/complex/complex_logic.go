package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

type UserProfile struct {
	Name       string
	Age        int
	Interests  []string
	VIPLevel   int
}

func main() {
	ctx := context.Background()

	template := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage("你是一个智能推荐助手"),
		schema.UserMessage(`用户信息：
姓名：{name}
年龄：{age}
兴趣：{interests}
VIP等级：{vip_level}

请根据以上信息推荐合适的内容。`),
	)

	// 准备用户数据
	user := UserProfile{
		Name:      "张三",
		Age:       28,
		Interests: []string{"编程", "阅读", "旅行"},
		VIPLevel:  3,
	}

	variables := map[string]any{
		"name":      user.Name,
		"age":       user.Age,
		"interests": fmt.Sprintf("%v", user.Interests),
		"vip_level": user.VIPLevel,
	}

	messages, err := template.Format(ctx, variables)
	if err != nil {
		log.Fatalf("格式化失败: %v", err)
	}

	for _, msg := range messages {
		fmt.Printf("[%s]\n%s\n\n", msg.Role, msg.Content)
	}
}
