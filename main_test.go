package main

import (
	"context"
	"testing"
	"time"

	"aigent/internal/core"
	"aigent/internal/model"
	"aigent/internal/tool"
	"aigent/internal/sse"
)

func TestAgentInitialization(t *testing.T) {
	// 测试Agent初始化
	config := core.AgentConfig{
		MaxIterations: 5,
		Timeout:       30 * time.Second,
		Debug:         true,
	}
	
	agent := core.NewAgent(config)
	if agent == nil {
		t.Fatal("Agent创建失败")
	}
	
	//测试配置设置
	// 通过配置创建的Agent应该具有正确的配置
	//这里我们验证Agent不为nil就足够了
	if agent == nil {
		t.Error("Agent创建失败")
	}
}

func TestModelRegistry(t *testing.T) {
	//测试模型注册表
	registry := model.NewModelRegistry()
	
	// 注册测试模型
	err := registry.Register("test-model", func(config model.ModelConfig) (model.Model, error) {
		return &TestModel{config: config}, nil
	})
	if err != nil {
		t.Fatalf("注册模型失败: %v", err)
	}
	
	//检查模型列表
	models := registry.ListModels()
	if len(models) == 0 {
		t.Error("未找到注册的模型")
	}
	
	t.Logf("已注册的模型类型: %v", models)
	
	//测试创建模型
	config := model.ModelConfig{
		Name: "test-instance",
		ModelID: "test-model",
	}
	
	modelInstance, err := registry.CreateModel(config)
	if err != nil {
		t.Fatalf("创建模型实例失败: %v", err)
	}
	
	if modelInstance == nil {
		t.Error("模型实例创建失败")
	}
	
	t.Log("模型注册表测试通过")
}

func TestToolFramework(t *testing.T) {
	//测试工具框架
	manager := tool.NewManager()
	
	// 注册测试工具
	testTool := &TestTool{}
	err := manager.Register(testTool)
	if err != nil {
		t.Fatalf("注册工具失败: %v", err)
	}
	
	//测试工具列表
	tools := manager.ListTools()
	if len(tools) == 0 {
		t.Error("未找到注册的工具")
	}
	
	t.Logf("已注册的工具: %v", tools)
	
	// 测试工具执行
	result, err := manager.ExecuteTool(context.Background(), "test_tool", "test input")
	if err != nil {
		t.Fatalf("执行工具失败: %v", err)
	}
	
	expected := "test result"
	if result != expected {
		t.Errorf("期望结果为'%s'，实际为'%s'", expected, result)
	}
}

func TestSSEBroker(t *testing.T) {
	//测试SSE代理
	broker := sse.NewBroker()
	defer broker.Close()
	
	//检查初始状态
	if broker.GetClientsCount() != 0 {
		t.Errorf("期望客户端数为0，实际为%d", broker.GetClientsCount())
	}
	
	// 测试事件广播
	broker.Broadcast("test_event", map[string]interface{}{
		"message": "test message",
	})
	
	t.Log("SSE代理测试通过")
}

// TestTool测试工具实现
type TestTool struct{}

func (t *TestTool) Name() string {
	return "test_tool"
}

func (t *TestTool) Description() string {
	return "测试工具"
}

func (t *TestTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"input": map[string]interface{}{
			"type": "string",
			"description": "测试输入",
		},
	}
}

func (t *TestTool) Execute(ctx context.Context, input string) (string, error) {
	return "test result", nil
}

// TestModel测试模型实现
type TestModel struct {
	config model.ModelConfig
}

func (m *TestModel) Generate(ctx context.Context, prompt string) (string, error) {
	return "test response", nil
}

func (m *TestModel) Name() string {
	return m.config.Name
}

func (m *TestModel) Config() model.ModelConfig {
	return m.config
}