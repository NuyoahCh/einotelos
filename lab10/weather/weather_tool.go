package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// WeatherTool 天气查询工具
type WeatherTool struct {
	// 模拟的天气数据
	weatherData map[string]map[string]string
}

func NewWeatherTool() *WeatherTool {
	return &WeatherTool{
		weatherData: map[string]map[string]string{
			"北京": {
				"temperature": "25°C",
				"condition":   "晴天",
				"humidity":    "45%",
				"wind":        "北风3级",
			},
			"上海": {
				"temperature": "28°C",
				"condition":   "多云",
				"humidity":    "65%",
				"wind":        "东南风2级",
			},
			"深圳": {
				"temperature": "30°C",
				"condition":   "阴天",
				"humidity":    "75%",
				"wind":        "南风4级",
			},
		},
	}
}

func (t *WeatherTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "get_weather",
		Desc: "查询指定城市的天气信息",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"city": {
				Type:     "string",
				Desc:     "城市名称，例如：北京、上海、深圳",
				Required: true,
			},
		}),
	}, nil
}

type WeatherParams struct {
	City string `json:"city"`
}

func (t *WeatherTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params WeatherParams
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", err
	}

	weather, exists := t.weatherData[params.City]
	if !exists {
		result := map[string]string{
			"error": fmt.Sprintf("暂无 %s 的天气数据", params.City),
		}
		resultJSON, _ := json.Marshal(result)
		return string(resultJSON), nil
	}

	// 返回天气信息
	resultJSON, err := json.Marshal(weather)
	if err != nil {
		return "", err
	}

	return string(resultJSON), nil
}

func main() {
	ctx := context.Background()
	weatherTool := NewWeatherTool()

	cities := []string{"北京", "上海", "广州"}

	for _, city := range cities {
		params := WeatherParams{City: city}
		paramsJSON, _ := json.Marshal(params)

		result, err := weatherTool.InvokableRun(ctx, string(paramsJSON))
		if err != nil {
			log.Printf("查询失败: %v", err)
			continue
		}

		fmt.Printf("%s 天气: %s\\n", city, result)
	}
}
