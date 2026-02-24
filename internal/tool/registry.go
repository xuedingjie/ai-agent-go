// Package tool 实现可扩展的工具调用框架
package tool

import (
	"context"
	"fmt"
	"sync"
)

// Tool工具接口
type Tool interface {
	// Name工具名称
	Name() string
	
	// Description工具描述
	Description() string
	
	// Parameters工具参数定义
	Parameters() map[string]interface{}
	
	// Execute执行工具
	Execute(ctx context.Context, input string) (string, error)
}

// ToolFactory工具工厂函数
type ToolFactory func() Tool

// ToolRegistry工具注册表
type ToolRegistry struct {
	tools    map[string]Tool
	factories map[string]ToolFactory
	mu       sync.RWMutex
}

// NewToolRegistry 创建新的工具注册表
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools:    make(map[string]Tool),
		factories: make(map[string]ToolFactory),
	}
}

// Register 注册工具实例
func (r *ToolRegistry) Register(tool Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.tools[tool.Name()]; exists {
		return fmt.Errorf("工具 %s 已注册", tool.Name())
	}
	
	r.tools[tool.Name()] = tool
	return nil
}

// RegisterFactory 注册工具工厂
func (r *ToolRegistry) RegisterFactory(name string, factory ToolFactory) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.factories[name]; exists {
		return fmt.Errorf("工具工厂 %s已注册", name)
	}
	
	r.factories[name] = factory
	return nil
}

// GetTool 获取工具实例
func (r *ToolRegistry) GetTool(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	tool, exists := r.tools[name]
	return tool, exists
}

// CreateTool 通过工厂创建工具实例
func (r *ToolRegistry) CreateTool(name string) (Tool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	//检查是否已存在实例
	if tool, exists := r.tools[name]; exists {
		return tool, nil
	}
	
	// 通过工厂创建
	factory, exists := r.factories[name]
	if !exists {
		return nil, fmt.Errorf("未找到工具工厂: %s", name)
	}
	
	tool := factory()
	r.tools[name] = tool
	
	return tool, nil
}

// ListTools列出所有工具
func (r *ToolRegistry) ListTools() []ToolInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	tools := make([]ToolInfo, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, ToolInfo{
			Name:        tool.Name(),
			Description: tool.Description(),
			Parameters:  tool.Parameters(),
		})
	}
	
	return tools
}

// ToolInfo工具信息
type ToolInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// Manager工具管理器
type Manager struct {
	registry *ToolRegistry
}

// NewManager 创建工具管理器
func NewManager() *Manager {
	return &Manager{
		registry: NewToolRegistry(),
	}
}

// Register 注册工具
func (m *Manager) Register(tool Tool) error {
	return m.registry.Register(tool)
}

// RegisterFactory 注册工具工厂
func (m *Manager) RegisterFactory(name string, factory ToolFactory) error {
	return m.registry.RegisterFactory(name, factory)
}

// ExecuteTool执行工具
func (m *Manager) ExecuteTool(ctx context.Context, name string, input string) (string, error) {
	tool, err := m.registry.CreateTool(name)
	if err != nil {
		return "", fmt.Errorf("获取工具失败: %w", err)
	}
	
	result, err := tool.Execute(ctx, input)
	if err != nil {
		return "", fmt.Errorf("执行工具 %s失败: %w", name, err)
	}
	
	return result, nil
}

// ListTools列出所有工具
func (m *Manager) ListTools() []ToolInfo {
	return m.registry.ListTools()
}

// GetToolSchema 获取工具的JSON Schema
func (m *Manager) GetToolSchema(name string) (map[string]interface{}, error) {
	tool, exists := m.registry.GetTool(name)
	if !exists {
		return nil, fmt.Errorf("工具 %s 不存在", name)
	}
	
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type": "string",
				"const": tool.Name(),
			},
			"description": map[string]interface{}{
				"type": "string",
				"const": tool.Description(),
			},
			"parameters": tool.Parameters(),
		},
		"required": []string{"name"},
	}
	
	return schema, nil
}

// GlobalManager全局工具管理器
var GlobalManager = NewManager()

// RegisterTool 注册工具到全局管理器
func RegisterTool(tool Tool) error {
	return GlobalManager.Register(tool)
}

// RegisterToolFactory 注册工具工厂到全局管理器
func RegisterToolFactory(name string, factory ToolFactory) error {
	return GlobalManager.RegisterFactory(name, factory)
}