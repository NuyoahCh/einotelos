package main

import (
    "context"
    "fmt"
    "strings"

    "github.com/cloudwego/eino/compose"
)

// 1. 定义 Lambda 自己的 Option 类型
type Options struct {
    Upper bool
}

type MyOption func(*Options)

func WithUpper() MyOption {
    return func(o *Options) {
        o.Upper = true
    }
}

func main() {
    ctx := context.Background()

    // 2. 用 InvokableLambdaWithOption 创建 Lambda，第三个参数是 MyOption
    lambda := compose.InvokableLambdaWithOption(
        func(ctx context.Context, input string, opts ...MyOption) (string, error) {
            // 把可变参数的 MyOption 合并成一个 Options
            cfg := &Options{}
            for _, opt := range opts {
                opt(cfg)
            }

            if cfg.Upper {
                return strings.ToUpper(input), nil
            }
            return input, nil
        },
    )

    // 3. 丢到 Chain 里跑
    chain := compose.NewChain[string, string]()
    chain.AppendLambda(lambda)

    runner, err := chain.Compile(ctx)
    if err != nil {
        panic(err)
    }

    // 3.1 不带任何 Lambda Option，走默认行为
    out1, err := runner.Invoke(ctx, "hello eino")
    if err != nil {
        panic(err)
    }
    fmt.Println("default:", out1) // default: hello eino

    // 3.2 通过 WithLambdaOption 把 MyOption 适配为 compose.Option
    out2, err := runner.Invoke(
        ctx,
        "hello eino",
        compose.WithLambdaOption(WithUpper()), // ✅ 关键改动在这里
    )
    if err != nil {
        panic(err)
    }
    fmt.Println("with upper:", out2) // with upper: HELLO EINO
}
