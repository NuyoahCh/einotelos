package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cloudwego/eino/callbacks"
	einoRetriever "github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	callbacksHelper "github.com/cloudwego/eino/utils/callbacks"

	"github.com/cloudwego/eino-ext/components/embedding/ollama"
	rr "github.com/cloudwego/eino-ext/components/retriever/redis"
	"github.com/redis/go-redis/v9"
)

// ====== 你在工程里最终想得到的“可喂给大模型的材料” ======
type RagPack struct {
	OriginalQuery string
	RewriteQuery  string
	Docs          []*schema.Document
	Context       string // 拼好的证据上下文
	HasKnowledge  bool
}

func main() {
	ctx := context.Background()

	// 1) Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:          "localhost:6379",
		Protocol:      2,
		UnstableResp3: true,
	})

	// 2) Embedder（用于 query 向量化；Retriever 公共 option 里就有 Embedding）:contentReference[oaicite:2]{index=2}
	embedder, err := ollama.NewEmbedder(ctx, &ollama.EmbeddingConfig{
		Model:   "modelscope.cn/nomic-ai/nomic-embed-text-v1.5-GGUF:latest",
		BaseURL: "http://127.0.0.1:11434",
	})
	if err != nil {
		panic(err)
	}

	// 3) Redis Retriever
	ret, err := rr.NewRetriever(ctx, &rr.RetrieverConfig{
		Client:    rdb,
		Index:     "doc_index",
		Embedding: embedder,
	})
	if err != nil {
		panic(err)
	}

	// 4) Callback：把检索变“可观测”。官方示例同款 RetrieverCallbackHandler :contentReference[oaicite:3]{index=3}
	handler := &callbacksHelper.RetrieverCallbackHandler{
		OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *einoRetriever.CallbackInput) context.Context {
			log.Printf("[Retriever OnStart] query=%q", input.Query)
			return ctx
		},
		OnEnd: func(ctx context.Context, info *callbacks.RunInfo, output *einoRetriever.CallbackOutput) context.Context {
			log.Printf("[Retriever OnEnd] docs=%d", len(output.Docs))
			return ctx
		},
	}
	cb := callbacksHelper.NewHandlerHelper().Retriever(handler).Handler()

	// 5) 运行一次“检索增强”
	userQuery := "狗的常见品种有哪些？我想养一只适合新手的"
	pack, err := RetrieveAugment(
		ctx,
		ret,
		cb,
		userQuery,
		RetrievePolicy{
			TopK:           5,
			ScoreThreshold: 0.2,
			SubIndex:       "pet_kb", // 你有做子索引隔离的话就用；没有就留空
		},
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("=== RagPack ===")
	fmt.Println("original:", pack.OriginalQuery)
	fmt.Println("rewrite  :", pack.RewriteQuery)
	fmt.Println("has_kb   :", pack.HasKnowledge)
	fmt.Println("context  :\n", pack.Context)
}

// ====== 检索策略（你可以按业务扩展成 A/B、按 query 长度自适应等） ======
type RetrievePolicy struct {
	TopK           int
	ScoreThreshold float64
	Index          string
	SubIndex       string
	Timeout        time.Duration
}

// RetrieveAugment：
// 1) rewrite query（让检索更稳）
// 2) 调用 Retriever.Retrieve（用公共 Options 控制 TopK/阈值/子索引/embedding 等）:contentReference[oaicite:4]{index=4}
// 3) 空召回兜底（防止硬塞“最像但不相关”的片段）
// 4) 组装 context（可直接进 prompt）
func RetrieveAugment(
	ctx context.Context,
	ret einoRetriever.Retriever,
	cb callbacks.Handler,
	query string,
	p RetrievePolicy,
) (*RagPack, error) {
	if p.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.Timeout)
		defer cancel()
	}

	rewrite := RewriteQuery(query)

	// 公共 Option：TopK / ScoreThreshold / Index / SubIndex / Embedding 等 :contentReference[oaicite:6]{index=6}
	opts := make([]einoRetriever.Option, 0, 4)
	if p.TopK > 0 {
		opts = append(opts, einoRetriever.WithTopK(p.TopK))
	}
	if p.ScoreThreshold > 0 {
		opts = append(opts, einoRetriever.WithScoreThreshold(p.ScoreThreshold))
	}
	if p.Index != "" {
		opts = append(opts, einoRetriever.WithIndex(p.Index))
	}
	if p.SubIndex != "" {
		opts = append(opts, einoRetriever.WithSubIndex(p.SubIndex))
	}

	docs, err := ret.Retrieve(ctx, rewrite, opts...)
	if err != nil {
		return nil, err
	}

	// 空召回兜底：没有命中就不要硬拼“伪证据”
	hasKB := len(docs) > 0
	context := BuildContext(docs)
	if !hasKB {
		context = "（知识库未命中可靠内容：请基于常识回答，并提醒用户补充更多信息或更精确的问题）"
	}

	return &RagPack{
		OriginalQuery: query,
		RewriteQuery:  rewrite,
		Docs:          docs,
		Context:       context,
		HasKnowledge:  hasKB,
	}, nil
}

// RewriteQuery：不依赖 LLM 的“轻量 rewrite”，核心目标是让检索更像“关键词+意图”
// 真实项目里你可以替换成：LLM rewrite / 多路 query 扩展 / 同义词扩展等
func RewriteQuery(q string) string {
	s := strings.TrimSpace(q)
	s = strings.ReplaceAll(s, "？", "")
	s = strings.ReplaceAll(s, "。", "")
	// 非严格示例：把“我想/我要/适合新手”这类语气词压缩成更检索友好的表达
	s = strings.ReplaceAll(s, "我想", "")
	s = strings.ReplaceAll(s, "我要", "")
	s = strings.ReplaceAll(s, "适合新手的", "新手 适合")
	s = strings.ReplaceAll(s, "有哪些", "列表")
	// 让 query 更“检索化”
	return strings.Join(strings.Fields(s), " ")
}

// BuildContext：把 docs 转成可喂给大模型的证据块
func BuildContext(docs []*schema.Document) string {
	if len(docs) == 0 {
		return ""
	}
	var b strings.Builder
	for i, d := range docs {
		// 你也可以把 MetaData 的来源、分数、chunk_id 等拼进去做可追溯
		fmt.Fprintf(&b, "证据%d（id=%s）:\n%s\n\n", i+1, d.ID, d.Content)
	}
	return b.String()
}
