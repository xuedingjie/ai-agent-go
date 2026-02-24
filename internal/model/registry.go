// Package model 定义模型接口和注册表模式
package model

import (
	"context"
	"fmt"
	"sync"
)

// Model Model模型接口
type Model interface {
	// Generate 生成文本响应
	Generate(ctx context.Context, prompt string) (string, error)

	// Name 返回模型名称
	Name() string

	// Config 返回模型配置
	Config() ModelConfig
}

// ModelConfig ModelConfig模型配置
type ModelConfig struct {
	Name        string  `json:"name"`
	APIKey      string  `json:"api_key"`
	APIEndpoint string  `json:"api_endpoint"`
	ModelID     string  `json:"model_id"`
	Timeout     int     `json:"timeout"` //秒
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

// ModelFactory ModelFactory模型工厂函数
type ModelFactory func(config ModelConfig) (Model, error)

// ModelRegistry ModelRegistry模型注册表
type ModelRegistry struct {
	factories map[string]ModelFactory
	models    map[string]Model
	mu        sync.RWMutex
}

// NewModelRegistry 创建新的模型注册表
func NewModelRegistry() *ModelRegistry {
	return &ModelRegistry{
		factories: make(map[string]ModelFactory),
		models:    make(map[string]Model),
	}
}

// Register 注册模型工厂
func (r *ModelRegistry) Register(name string, factory ModelFactory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[name]; exists {
		return fmt.Errorf("模型 %s已注册", name)
	}

	r.factories[name] = factory
	return nil
}

// CreateModel 创建模型实例
func (r *ModelRegistry) CreateModel(config ModelConfig) (Model, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	//检查是否已存在缓存的模型
	if model, exists := r.models[config.Name]; exists {
		return model, nil
	}

	// 获取工厂函数
	factory, exists := r.factories[config.ModelID]
	if !exists {
		// 如果没有找到特定模型的工厂，使用默认工厂
		factory = r.getDefaultFactory(config.ModelID)
		if factory == nil {
			return nil, fmt.Errorf("不支持的模型类型: %s", config.ModelID)
		}
	}

	// 创建模型实例
	model, err := factory(config)
	if err != nil {
		return nil, fmt.Errorf("创建模型失败: %w", err)
	}

	//缓模型模型实例
	r.models[config.Name] = model

	return model, nil
}

// GetModel 获取已创建的模型
func (r *ModelRegistry) GetModel(name string) (Model, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	model, exists := r.models[name]
	return model, exists
}

// ListModels列出所有已注册的模型类型
func (r *ModelRegistry) ListModels() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	models := make([]string, 0, len(r.factories))
	for name := range r.factories {
		models = append(models, name)
	}

	return models
}

// getDefaultFactory 获取默认工厂函数
func (r *ModelRegistry) getDefaultFactory(modelType string) ModelFactory {
	switch {
	case isLLaMA(modelType):
		return NewLLaMAModel
	case isQwen(modelType):
		return NewQwenModel
	case isOpenAI(modelType):
		return NewOpenAIModel
	default:
		return nil
	}
}

// isLLaMA检查是否为LLaMA模型
func isLLaMA(modelType string) bool {
	return modelType == "llama" || modelType == "llama2" || modelType == "llama3"
}

// isQwen检查是否为通义千问模型
func isQwen(modelType string) bool {
	return modelType == "qwen" || modelType == "qwen-turbo" || modelType == "qwen-plus"
}

// isOpenAI检查是否为OpenAI模型
func isOpenAI(modelType string) bool {
	return modelType == "gpt-3.5-turbo" || modelType == "gpt-4" || modelType == "gpt-4-turbo"
}

// GlobalRegistry全局模型注册表
var GlobalRegistry = NewModelRegistry()

// RegisterModel 注册模型到全局注册表
func RegisterModel(name string, factory ModelFactory) error {
	return GlobalRegistry.Register(name, factory)
}

// CreateModel 创建模型实例
func CreateModel(config ModelConfig) (Model, error) {
	return GlobalRegistry.CreateModel(config)
}
