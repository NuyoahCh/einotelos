package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// DatabaseQueryTool 数据库查询工具（模拟）
type DatabaseQueryTool struct {
	// 模拟的用户数据
	users []User
}

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func NewDatabaseQueryTool() *DatabaseQueryTool {
	return &DatabaseQueryTool{
		users: []User{
			{ID: 1, Name: "张三", Email: "zhangsan@example.com", Age: 28},
			{ID: 2, Name: "李四", Email: "lisi@example.com", Age: 32},
			{ID: 3, Name: "王五", Email: "wangwu@example.com", Age: 25},
		},
	}
}

func (t *DatabaseQueryTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "query_user",
		Desc: "查询用户信息",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"user_id": {
				Type: "number",
				Desc: "用户ID",
			},
			"name": {
				Type: "string",
				Desc: "用户姓名",
			},
		}),
	}, nil
}

type QueryUserParams struct {
	UserID int    `json:"user_id,omitempty"`
	Name   string `json:"name,omitempty"`
}

func (t *DatabaseQueryTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params QueryUserParams
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", err
	}

	var results []User

	// 根据条件查询
	for _, user := range t.users {
		if params.UserID > 0 && user.ID == params.UserID {
			results = append(results, user)
			break
		}
		if params.Name != "" && user.Name == params.Name {
			results = append(results, user)
		}
	}

	if len(results) == 0 {
		resultJSON, _ := json.Marshal(map[string]string{
			"message": "未找到匹配的用户",
		})
		return string(resultJSON), nil
	}

	resultJSON, err := json.Marshal(results)
	if err != nil {
		return "", err
	}

	return string(resultJSON), nil
}

func main() {
	ctx := context.Background()
	dbTool := NewDatabaseQueryTool()

	// 测试查询
	testCases := []QueryUserParams{
		{UserID: 1},
		{Name: "李四"},
		{UserID: 999}, // 不存在的用户
	}

	for _, tc := range testCases {
		paramsJSON, _ := json.Marshal(tc)
		result, err := dbTool.InvokableRun(ctx, string(paramsJSON))
		if err != nil {
			fmt.Printf("查询失败: %v\\n", err)
			continue
		}

		fmt.Printf("查询参数: %s\\n结果: %s\\n\\n", paramsJSON, result)
	}
}
