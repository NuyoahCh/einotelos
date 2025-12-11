package main

import (
    "context"
    "fmt"
    "strings"

    "github.com/cloudwego/eino/compose"
)

func main() {
    ctx := context.Background()

    // 1. 定义一个 Lambda 组件：string → string
    lambda := compose.InvokableLambda(
        func(ctx context.Context, input string) (string, error) {
            s := strings.TrimSpace(input)
            s = strings.ToLower(s)
            return s, nil
        },
    )

    // 2. 把 Lambda 挂到 Chain 里
    chain := compose.NewChain[string, string]()
    chain.AppendLambda(lambda)

    // 3. 编译成 runner，再用 runner.Invoke 调用
    runner, err := chain.Compile(ctx)
    if err != nil {
        panic(err)
    }

    out, err := runner.Invoke(ctx, "   HeLLo EinO   ")
    if err != nil {
        panic(err)
    }
    fmt.Println("cleaned:", out)
}
