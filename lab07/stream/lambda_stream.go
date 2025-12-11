package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func main() {
	ctx := context.Background()

	// 1. 定义 Lambda：string → *schema.StreamReader[string]
	lambda := compose.StreamableLambda(
		func(ctx context.Context, input string) (*schema.StreamReader[string], error) {
			words := strings.Fields(input)

			// 用 Pipe 创建一个可写的 StreamWriter 和只读的 StreamReader
			sr, sw := schema.Pipe[string](0)

			go func() {
				defer sw.Close()
				for _, w := range words {
					select {
					case <-ctx.Done():
						return
					default:
						// Send 返回 true 说明下游已经关闭，不需要再写了
						closed := sw.Send(w, nil)
						if closed {
							return
						}
						time.Sleep(200 * time.Millisecond) // 模拟流式延迟
					}
				}
			}()

			return sr, nil
		},
	)

	// 2. 放到 Chain 里
	chain := compose.NewChain[string, string]() // Stream 模式下 output 类型是流里的元素类型
	chain.AppendLambda(lambda)

	// 3. 编译 runner，用 runner.Stream 获取流
	runner, err := chain.Compile(ctx)
	if err != nil {
		panic(err)
	}

	stream, err := runner.Stream(ctx, "Eino Lambda makes custom logic easy")
	if err != nil {
		panic(err)
	}
	defer stream.Close()

	for {
		w, err := stream.Recv()
		if err != nil {
			break
		}
		fmt.Println("chunk:", w)
	}
}
