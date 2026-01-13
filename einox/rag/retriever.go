// Package rag 提供检索器组件
package rag

import (
	"context"
	"fmt"
	"strings"
)

// Retriever 检索器
type Retriever struct {
	store     VectorStore
	embedder  Embedder
	topK      int
	threshold float64
}

// RetrieverConfig 检索器配置
type RetrieverConfig struct {
	Store     VectorStore
	Embedder  Embedder
	TopK      int
	Threshold float64
}

// NewRetriever 创建检索器
func NewRetriever(cfg *RetrieverConfig) *Retriever {
	topK := cfg.TopK
	if topK <= 0 {
		topK = 5
	}

	threshold := cfg.Threshold
	if threshold <= 0 {
		threshold = 0.5
	}

	return &Retriever{
		store:     cfg.Store,
		embedder:  cfg.Embedder,
		topK:      topK,
		threshold: threshold,
	}
}

// Retrieve 检索相关文档
func (r *Retriever) Retrieve(ctx context.Context, query string) ([]*Document, error) {
	// 1. 向量化查询
	vectors, err := r.embedder.EmbedStrings(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("向量化查询失败: %w", err)
	}

	if len(vectors) == 0 || len(vectors[0]) == 0 {
		return nil, fmt.Errorf("向量化结果为空")
	}

	// 2. 搜索相似文档
	docs, err := r.store.Search(ctx, vectors[0], r.topK, r.threshold)
	if err != nil {
		return nil, fmt.Errorf("搜索失败: %w", err)
	}

	return docs, nil
}

// RAGPipeline RAG 流水线
type RAGPipeline struct {
	retriever *Retriever
	splitter  DocumentSplitter
}

// NewRAGPipeline 创建 RAG 流水线
func NewRAGPipeline(retriever *Retriever, splitter DocumentSplitter) *RAGPipeline {
	return &RAGPipeline{
		retriever: retriever,
		splitter:  splitter,
	}
}

// IndexDocuments 索引文档
func (p *RAGPipeline) IndexDocuments(ctx context.Context, docs []*Document) error {
	// 1. 分割文档
	if p.splitter != nil {
		var err error
		docs, err = p.splitter.Split(ctx, docs)
		if err != nil {
			return fmt.Errorf("分割文档失败: %w", err)
		}
	}

	// 2. 向量化
	texts := make([]string, len(docs))
	for i, doc := range docs {
		texts[i] = doc.Content
	}

	vectors, err := p.retriever.embedder.EmbedStrings(ctx, texts)
	if err != nil {
		return fmt.Errorf("向量化失败: %w", err)
	}

	// 3. 设置向量
	for i, doc := range docs {
		doc.Vector = vectors[i]
	}

	// 4. 存储
	if err := p.retriever.store.Add(ctx, docs); err != nil {
		return fmt.Errorf("存储失败: %w", err)
	}

	return nil
}

// Query 查询并构建上下文
func (p *RAGPipeline) Query(ctx context.Context, query string) (*RAGResult, error) {
	// 检索相关文档
	docs, err := p.retriever.Retrieve(ctx, query)
	if err != nil {
		return nil, err
	}

	// 构建上下文
	context := BuildContext(docs)

	return &RAGResult{
		Query:        query,
		Documents:    docs,
		Context:      context,
		HasKnowledge: len(docs) > 0,
	}, nil
}

// RAGResult RAG 查询结果
type RAGResult struct {
	Query        string
	Documents    []*Document
	Context      string
	HasKnowledge bool
}

// BuildContext 构建上下文
func BuildContext(docs []*Document) string {
	if len(docs) == 0 {
		return "（知识库未命中相关内容）"
	}

	var builder strings.Builder
	for i, doc := range docs {
		builder.WriteString(fmt.Sprintf("【参考资料 %d】(相关度: %.2f)\n", i+1, doc.Score))
		builder.WriteString(doc.Content)
		builder.WriteString("\n\n")
	}

	return builder.String()
}

// SimpleRAG 简化的 RAG 实现
type SimpleRAG struct {
	store    VectorStore
	embedder Embedder
	splitter DocumentSplitter
	topK     int
}

// NewSimpleRAG 创建简化 RAG
func NewSimpleRAG(embedder Embedder, topK int) *SimpleRAG {
	return &SimpleRAG{
		store:    NewMemoryVectorStore(),
		embedder: embedder,
		splitter: NewRecursiveSplitter(500, 50),
		topK:     topK,
	}
}

// AddText 添加文本到知识库
func (r *SimpleRAG) AddText(ctx context.Context, text string, metadata map[string]any) error {
	doc := &Document{
		ID:       generateID(text),
		Content:  text,
		MetaData: metadata,
	}

	// 分割
	docs, err := r.splitter.Split(ctx, []*Document{doc})
	if err != nil {
		return err
	}

	// 向量化
	texts := make([]string, len(docs))
	for i, d := range docs {
		texts[i] = d.Content
	}

	vectors, err := r.embedder.EmbedStrings(ctx, texts)
	if err != nil {
		return err
	}

	for i, d := range docs {
		d.Vector = vectors[i]
	}

	// 存储
	return r.store.Add(ctx, docs)
}

// Query 查询
func (r *SimpleRAG) Query(ctx context.Context, query string) (string, []*Document, error) {
	// 向量化查询
	vectors, err := r.embedder.EmbedStrings(ctx, []string{query})
	if err != nil {
		return "", nil, err
	}

	// 搜索
	docs, err := r.store.Search(ctx, vectors[0], r.topK, 0.3)
	if err != nil {
		return "", nil, err
	}

	// 构建上下文
	context := BuildContext(docs)
	return context, docs, nil
}

// GetStore 获取向量存储
func (r *SimpleRAG) GetStore() VectorStore {
	return r.store
}
