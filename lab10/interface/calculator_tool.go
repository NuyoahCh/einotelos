package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// CalculatorTool 计算器工具
type CalculatorTool struct{}

// Info 返回工具信息
func (t *CalculatorTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "calculator",
		Desc: "执行基本的数学计算（加、减、乘、除）",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"operation": {
				Type:     "string",
				Desc:     "运算类型: add(加), subtract(减), multiply(乘), divide(除)",
				Required: true,
			},
			"a": {
				Type:     "number",
				Desc:     "第一个数字",
				Required: true,
			},
			"b": {
				Type:     "number",
				Desc:     "第二个数字",
				Required: true,
			},
		}),
	}, nil
}

// 参数结构
type CalculatorParams struct {
	Operation string  `json:"operation"`
	A         float64 `json:"a"`
	B         float64 `json:"b"`
}

// 结果结构
type CalculatorResult struct {
	Result float64 `json:"result"`
	Error  string  `json:"error,omitempty"`
}

// InvokableRun 执行计算
func (t *CalculatorTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	// 1. 解析参数
	var params CalculatorParams
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	// 2. 执行计算
	var result float64
	switch params.Operation {
	case "add":
		result = params.A + params.B
	case "subtract":
		result = params.A - params.B
	case "multiply":
		result = params.A * params.B
	case "divide":
		if params.B == 0 {
			resultJSON, _ := json.Marshal(CalculatorResult{
				Error: "除数不能为0",
			})
			return string(resultJSON), nil
		}
		result = params.A / params.B
	default:
		resultJSON, _ := json.Marshal(CalculatorResult{
			Error: "不支持的运算类型",
		})
		return string(resultJSON), nil
	}

	// 3. 返回结果
	resultJSON, err := json.Marshal(CalculatorResult{
		Result: result,
	})
	if err != nil {
		return "", err
	}

	return string(resultJSON), nil
}

func main() {
	ctx := context.Background()
	calculator := &CalculatorTool{}

	// 测试工具
	testCases := []struct {
		operation string
		a, b      float64
	}{
		{"add", 10, 5},
		{"subtract", 10, 5},
		{"multiply", 10, 5},
		{"divide", 10, 5},
		{"divide", 10, 0}, // 测试除零
	}

	for _, tc := range testCases {
		params := CalculatorParams{
			Operation: tc.operation,
			A:         tc.a,
			B:         tc.b,
		}

		paramsJSON, _ := json.Marshal(params)
		result, err := calculator.InvokableRun(ctx, string(paramsJSON))
		if err != nil {
			log.Printf("执行失败: %v", err)
			continue
		}

		fmt.Printf("%s(%g, %g) = %s\\n", tc.operation, tc.a, tc.b, result)
	}
}
