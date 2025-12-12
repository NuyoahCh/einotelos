package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// FileReaderTool 文件读取工具
type FileReaderTool struct{}

func (t *FileReaderTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "read_file",
		Desc: "读取文件内容",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"filepath": {
				Type:     "string",
				Desc:     "文件路径",
				Required: true,
			},
		}),
	}, nil
}

type FileReaderParams struct {
	FilePath string `json:"filepath"`
}

type FileReaderResult struct {
	Content string `json:"content,omitempty"`
	Error   string `json:"error,omitempty"`
}

func (t *FileReaderTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params FileReaderParams
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", err
	}

	// 读取文件
	content, err := os.ReadFile(params.FilePath)
	if err != nil {
		result := FileReaderResult{
			Error: fmt.Sprintf("读取文件失败: %v", err),
		}
		resultJSON, _ := json.Marshal(result)
		return string(resultJSON), nil
	}

	result := FileReaderResult{
		Content: string(content),
	}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", err
	}

	return string(resultJSON), nil
}

// FileWriterTool 文件写入工具
type FileWriterTool struct{}

func (t *FileWriterTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "write_file",
		Desc: "写入内容到文件",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"filepath": {
				Type:     "string",
				Desc:     "文件路径",
				Required: true,
			},
			"content": {
				Type:     "string",
				Desc:     "要写入的内容",
				Required: true,
			},
		}),
	}, nil
}

type FileWriterParams struct {
	FilePath string `json:"filepath"`
	Content  string `json:"content"`
}

type FileWriterResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (t *FileWriterTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params FileWriterParams
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", err
	}

	// 写入文件
	err := os.WriteFile(params.FilePath, []byte(params.Content), 0644)

	var result FileWriterResult
	if err != nil {
		result = FileWriterResult{
			Success: false,
			Message: fmt.Sprintf("写入失败: %v", err),
		}
	} else {
		result = FileWriterResult{
			Success: true,
			Message: "写入成功",
		}
	}

	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

func main() {
	ctx := context.Background()

	// 测试文件写入
	writer := &FileWriterTool{}
	writeParams := FileWriterParams{
		FilePath: "test_output.txt",
		Content:  "Hello from Eino Tool!\\n这是测试内容。",
	}
	writeParamsJSON, _ := json.Marshal(writeParams)
	writeResult, err := writer.InvokableRun(ctx, string(writeParamsJSON))
	if err != nil {
		log.Fatalf("写入失败: %v", err)
	}
	fmt.Printf("写入结果: %s\\n", writeResult)

	// 测试文件读取
	reader := &FileReaderTool{}
	readParams := FileReaderParams{
		FilePath: "test_output.txt",
	}
	readParamsJSON, _ := json.Marshal(readParams)
	readResult, err := reader.InvokableRun(ctx, string(readParamsJSON))
	if err != nil {
		log.Fatalf("读取失败: %v", err)
	}
	fmt.Printf("读取结果: %s\\n", readResult)

	// 清理测试文件
	os.Remove("test_output.txt")
}
