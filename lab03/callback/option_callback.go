// demo_option_callback_deepseek.go
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"
	"unsafe"

	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

/*
本 demo 适配 eino v0.7.8 的 Option 结构：

type Option struct {
  apply func(opts *Options)
  implSpecificOptFn any
}

由于 apply/implSpecificOptFn 是未导出字段（小写），在 main 包不能直接访问，
*/

// -----------------------------
// 1) 自定义 Options（实现特定）
// -----------------------------
type MyChatModelOptions struct {
	Options    *model.Options
	RetryCount int
	Timeout    time.Duration
}

// -----------------------------
// 2) 自定义 Option：WithRetryCount / WithTimeout
// -----------------------------
func WithRetryCount(count int) model.Option {
	return model.WrapImplSpecificOptFn(func(o *MyChatModelOptions) {
		o.RetryCount = count
	})
}

func WithTimeout(timeout time.Duration) model.Option {
	return model.WrapImplSpecificOptFn(func(o *MyChatModelOptions) {
		o.Timeout = timeout
	})
}

// -----------------------------
// 3) Callback 机制（最简 demo）
// -----------------------------
type CallbackInput struct {
	Model    string
	Messages []*schema.Message
	Extra    map[string]any
}

type CallbackOutput struct {
	Message *schema.Message
	Usage   *schema.TokenUsage
	Extra   map[string]any
}

type ChatCallback interface {
	OnStart(ctx context.Context, in CallbackInput)
	OnEnd(ctx context.Context, out CallbackOutput, err error)
}

type LoggingCallback struct{}

func (LoggingCallback) OnStart(ctx context.Context, in CallbackInput) {
	fmt.Println("== [callback] start ==")
	fmt.Println("model:", in.Model)
	for i, m := range in.Messages {
		fmt.Printf("  [%d] role=%s content=%q\n", i, m.Role, m.Content)
	}
	if len(in.Extra) > 0 {
		fmt.Println("extra:", in.Extra)
	}
}

func (LoggingCallback) OnEnd(ctx context.Context, out CallbackOutput, err error) {
	fmt.Println("== [callback] end ==")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	if out.Message != nil {
		fmt.Println("assistant:", out.Message.Content)
	}
	if out.Usage != nil {
		fmt.Printf("usage: prompt=%d completion=%d total=%d\n",
			out.Usage.PromptTokens, out.Usage.CompletionTokens, out.Usage.TotalTokens)
	}
	if len(out.Extra) > 0 {
		fmt.Println("extra:", out.Extra)
	}
}

// -----------------------------
// 4) WrappedChatModel：在 Generate 中应用 options + callbacks + retry/timeout
// -----------------------------
type WrappedChatModel struct {
	inner     *deepseek.ChatModel
	modelName string
	callbacks []ChatCallback
}

func NewWrappedDeepSeek(apiKey string, callbacks ...ChatCallback) (*WrappedChatModel, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, errors.New("missing DEEPSEEK_API_KEY")
	}
	inner, err := deepseek.NewChatModel(context.Background(), &deepseek.ChatModelConfig{
		APIKey:  apiKey,
		Model:   "deepseek-chat",
		BaseURL: "https://api.deepseek.com",
	})
	if err != nil {
		return nil, err
	}
	return &WrappedChatModel{
		inner:     inner,
		modelName: "deepseek-chat",
		callbacks: callbacks,
	}, nil
}

func (m *WrappedChatModel) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	// 1) 解析 options（通用 + 实现特定）
	impl := &MyChatModelOptions{Options: &model.Options{}}
	if err := applyMyOptions(impl, opts...); err != nil {
		return nil, err
	}

	// 2) timeout（用 option 创建子 context）
	callCtx := ctx
	cancel := func() {}
	if impl.Timeout > 0 {
		callCtx, cancel = context.WithTimeout(ctx, impl.Timeout)
	}
	defer cancel()

	// 3) callback start
	in := CallbackInput{
		Model:    m.modelName,
		Messages: messages,
		Extra: map[string]any{
			"retry_count": impl.RetryCount,
			"timeout":     impl.Timeout.String(),
		},
	}
	for _, cb := range m.callbacks {
		cb.OnStart(callCtx, in)
	}

	// 4) retry
	var (
		resp    *schema.Message
		usage   *schema.TokenUsage
		lastErr error
	)
	start := time.Now()

	retries := impl.RetryCount
	if retries < 0 {
		retries = 0
	}

	for attempt := 0; attempt <= retries; attempt++ {
		r, err := m.inner.Generate(callCtx, messages)
		if err == nil {
			resp = r
			if r != nil && r.ResponseMeta != nil && r.ResponseMeta.Usage != nil {
				usage = r.ResponseMeta.Usage
			}
			lastErr = nil
			break
		}
		lastErr = err

		// ctx 超时/取消就别重试
		if errors.Is(callCtx.Err(), context.DeadlineExceeded) || errors.Is(callCtx.Err(), context.Canceled) {
			break
		}
		// 简单退避
		time.Sleep(time.Duration(attempt+1) * 250 * time.Millisecond)
	}

	// 5) callback end
	out := CallbackOutput{
		Message: resp,
		Usage:   usage,
		Extra: map[string]any{
			"latency_ms": time.Since(start).Milliseconds(),
		},
	}
	for _, cb := range m.callbacks {
		cb.OnEnd(callCtx, out, lastErr)
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return resp, nil
}

// -----------------------------
// 5) applyMyOptions：适配 v0.7.8 Option（含未导出字段）
// -----------------------------
func applyMyOptions(implOpts *MyChatModelOptions, opts ...model.Option) error {
	if implOpts == nil {
		return errors.New("nil impl options")
	}
	if implOpts.Options == nil {
		implOpts.Options = &model.Options{}
	}

	for _, opt := range opts {
		// 1) 通用 options：调用 opt.apply(*model.Options)
		if applyFn := getApplyFn(opt); applyFn != nil {
			applyFn(implOpts.Options)
		}
		// 2) 实现特定 options：opt.implSpecificOptFn 断言 func(*MyChatModelOptions)
		if implFn := getImplSpecificFn(opt); implFn != nil {
			implFn(implOpts)
		}
	}
	return nil
}

func getApplyFn(opt model.Option) func(*model.Options) {
	v := reflect.ValueOf(opt)
	f := v.FieldByName("apply")
	if !f.IsValid() || f.IsZero() {
		return nil
	}
	// 未导出字段 Interface() 可能 panic，先用 CanInterface 判断
	if !f.CanInterface() {
		// 通过 reflect.NewAt 绕过（仅用于 demo/学习）
		f = reflect.NewAt(f.Type(), unsafePointer(f)).Elem()
	}
	fn, ok := f.Interface().(func(*model.Options))
	if !ok {
		return nil
	}
	return fn
}

func getImplSpecificFn(opt model.Option) func(*MyChatModelOptions) {
	v := reflect.ValueOf(opt)
	f := v.FieldByName("implSpecificOptFn")
	if !f.IsValid() || f.IsZero() {
		return nil
	}
	if !f.CanInterface() {
		f = reflect.NewAt(f.Type(), unsafePointer(f)).Elem()
	}
	fn, ok := f.Interface().(func(*MyChatModelOptions))
	if !ok {
		return nil
	}
	return fn
}

// --- 小工具：把 reflect.Value 的地址转成 unsafe.Pointer（为读取未导出字段服务） ---
func unsafePointer(v reflect.Value) unsafe.Pointer {
	return unsafe.Pointer(v.UnsafeAddr())
}

// -----------------------------
// 6) main：运行 demo
// -----------------------------
func main() {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	w, err := NewWrappedDeepSeek(apiKey, LoggingCallback{})
	if err != nil {
		log.Fatal(err)
	}

	msgs := []*schema.Message{
		schema.SystemMessage("你是一个简洁、专业的篮球教练。"),
		schema.UserMessage("我每周打球2次，想提升运球和终结，请给我一周训练计划。"),
	}

	_, err = w.Generate(context.Background(), msgs,
		WithRetryCount(2),
		WithTimeout(12*time.Second),
	)
	if err != nil {
		log.Fatal("generate failed:", err)
	}
}
