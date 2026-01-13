// Package tools 提供工具组件
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

// CalculatorTool 计算器工具
type CalculatorTool struct{}

type CalculatorParams struct {
	Expression string `json:"expression"`
}

type CalculatorResult struct {
	Result string `json:"result"`
	Error  string `json:"error,omitempty"`
}

func (t *CalculatorTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "calculator",
		Desc: "执行数学计算，支持加减乘除和基本数学函数",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"expression": {
				Type:     "string",
				Desc:     "数学表达式，例如: 10 + 20, 100 * 2, sqrt(16)",
				Required: true,
			},
		}),
	}, nil
}

func (t *CalculatorTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params CalculatorParams
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	result, err := evaluateExpression(params.Expression)
	if err != nil {
		res := CalculatorResult{Error: err.Error()}
		b, _ := json.Marshal(res)
		return string(b), nil
	}

	res := CalculatorResult{Result: fmt.Sprintf("%v", result)}
	b, _ := json.Marshal(res)
	return string(b), nil
}

// evaluateExpression 简单的表达式求值
func evaluateExpression(expr string) (float64, error) {
	expr = strings.TrimSpace(expr)
	expr = strings.ToLower(expr)

	// 处理 sqrt
	if strings.HasPrefix(expr, "sqrt(") && strings.HasSuffix(expr, ")") {
		inner := expr[5 : len(expr)-1]
		val, err := strconv.ParseFloat(inner, 64)
		if err != nil {
			return 0, fmt.Errorf("无效的数字: %s", inner)
		}
		return math.Sqrt(val), nil
	}

	// 处理 pow
	if strings.HasPrefix(expr, "pow(") && strings.HasSuffix(expr, ")") {
		inner := expr[4 : len(expr)-1]
		parts := strings.Split(inner, ",")
		if len(parts) != 2 {
			return 0, fmt.Errorf("pow 需要两个参数")
		}
		base, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		if err != nil {
			return 0, err
		}
		exp, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			return 0, err
		}
		return math.Pow(base, exp), nil
	}

	// 处理基本运算
	for _, op := range []string{"+", "-", "*", "/"} {
		if idx := strings.LastIndex(expr, op); idx > 0 {
			left, err := strconv.ParseFloat(strings.TrimSpace(expr[:idx]), 64)
			if err != nil {
				continue
			}
			right, err := strconv.ParseFloat(strings.TrimSpace(expr[idx+1:]), 64)
			if err != nil {
				continue
			}

			switch op {
			case "+":
				return left + right, nil
			case "-":
				return left - right, nil
			case "*":
				return left * right, nil
			case "/":
				if right == 0 {
					return 0, fmt.Errorf("除数不能为零")
				}
				return left / right, nil
			}
		}
	}

	// 尝试直接解析为数字
	return strconv.ParseFloat(expr, 64)
}

// WeatherTool 天气查询工具
type WeatherTool struct {
	data map[string]map[string]string
}

func NewWeatherTool() *WeatherTool {
	return &WeatherTool{
		data: map[string]map[string]string{
			"北京": {"temperature": "25°C", "condition": "晴天", "humidity": "45%", "wind": "北风3级"},
			"上海": {"temperature": "28°C", "condition": "多云", "humidity": "65%", "wind": "东南风2级"},
			"深圳": {"temperature": "30°C", "condition": "阴天", "humidity": "75%", "wind": "南风4级"},
			"杭州": {"temperature": "26°C", "condition": "晴天", "humidity": "55%", "wind": "东风2级"},
			"成都": {"temperature": "22°C", "condition": "小雨", "humidity": "80%", "wind": "微风"},
		},
	}
}

type WeatherParams struct {
	City string `json:"city"`
}

func (t *WeatherTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "get_weather",
		Desc: "查询指定城市的天气信息，包括温度、天气状况、湿度和风力",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"city": {
				Type:     "string",
				Desc:     "城市名称，例如：北京、上海、深圳",
				Required: true,
			},
		}),
	}, nil
}

func (t *WeatherTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params WeatherParams
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", err
	}

	weather, exists := t.data[params.City]
	if !exists {
		result := map[string]string{"error": fmt.Sprintf("暂无 %s 的天气数据", params.City)}
		b, _ := json.Marshal(result)
		return string(b), nil
	}

	b, _ := json.Marshal(weather)
	return string(b), nil
}

// TimeTool 时间工具
type TimeParams struct {
	Format   string `json:"format"`
	Timezone string `json:"timezone"`
}

type TimeResult struct {
	CurrentTime string `json:"current_time"`
	Timezone    string `json:"timezone"`
}

func NewTimeTool() tool.BaseTool {
	return utils.NewTool(
		&schema.ToolInfo{
			Name: "get_current_time",
			Desc: "获取当前时间，支持不同格式和时区",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"format": {
					Type: "string",
					Desc: "时间格式: date(日期), time(时间), datetime(日期时间), unix(时间戳)",
				},
				"timezone": {
					Type: "string",
					Desc: "时区，例如: Asia/Shanghai, America/New_York",
				},
			}),
		},
		func(ctx context.Context, params *TimeParams) (*TimeResult, error) {
			loc := time.Local
			if params.Timezone != "" {
				if l, err := time.LoadLocation(params.Timezone); err == nil {
					loc = l
				}
			}

			now := time.Now().In(loc)
			var result string

			switch params.Format {
			case "date":
				result = now.Format("2006-01-02")
			case "time":
				result = now.Format("15:04:05")
			case "unix":
				result = fmt.Sprintf("%d", now.Unix())
			default:
				result = now.Format("2006-01-02 15:04:05")
			}

			return &TimeResult{
				CurrentTime: result,
				Timezone:    loc.String(),
			}, nil
		},
	)
}

// SearchTool 搜索工具（模拟）
type SearchParams struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

type SearchResult struct {
	Results []SearchItem `json:"results"`
	Total   int          `json:"total"`
}

type SearchItem struct {
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
	URL     string `json:"url"`
}

func NewSearchTool() tool.BaseTool {
	return utils.NewTool(
		&schema.ToolInfo{
			Name: "web_search",
			Desc: "搜索网络信息（模拟）",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"query": {
					Type:     "string",
					Desc:     "搜索关键词",
					Required: true,
				},
				"limit": {
					Type: "integer",
					Desc: "返回结果数量，默认3",
				},
			}),
		},
		func(ctx context.Context, params *SearchParams) (*SearchResult, error) {
			// 模拟搜索结果
			limit := params.Limit
			if limit <= 0 {
				limit = 3
			}

			results := []SearchItem{
				{
					Title:   fmt.Sprintf("关于 %s 的介绍", params.Query),
					Snippet: fmt.Sprintf("这是关于 %s 的详细介绍...", params.Query),
					URL:     "https://example.com/1",
				},
				{
					Title:   fmt.Sprintf("%s 最新资讯", params.Query),
					Snippet: fmt.Sprintf("最新的 %s 相关新闻和动态...", params.Query),
					URL:     "https://example.com/2",
				},
				{
					Title:   fmt.Sprintf("%s 使用指南", params.Query),
					Snippet: fmt.Sprintf("如何使用 %s 的完整指南...", params.Query),
					URL:     "https://example.com/3",
				},
			}

			if limit < len(results) {
				results = results[:limit]
			}

			return &SearchResult{
				Results: results,
				Total:   len(results),
			}, nil
		},
	)
}

// GetAllTools 获取所有工具
func GetAllTools() []tool.BaseTool {
	return []tool.BaseTool{
		&CalculatorTool{},
		NewWeatherTool(),
		NewTimeTool(),
		NewSearchTool(),
	}
}

// GetToolInfos 获取所有工具信息
func GetToolInfos(ctx context.Context, tools []tool.BaseTool) ([]*schema.ToolInfo, error) {
	infos := make([]*schema.ToolInfo, 0, len(tools))
	for _, t := range tools {
		info, err := t.Info(ctx)
		if err != nil {
			return nil, err
		}
		infos = append(infos, info)
	}
	return infos, nil
}
