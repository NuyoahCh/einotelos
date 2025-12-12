package main

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
)

// ======= (1) 定义 Document 结构体 =======
type Document struct {
	ID      string
	Content string
	Meta    map[string]any
	Vector  []float32
}

func (d *Document) WithVector(v []float32) { d.Vector = v }

// ======= (2) 定义 Embedder 接口及其简单实现 =======
type Embedder interface {
	EmbedStrings(ctx context.Context, texts []string) ([][]float32, error)
}

type FakeEmbedder struct{}

func (e *FakeEmbedder) EmbedStrings(ctx context.Context, texts []string) ([][]float32, error) {
	out := make([][]float32, 0, len(texts))
	for _, t := range texts {
		// 生成一个非常简单的“假向量”：长度=4，用内容长度等特征填充
		out = append(out, []float32{
			float32(len(t)),
			float32(len([]rune(t))),
			float32(t[0]),
			1.0,
		})
	}
	return out, nil
}

// ======= (3) 定义 Option 结构体及其构造函数 =======
type Options struct {
	SubIndexes []string
	Embedding  Embedder
}

type Option func(*Options)

func WithSubIndexes(sub []string) Option {
	return func(o *Options) { o.SubIndexes = sub }
}
func WithEmbedding(emb Embedder) Option {
	return func(o *Options) { o.Embedding = emb }
}

// ======= (4) 定义 Indexer 接口 =======
type Indexer interface {
	Store(ctx context.Context, docs []*Document, opts ...Option) ([]string, error)
}

// ======= (5) 实现内存索引器 MemoryIndexer =======
type MemoryIndexer struct {
	store map[string]*Document
}

func NewMemoryIndexer() *MemoryIndexer {
	return &MemoryIndexer{store: map[string]*Document{}}
}

func (mi *MemoryIndexer) Store(ctx context.Context, docs []*Document, opts ...Option) ([]string, error) {
	// 1) 合并公共 option
	o := &Options{}
	for _, opt := range opts {
		opt(o)
	}
	if len(o.SubIndexes) == 0 {
		o.SubIndexes = []string{"default"}
	}

	// 2) 如果注入了 Embedding，则先向量化（这对应官方“Embedding 是用于生成文档向量的组件”）:contentReference[oaicite:4]{index=4}
	if o.Embedding != nil {
		texts := make([]string, len(docs))
		for i, d := range docs {
			texts[i] = d.Content
		}
		vecs, err := o.Embedding.EmbedStrings(ctx, texts)
		if err != nil {
			return nil, err
		}
		for i, d := range docs {
			d.WithVector(vecs[i])
		}
	}

	// 3) 写入存储：对每个 SubIndex 都建一份逻辑索引
	ids := make([]string, 0, len(docs))
	for _, d := range docs {
		if d.ID == "" {
			d.ID = hashID(d.Content)
		}
		ids = append(ids, d.ID)

		for _, sub := range o.SubIndexes {
			key := fmt.Sprintf("%s:%s", sub, d.ID)
			mi.store[key] = d
		}
	}
	return ids, nil
}

func hashID(s string) string {
	h := sha1.Sum([]byte(s))
	return hex.EncodeToString(h[:8])
}

// ======= (6) 运行实验 =======
func main() {
	ctx := context.Background()
	indexer := NewMemoryIndexer()

	docs := []*Document{
		{Content: "Eino Indexer lab doc A", Meta: map[string]any{"source": "lab"}},
		{Content: "Eino Indexer lab doc B", Meta: map[string]any{"source": "lab"}},
	}

	ids, err := indexer.Store(
		ctx,
		docs,
		WithSubIndexes([]string{"kb_1", "kb_2"}), // 对应官方 Option 示例 :contentReference[oaicite:5]{index=5}
		WithEmbedding(&FakeEmbedder{}),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("store ids:", ids)
	fmt.Println("doc[0] vector:", docs[0].Vector)
	fmt.Println("in-memory keys sample: kb_1:<id>, kb_2:<id>")
}
