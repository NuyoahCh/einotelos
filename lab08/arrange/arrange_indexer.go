package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cloudwego/eino/callbacks"
	einoIndexer "github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	callbacksHelper "github.com/cloudwego/eino/utils/callbacks"

	"github.com/cloudwego/eino-ext/components/indexer/volc_vikingdb"
)

func main() {
	ctx := context.Background()

	ak := os.Getenv("VIKING_AK")
	sk := os.Getenv("VIKING_SK")

	collectionName := "eino_test" // 确保你已按官方说明创建好字段和向量维度 :contentReference[oaicite:10]{index=10}

	cfg := &volc_vikingdb.IndexerConfig{
		Host:   "api-vikingdb.volces.com",
		Region: "cn-beijing",
		AK:     ak,
		SK:     sk,
		Scheme: "https",
		Collection: collectionName,
		EmbeddingConfig: volc_vikingdb.EmbeddingConfig{
			UseBuiltin: true,
			ModelName:  "bge-m3",
			UseSparse:  true,
		},
		AddBatchSize: 10,
	}

	volcIndexer, err := volc_vikingdb.NewIndexer(ctx, cfg) // 单独使用示例 :contentReference[oaicite:11]{index=11}
	if err != nil {
		panic(err)
	}

	// 1) 准备 docs
	doc := &schema.Document{
		ID:      "mock_id_1",
		Content: "Indexer Lab: store docs with callbacks via compose chain",
	}
	volc_vikingdb.SetExtraDataFields(doc, map[string]interface{}{"extra_field_1": "mock_ext_abc"})
	volc_vikingdb.SetExtraDataTTL(doc, 1000)
	docs := []*schema.Document{doc}

	// 2) Callback：在 Store 前后观察 input/output（官方示例同款）:contentReference[oaicite:12]{index=12}
	handler := &callbacksHelper.IndexerCallbackHandler{
		OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *einoIndexer.CallbackInput) context.Context {
			log.Printf("[Indexer OnStart] docs=%d, first=%q\n", len(input.Docs), input.Docs[0].Content)
			return ctx
		},
		OnEnd: func(ctx context.Context, info *callbacks.RunInfo, output *einoIndexer.CallbackOutput) context.Context {
			log.Printf("[Indexer OnEnd] ids=%v\n", output.IDs)
			return ctx
		},
	}
	helper := callbacksHelper.NewHandlerHelper().Indexer(handler).Handler()

	// 3) 编排：把 Indexer 放进 chain，然后 Invoke
	chain := compose.NewChain[[]*schema.Document, []string]()
	chain.AppendIndexer(volcIndexer) // 在编排中使用 :contentReference[oaicite:13]{index=13}

	run, err := chain.Compile(ctx)
	if err != nil {
		panic(err)
	}

	outIDs, err := run.Invoke(ctx, docs, compose.WithCallbacks(helper))
	if err != nil {
		panic(err)
	}

	fmt.Printf("store success, outIDs=%v\n", outIDs)
}
