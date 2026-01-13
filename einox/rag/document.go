// Package rag 提供 RAG 检索增强生成组件
package rag

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/cloudwego/eino/schema"
)

// Document 文档结构
type Document struct {
	ID       string
	Content  string
	MetaData map[string]any
	Vector   []float64
	Score    float64
}

// ToSchemaDocument 转换为 schema.Document
func (d *Document) ToSchemaDocument() *schema.Document {
	return &schema.Document{
		ID:       d.ID,
		Content:  d.Content,
		MetaData: d.MetaData,
	}
}

// FromSchemaDocument 从 schema.Document 转换
func FromSchemaDocument(doc *schema.Document) *Document {
	return &Document{
		ID:       doc.ID,
		Content:  doc.Content,
		MetaData: doc.MetaData,
	}
}

// DocumentLoader 文档加载器接口
type DocumentLoader interface {
	Load(ctx context.Context, source string) ([]*Document, error)
}

// TextLoader 文本加载器
type TextLoader struct{}

func (l *TextLoader) Load(ctx context.Context, source string) ([]*Document, error) {
	// 直接将文本作为文档
	doc := &Document{
		ID:      generateID(source),
		Content: source,
		MetaData: map[string]any{
			"source": "text",
		},
	}
	return []*Document{doc}, nil
}

// DocumentSplitter 文档分割器接口
type DocumentSplitter interface {
	Split(ctx context.Context, docs []*Document) ([]*Document, error)
}

// RecursiveSplitter 递归分割器
type RecursiveSplitter struct {
	ChunkSize    int
	ChunkOverlap int
	Separators   []string
}

func NewRecursiveSplitter(chunkSize, overlap int) *RecursiveSplitter {
	return &RecursiveSplitter{
		ChunkSize:    chunkSize,
		ChunkOverlap: overlap,
		Separators:   []string{"\n\n", "\n", "。", ".", " "},
	}
}

func (s *RecursiveSplitter) Split(ctx context.Context, docs []*Document) ([]*Document, error) {
	var result []*Document

	for _, doc := range docs {
		chunks := s.splitText(doc.Content)
		for i, chunk := range chunks {
			newDoc := &Document{
				ID:      doc.ID + "_" + string(rune('a'+i)),
				Content: chunk,
				MetaData: map[string]any{
					"source":      doc.MetaData["source"],
					"chunk_index": i,
					"parent_id":   doc.ID,
				},
			}
			result = append(result, newDoc)
		}
	}

	return result, nil
}

func (s *RecursiveSplitter) splitText(text string) []string {
	if len(text) <= s.ChunkSize {
		return []string{text}
	}

	var chunks []string
	currentChunk := ""

	// 按段落分割
	paragraphs := strings.Split(text, "\n\n")

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		if len(currentChunk)+len(para)+2 <= s.ChunkSize {
			if currentChunk != "" {
				currentChunk += "\n\n"
			}
			currentChunk += para
		} else {
			if currentChunk != "" {
				chunks = append(chunks, currentChunk)
			}

			// 如果单个段落超过 ChunkSize，进一步分割
			if len(para) > s.ChunkSize {
				subChunks := s.splitLongParagraph(para)
				chunks = append(chunks, subChunks...)
				currentChunk = ""
			} else {
				currentChunk = para
			}
		}
	}

	if currentChunk != "" {
		chunks = append(chunks, currentChunk)
	}

	// 添加重叠
	if s.ChunkOverlap > 0 && len(chunks) > 1 {
		chunks = s.addOverlap(chunks)
	}

	return chunks
}

func (s *RecursiveSplitter) splitLongParagraph(para string) []string {
	var chunks []string
	sentences := strings.Split(para, "。")

	currentChunk := ""
	for _, sent := range sentences {
		sent = strings.TrimSpace(sent)
		if sent == "" {
			continue
		}
		sent += "。"

		if len(currentChunk)+len(sent) <= s.ChunkSize {
			currentChunk += sent
		} else {
			if currentChunk != "" {
				chunks = append(chunks, currentChunk)
			}
			currentChunk = sent
		}
	}

	if currentChunk != "" {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}

func (s *RecursiveSplitter) addOverlap(chunks []string) []string {
	result := make([]string, len(chunks))
	result[0] = chunks[0]

	for i := 1; i < len(chunks); i++ {
		prev := chunks[i-1]
		overlap := ""
		if len(prev) > s.ChunkOverlap {
			overlap = prev[len(prev)-s.ChunkOverlap:]
		} else {
			overlap = prev
		}
		result[i] = overlap + chunks[i]
	}

	return result
}

func generateID(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:8])
}
