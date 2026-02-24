package model

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenAIModel OpenAI模型实现
type OpenAIModel struct {
	config ModelConfig
	client *http.Client
}

// NewOpenAIModel 创建OpenAI模型
func NewOpenAIModel(config ModelConfig) (Model, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}
	
	if config.ModelID == "" {
		config.ModelID = "gpt-3.5-turbo"
	}
	
	client := &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Second,
	}
	
	return &OpenAIModel{
		config: config,
		client: client,
	}, nil
}

// Generate 生成文本响应
func (m *OpenAIModel) Generate(ctx context.Context, prompt string) (string, error) {
	request := OpenAIRequest{
		Model: m.config.ModelID,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   m.config.MaxTokens,
		Temperature: m.config.Temperature,
	}
	
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", 
		"https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.config.APIKey)
	
	resp, err := m.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API请求失败: %s - %s", resp.Status, string(body))
	}
	
	var response OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}
	
	if len(response.Choices) == 0 {
		return "", fmt.Errorf("API返回空响应")
	}
	
	return response.Choices[0].Message.Content, nil
}

// Name 返回模型名称
func (m *OpenAIModel) Name() string {
	return m.config.Name
}

// Config 返回模型配置
func (m *OpenAIModel) Config() ModelConfig {
	return m.config
}

// OpenAIRequest OpenAI API请求结构
type OpenAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

// Message消息结构
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse OpenAI API响应结构
type OpenAIResponse struct {
	Choices []Choice `json:"choices"`
}

// Choice 选择项
type Choice struct {
	Message Message `json:"message"`
}

// QwenModel 通义千问模型实现
type QwenModel struct {
	config ModelConfig
	client *http.Client
}

// NewQwenModel 创建通义千问模型
func NewQwenModel(config ModelConfig) (Model, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("通义千问 API key is required")
	}
	
	if config.APIEndpoint == "" {
		config.APIEndpoint = "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"
	}
	
	if config.ModelID == "" {
		config.ModelID = "qwen-turbo"
	}
	
	client := &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Second,
	}
	
	return &QwenModel{
		config: config,
		client: client,
	}, nil
}

// Generate 生成文本响应
func (m *QwenModel) Generate(ctx context.Context, prompt string) (string, error) {
	request := QwenRequest{
		Model: m.config.ModelID,
		Input: QwenInput{
			Prompt: prompt,
		},
		Parameters: QwenParameters{
			MaxTokens:   m.config.MaxTokens,
			Temperature: m.config.Temperature,
		},
	}
	
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", m.config.APIEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.config.APIKey)
	req.Header.Set("X-DashScope-SSE", "enable")
	
	resp, err := m.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API请求失败: %s - %s", resp.Status, string(body))
	}
	
	var response QwenResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}
	
	if response.Output.Text == "" {
		return "", fmt.Errorf("API返回空响应")
	}
	
	return response.Output.Text, nil
}

// Name 返回模型名称
func (m *QwenModel) Name() string {
	return m.config.Name
}

// Config 返回模型配置
func (m *QwenModel) Config() ModelConfig {
	return m.config
}

// QwenRequest 通义千问API请求结构
type QwenRequest struct {
	Model      string          `json:"model"`
	Input      QwenInput       `json:"input"`
	Parameters QwenParameters  `json:"parameters"`
}

// QwenInput 输入参数
type QwenInput struct {
	Prompt string `json:"prompt"`
}

// QwenParameters 参数配置
type QwenParameters struct {
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

// QwenResponse 通义千问API响应结构
type QwenResponse struct {
	Output QwenOutput `json:"output"`
}

// QwenOutput 输出结果
type QwenOutput struct {
	Text string `json:"text"`
}

// LLaMAModel LLaMA模型实现（本地模型示例）
type LLaMAModel struct {
	config ModelConfig
}

// NewLLaMAModel 创建LLaMA模型
func NewLLaMAModel(config ModelConfig) (Model, error) {
	if config.APIEndpoint == "" {
		config.APIEndpoint = "http://localhost:8000/v1/completions"
	}
	
	if config.ModelID == "" {
		config.ModelID = "llama"
	}
	
	return &LLaMAModel{
		config: config,
	}, nil
}

// Generate 生成文本响应
func (m *LLaMAModel) Generate(ctx context.Context, prompt string) (string, error) {
	//这里是本地LLaMA模型的示例实现
	// 实际使用时需要连接到本地运行的LLaMA服务
	
	request := LLaMARequest{
		Prompt:      prompt,
		MaxTokens:   m.config.MaxTokens,
		Temperature: m.config.Temperature,
	}
	
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}
	
	client := &http.Client{
		Timeout: time.Duration(m.config.Timeout) * time.Second,
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", m.config.APIEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API请求失败: %s - %s", resp.Status, string(body))
	}
	
	var response LLaMAResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}
	
	if len(response.Choices) == 0 {
		return "", fmt.Errorf("API返回空响应")
	}
	
	return response.Choices[0].Text, nil
}

// Name 返回模型名称
func (m *LLaMAModel) Name() string {
	return m.config.Name
}

// Config 返回模型配置
func (m *LLaMAModel) Config() ModelConfig {
	return m.config
}

// LLaMARequest LLaMA API请求结构
type LLaMARequest struct {
	Prompt      string  `json:"prompt"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

// LLaMAResponse LLaMA API响应结构
type LLaMAResponse struct {
	Choices []LLaMAChoice `json:"choices"`
}

// LLaMAChoice 选择项
type LLaMAChoice struct {
	Text string `json:"text"`
}

// 初始化时注册默认模型
func init() {
	// 注册OpenAI模型
	RegisterModel("openai", NewOpenAIModel)
	RegisterModel("gpt-3.5-turbo", NewOpenAIModel)
	RegisterModel("gpt-4", NewOpenAIModel)
	RegisterModel("gpt-4-turbo", NewOpenAIModel)
	
	// 注册通义千问模型
	RegisterModel("qwen", NewQwenModel)
	RegisterModel("qwen-turbo", NewQwenModel)
	RegisterModel("qwen-plus", NewQwenModel)
	
	// 注册LLaMA模型
	RegisterModel("llama", NewLLaMAModel)
	RegisterModel("llama2", NewLLaMAModel)
	RegisterModel("llama3", NewLLaMAModel)
}