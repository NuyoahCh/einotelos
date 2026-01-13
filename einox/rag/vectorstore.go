// Package rag 提供向量存储组件
package rag

import (
	"context"
	"math"
	"sort"
	"sync"
)

// VectorStore 向量存储接口
type VectorStore interface {
	// Add 添加文档
	Add(ctx context.Context, docs []*Document) error
	// Search 搜索相似文档
	Search(ctx context.Context, query []float64, topK int, threshold float64) ([]*Document, error)
	// Delete 删除文档
	Delete(ctx context.Context, ids []string) error
	// Clear 清空存储
	Clear(ctx context.Context) error
}

// MemoryVectorStore 内存向量存储
type MemoryVectorStore struct {
	mu        sync.RWMutex
	documents map[string]*Document
}

// NewMemoryVectorStore 创建内存向量存储
func NewMemoryVectorStore() *MemoryVectorStore {
	return &MemoryVectorStore{
		documents: make(map[string]*Document),
	}
}

// Add 添加文档
func (s *MemoryVectorStore) Add(ctx context.Context, docs []*Document) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, doc := range docs {
		s.documents[doc.ID] = doc
	}
	return nil
}

// Search 搜索相似文档
func (s *MemoryVectorStore) Search(ctx context.Context, query []float64, topK int, threshold float64) ([]*Document, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	type scored struct {
		doc   *Document
		score float64
	}

	var results []scored

	for _, doc := range s.documents {
		if len(doc.Vector) == 0 {
			continue
		}

		score := cosineSimilarity(query, doc.Vector)
		if score >= threshold {
			results = append(results, scored{doc: doc, score: score})
		}
	}

	// 按分数降序排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	// 取 topK
	if len(results) > topK {
		results = results[:topK]
	}

	// 转换结果
	docs := make([]*Document, len(results))
	for i, r := range results {
		doc := *r.doc // 复制
		doc.Score = r.score
		docs[i] = &doc
	}

	return docs, nil
}

// Delete 删除文档
func (s *MemoryVectorStore) Delete(ctx context.Context, ids []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, id := range ids {
		delete(s.documents, id)
	}
	return nil
}

// Clear 清空存储
func (s *MemoryVectorStore) Clear(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.documents = make(map[string]*Document)
	return nil
}

// Count 获取文档数量
func (s *MemoryVectorStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.documents)
}

// cosineSimilarity 计算余弦相似度
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// Embedder 向量化接口
type Embedder interface {
	EmbedStrings(ctx context.Context, texts []string) ([][]float64, error)
}

// MockEmbedder 模拟向量化器（用于测试）
type MockEmbedder struct {
	Dimension int
}

func NewMockEmbedder(dim int) *MockEmbedder {
	return &MockEmbedder{Dimension: dim}
}

func (e *MockEmbedder) EmbedStrings(ctx context.Context, texts []string) ([][]float64, error) {
	vectors := make([][]float64, len(texts))
	for i, text := range texts {
		vectors[i] = e.generateVector(text)
	}
	return vectors, nil
}

func (e *MockEmbedder) generateVector(text string) []float64 {
	// 简单的模拟向量生成
	vec := make([]float64, e.Dimension)
	for i := 0; i < e.Dimension && i < len(text); i++ {
		vec[i] = float64(text[i]) / 255.0
	}
	// 归一化
	var norm float64
	for _, v := range vec {
		norm += v * v
	}
	if norm > 0 {
		norm = math.Sqrt(norm)
		for i := range vec {
			vec[i] /= norm
		}
	}
	return vec
}
