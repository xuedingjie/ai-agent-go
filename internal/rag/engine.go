// Package rag 实现RAG向量检索功能
package rag

import (
	"context"
	"fmt"
	"math"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Document文档结构
type Document struct {
	ID       string  `json:"id"`
	Content  string  `json:"content"`
	Metadata string  `json:"metadata"`
	Embedding []float32 `json:"embedding,omitempty"`
}

// SearchResult搜索结果
type SearchResult struct {
	Document  Document `json:"document"`
	Similarity float64  `json:"similarity"`
}

// EmbeddingModel嵌入模型接口
type EmbeddingModel interface {
	// Embed生成文本嵌入向量
	Embed(ctx context.Context, text string) ([]float32, error)
	
	// Name模型名称
	Name() string
}

// Engine RAG引擎
type Engine struct {
	dbPool         *pgxpool.Pool
	embeddingModel EmbeddingModel
	dimensions     int
	mu             sync.RWMutex
}

// Config RAG配置
type Config struct {
	DatabaseURL    string
	EmbeddingModel EmbeddingModel
	Dimensions     int
	TableName      string
}

// NewEngine创建新的RAG引擎
func NewEngine(config Config) (*Engine, error) {
	if config.DatabaseURL == "" {
		return nil, fmt.Errorf("数据库URL不能为空")
	}
	
	if config.EmbeddingModel == nil {
		return nil, fmt.Errorf("嵌入模型不能为空")
	}
	
	if config.Dimensions <= 0 {
		config.Dimensions = 1536 // 默认维度
	}
	
	// 创建数据库连接池
	pool, err := createDBPool(config.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("创建数据库连接失败: %w", err)
	}
	
	// 初始化表结构
	if err := initSchema(pool, config.TableName); err != nil {
		pool.Close()
		return nil, fmt.Errorf("初始化数据库表失败: %w", err)
	}
	
	engine := &Engine{
		dbPool:         pool,
		embeddingModel: config.EmbeddingModel,
		dimensions:     config.Dimensions,
	}
	
	return engine, nil
}

// createDBPool创建数据库连接池
func createDBPool(databaseURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("解析数据库配置失败: %w", err)
	}
	
	config.MaxConns = 20
	config.MinConns = 5
	
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("创建连接池失败: %w", err)
	}
	
	return pool, nil
}

// initSchema初始化数据库表结构
func initSchema(pool *pgxpool.Pool, tableName string) error {
	if tableName == "" {
		tableName = "documents"
	}
	
	// 创建表和索引的SQL
	schemaSQL := fmt.Sprintf(`
		CREATE EXTENSION IF NOT EXISTS vector;
		
		CREATE TABLE IF NOT EXISTS %s (
			id TEXT PRIMARY KEY,
			content TEXT NOT NULL,
			metadata JSONB,
			embedding vector(%d),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS idx_%s_embedding ON %s 
		USING ivfflat (embedding vector_cosine_ops);
	`, tableName, 1536, tableName, tableName)
	
	_, err := pool.Exec(context.Background(), schemaSQL)
	if err != nil {
		return fmt.Errorf("执行数据库表初始化失败: %w", err)
	}
	
	return nil
}

// AddDocument添加文档
func (e *Engine) AddDocument(ctx context.Context, doc Document) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	// 生成嵌入向量
	embedding, err := e.embeddingModel.Embed(ctx, doc.Content)
	if err != nil {
		return fmt.Errorf("生成文档嵌入向量失败: %w", err)
	}
	
	doc.Embedding = embedding
	
	//插入数据库
	tableName := "documents"
	_, err = e.dbPool.Exec(ctx, 
		fmt.Sprintf("INSERT INTO %s (id, content, metadata, embedding) VALUES ($1, $2, $3, $4)", tableName),
		doc.ID, doc.Content, doc.Metadata, embedding)
	
	if err != nil {
		return fmt.Errorf("插入文档到数据库失败: %w", err)
	}
	
	return nil
}

// AddDocuments批量添加文档
func (e *Engine) AddDocuments(ctx context.Context, docs []Document) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	tableName := "documents"
	
	//批处理
	batch := &pgx.Batch{}
	
	for _, doc := range docs {
		// 生成嵌入向量
		embedding, err := e.embeddingModel.Embed(ctx, doc.Content)
		if err != nil {
			return fmt.Errorf("生成文档 %s的嵌入向量失败: %w", doc.ID, err)
		}
		
		doc.Embedding = embedding
		batch.Queue(fmt.Sprintf("INSERT INTO %s (id, content, metadata, embedding) VALUES ($1, $2, $3, $4)", tableName),
			doc.ID, doc.Content, doc.Metadata, embedding)
	}
	
	br := e.dbPool.SendBatch(ctx, batch)
	defer br.Close()
	
	for i := 0; i < batch.Len(); i++ {
		_, err := br.Exec()
		if err != nil {
			return fmt.Errorf("批量插入文档失败: %w", err)
		}
	}
	
	return nil
}

// Search向量检索
func (e *Engine) Search(ctx context.Context, query string, topK int) ([]SearchResult, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	if topK <= 0 {
		topK = 5
	}
	
	// 生成查询向量
	queryEmbedding, err := e.embeddingModel.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("生成查询嵌入向量失败: %w", err)
	}
	
	//执行向量相似度搜索
	tableName := "documents"
	rows, err := e.dbPool.Query(ctx, 
		fmt.Sprintf("SELECT id, content, metadata, embedding <=> $1 AS similarity FROM %s ORDER BY embedding <=> $1 LIMIT $2", tableName),
		queryEmbedding, topK)
	
	if err != nil {
		return nil, fmt.Errorf("执行向量检索失败: %w", err)
	}
	defer rows.Close()
	
	results := []SearchResult{}
	for rows.Next() {
		var doc Document
		var similarity float64
		
		err := rows.Scan(&doc.ID, &doc.Content, &doc.Metadata, &similarity)
		if err != nil {
			return nil, fmt.Errorf("扫描检索结果失败: %w", err)
		}
		
		results = append(results, SearchResult{
			Document:  doc,
			Similarity: similarity,
		})
	}
	
	return results, nil
}

// DeleteDocument删除文档
func (e *Engine) DeleteDocument(ctx context.Context, id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	tableName := "documents"
	_, err := e.dbPool.Exec(ctx, fmt.Sprintf("DELETE FROM %s WHERE id = $1", tableName), id)
	if err != nil {
		return fmt.Errorf("删除文档失败: %w", err)
	}
	
	return nil
}

// UpdateDocument更新文档
func (e *Engine) UpdateDocument(ctx context.Context, doc Document) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	// 重新生成嵌入向量
	embedding, err := e.embeddingModel.Embed(ctx, doc.Content)
	if err != nil {
		return fmt.Errorf("生成文档嵌入向量失败: %w", err)
	}
	
	doc.Embedding = embedding
	
	// 更新数据库
	tableName := "documents"
	_, err = e.dbPool.Exec(ctx, 
		fmt.Sprintf("UPDATE %s SET content = $1, metadata = $2, embedding = $3 WHERE id = $4", tableName),
		doc.Content, doc.Metadata, embedding, doc.ID)
	
	if err != nil {
		return fmt.Errorf("更新文档失败: %w", err)
	}
	
	return nil
}

// GetDocument获取文档
func (e *Engine) GetDocument(ctx context.Context, id string) (*Document, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	tableName := "documents"
	row := e.dbPool.QueryRow(ctx, 
		fmt.Sprintf("SELECT id, content, metadata FROM %s WHERE id = $1", tableName), id)
	
	var doc Document
	err := row.Scan(&doc.ID, &doc.Content, &doc.Metadata)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("文档不存在: %s", id)
		}
		return nil, fmt.Errorf("获取文档失败: %w", err)
	}
	
	return &doc, nil
}

// ListDocuments列出文档
func (e *Engine) ListDocuments(ctx context.Context, limit, offset int) ([]Document, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	if limit <= 0 {
		limit = 10
	}
	
	tableName := "documents"
	rows, err := e.dbPool.Query(ctx, 
		fmt.Sprintf("SELECT id, content, metadata FROM %s ORDER BY created_at DESC LIMIT $1 OFFSET $2", tableName),
		limit, offset)
	
	if err != nil {
		return nil, fmt.Errorf("列出文档失败: %w", err)
	}
	defer rows.Close()
	
	docs := []Document{}
	for rows.Next() {
		var doc Document
		err := rows.Scan(&doc.ID, &doc.Content, &doc.Metadata)
		if err != nil {
			return nil, fmt.Errorf("扫描文档失败: %w", err)
		}
		docs = append(docs, doc)
	}
	
	return docs, nil
}

// Close关闭引擎
func (e *Engine) Close() {
	if e.dbPool != nil {
		e.dbPool.Close()
	}
}

// MockEmbeddingModel模拟嵌入模型（用于测试）
type MockEmbeddingModel struct {
	name string
}

// NewMockEmbeddingModel创建模拟嵌入模型
func NewMockEmbeddingModel() *MockEmbeddingModel {
	return &MockEmbeddingModel{name: "mock-embedding"}
}

// Embed生成模拟嵌入向量
func (m *MockEmbeddingModel) Embed(ctx context.Context, text string) ([]float32, error) {
	//简单的模拟实现：基于文本哈希生成固定维度的向量
	hash := simpleHash(text)
	embedding := make([]float32, 1536)
	
	//基于哈希值生成向量
	for i := 0; i < len(embedding); i++ {
		embedding[i] = float32((hash + int64(i)) % 1000) / 1000.0
	}
	
	return embedding, nil
}

// Name模型名称
func (m *MockEmbeddingModel) Name() string {
	return m.name
}

// simpleHash简单的哈希函数
func simpleHash(text string) int64 {
	var hash int64 = 0
	for _, char := range text {
		hash = (hash*31 + int64(char)) % 1000000007
	}
	return hash
}

// CosineSimilarity计算余弦相似度
func CosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}
	
	var dotProduct float64
	var normA, normB float64
	
	for i := 0; i < len(a); i++ {
		dotProduct += float64(a[i] * b[i])
		normA += float64(a[i] * a[i])
		normB += float64(b[i] * b[i])
	}
	
	if normA == 0 || normB == 0 {
		return 0
	}
	
	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}