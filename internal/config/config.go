// Package config配置管理
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"aigent/internal/model"
	"aigent/internal/core"
	"aigent/internal/sse"
	"aigent/internal/http"
)

// Config应用配置
type Config struct {
	Server     ServerConfig     `json:"server"`
	Agent      AgentConfig      `json:"agent"`
	Models     []ModelConfig    `json:"models"`
	Database   DatabaseConfig   `json:"database"`
	Logging    LoggingConfig    `json:"logging"`
	Features   FeaturesConfig   `json:"features"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         string `json:"port"`
	Host         string `json:"host"`
	ReadTimeout  int    `json:"read_timeout"`
	WriteTimeout int    `json:"write_timeout"`
	IdleTimeout  int    `json:"idle_timeout"`
}

// AgentConfig Agent配置
type AgentConfig struct {
	MaxIterations int           `json:"max_iterations"`
	Timeout       time.Duration `json:"timeout"`
	Debug         bool          `json:"debug"`
}

// ModelConfig模型配置
type ModelConfig struct {
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	APIKey      string  `json:"api_key"`
	APIEndpoint string  `json:"api_endpoint"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
	Timeout     int     `json:"timeout"`
	Enabled     bool    `json:"enabled"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	URL      string `json:"url"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	SSLMode  string `json:"ssl_mode"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"`
	Output     string `json:"output"`
	MaxSize    int    `json:"max_size"`
	MaxAge     int    `json:"max_age"`
	MaxBackups int    `json:"max_backups"`
	Compress   bool   `json:"compress"`
}

// FeaturesConfig功能配置
type FeaturesConfig struct {
	EnableRAG     bool `json:"enable_rag"`
	EnableTools   bool `json:"enable_tools"`
	EnableSSE     bool `json:"enable_sse"`
	EnableMetrics bool `json:"enable_metrics"`
}

// LoadConfig 加载配置
func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = "config.json"
	}
	
	//检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 如果配置文件不存在，创建默认配置
		defaultConfig := GetDefaultConfig()
		if err := SaveConfig(defaultConfig, configPath); err != nil {
			return nil, fmt.Errorf("创建默认配置文件失败: %w", err)
		}
		return defaultConfig, nil
	}
	
	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}
	
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}
	
	//验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}
	
	return &config, nil
}

// SaveConfig 保存配置
func SaveConfig(config *Config, configPath string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}
	
	// 创建目录（如果不存在）
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}
	
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}
	
	return nil
}

// GetDefaultConfig 获取默认配置
func GetDefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         "8080",
			Host:         "localhost",
			ReadTimeout:  30,
			WriteTimeout: 30,
			IdleTimeout:  120,
		},
		Agent: AgentConfig{
			MaxIterations: 10,
			Timeout:       300 * time.Second,
			Debug:         false,
		},
		Models: []ModelConfig{
			{
				Name:        "default-openai",
				Type:        "gpt-3.5-turbo",
				APIKey:      "your-openai-api-key",
				MaxTokens:   2000,
				Temperature: 0.7,
				Timeout:     300,
				Enabled:     false,
			},
			{
				Name:        "default-qwen",
				Type:        "qwen-turbo",
				APIKey:      "your-qwen-api-key",
				MaxTokens:   2000,
				Temperature: 0.7,
				Timeout:     300,
				Enabled:     false,
			},
		},
		Database: DatabaseConfig{
			URL:      "",
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "postgres",
			Database: "aigent",
			SSLMode:  "disable",
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "text",
			Output:     "stdout",
			MaxSize:    100,
			MaxAge:     30,
			MaxBackups: 3,
			Compress:   true,
		},
		Features: FeaturesConfig{
			EnableRAG:     false,
			EnableTools:   true,
			EnableSSE:     true,
			EnableMetrics: false,
		},
	}
}

// Validate验证配置
func (c *Config) Validate() error {
	if c.Server.Port == "" {
		return fmt.Errorf("服务器端口不能为空")
	}
	
	if c.Agent.MaxIterations <= 0 {
		return fmt.Errorf("最大迭代次数必须大于0")
	}
	
	if c.Agent.Timeout <= 0 {
		return fmt.Errorf("超时时间必须大于0")
	}
	
	// 验证模型配置
	for i, model := range c.Models {
		if model.Name == "" {
			return fmt.Errorf("第%d个模型名称不能为空", i+1)
		}
		if model.Type == "" {
			return fmt.Errorf("第%d个模型类型不能为空", i+1)
		}
	}
	
	//验证数据库配置（如果启用了RAG）
	if c.Features.EnableRAG {
		if c.Database.URL == "" && c.Database.Host == "" {
			return fmt.Errorf("启用RAG时必须配置数据库连接")
		}
	}
	
	return nil
}

// GetDatabaseURL 获取数据库连接URL
func (c *Config) GetDatabaseURL() string {
	if c.Database.URL != "" {
		return c.Database.URL
	}
	
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Database,
		c.Database.SSLMode)
}

// GetModelConfigs 获取启用的模型配置
func (c *Config) GetModelConfigs() []model.ModelConfig {
	var configs []model.ModelConfig
	
	for _, modelConfig := range c.Models {
		if !modelConfig.Enabled {
			continue
		}
		
		config := model.ModelConfig{
			Name:        modelConfig.Name,
			ModelID:     modelConfig.Type,
			APIKey:      modelConfig.APIKey,
			APIEndpoint: modelConfig.APIEndpoint,
			MaxTokens:   modelConfig.MaxTokens,
			Temperature: modelConfig.Temperature,
			Timeout:     modelConfig.Timeout,
		}
		
		if config.Timeout <= 0 {
			config.Timeout = 300
		}
		
		configs = append(configs, config)
	}
	
	return configs
}

// ToCoreAgentConfig转为Core Agent配置
func (c *Config) ToCoreAgentConfig() core.AgentConfig {
	return core.AgentConfig{
		MaxIterations: c.Agent.MaxIterations,
		Timeout:       c.Agent.Timeout,
		Debug:         c.Agent.Debug,
	}
}

// ToHTTPServerConfig转为HTTP服务器配置
func (c *Config) ToHTTPServerConfig() http.Config {
	return http.Config{
		Port:  c.Server.Port,
		Debug: c.Agent.Debug,
	}
}

// ToSSEConfig转换为SSE配置
func (c *Config) ToSSEConfig() *sse.Broker {
	if !c.Features.EnableSSE {
		return nil
	}
	return sse.NewBroker()
}

// LoadFromEnvironment 从环境变量加载配置
func LoadFromEnvironment() *Config {
	config := GetDefaultConfig()
	
	// 服务器配置
	if port := os.Getenv("SERVER_PORT"); port != "" {
		config.Server.Port = port
	}
	
	if host := os.Getenv("SERVER_HOST"); host != "" {
		config.Server.Host = host
	}
	
	// Agent配置
	if maxIter := os.Getenv("AGENT_MAX_ITERATIONS"); maxIter != "" {
		if iter, err := getEnvInt(maxIter); err == nil {
			config.Agent.MaxIterations = iter
		}
	}
	
	if timeout := os.Getenv("AGENT_TIMEOUT"); timeout != "" {
		if t, err := getEnvInt(timeout); err == nil {
			config.Agent.Timeout = time.Duration(t) * time.Second
		}
	}
	
	if debug := os.Getenv("AGENT_DEBUG"); debug != "" {
		config.Agent.Debug = strings.ToLower(debug) == "true"
	}
	
	// 数据库配置
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		config.Database.URL = dbURL
	}
	
	if dbHost := os.Getenv("DATABASE_HOST"); dbHost != "" {
		config.Database.Host = dbHost
	}
	
	if dbPort := os.Getenv("DATABASE_PORT"); dbPort != "" {
		if port, err := getEnvInt(dbPort); err == nil {
			config.Database.Port = port
		}
	}
	
	if dbUser := os.Getenv("DATABASE_USER"); dbUser != "" {
		config.Database.User = dbUser
	}
	
	if dbPass := os.Getenv("DATABASE_PASSWORD"); dbPass != "" {
		config.Database.Password = dbPass
	}
	
	if dbName := os.Getenv("DATABASE_NAME"); dbName != "" {
		config.Database.Database = dbName
	}
	
	//功能配置
	if enableRAG := os.Getenv("ENABLE_RAG"); enableRAG != "" {
		config.Features.EnableRAG = strings.ToLower(enableRAG) == "true"
	}
	
	if enableTools := os.Getenv("ENABLE_TOOLS"); enableTools != "" {
		config.Features.EnableTools = strings.ToLower(enableTools) == "true"
	}
	
	if enableSSE := os.Getenv("ENABLE_SSE"); enableSSE != "" {
		config.Features.EnableSSE = strings.ToLower(enableSSE) == "true"
	}
	
	return config
}

// getEnvInt 获取环境变量整数值
func getEnvInt(value string) (int, error) {
	//简单的字符串转整数实现
	result := 0
	for _, char := range value {
		if char < '0' || char > '9' {
			return 0, fmt.Errorf("无效的数字格式")
		}
		result = result*10 + int(char-'0')
	}
	return result, nil
}

// MergeConfig合配置（优先级：环境变量 >配置文件 > 默认值）
func MergeConfig(configFile string) (*Config, error) {
	// 1. 加载配置文件
	fileConfig, err := LoadConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("加载配置文件失败: %w", err)
	}
	
	// 2. 加载环境变量配置
	envConfig := LoadFromEnvironment()
	
	// 3.合配置（环境变量优先）
	merged := &Config{
		Server: ServerConfig{
			Port:         getEnvOrDefault(envConfig.Server.Port, fileConfig.Server.Port),
			Host:         getEnvOrDefault(envConfig.Server.Host, fileConfig.Server.Host),
			ReadTimeout:  getEnvOrDefaultInt(envConfig.Server.ReadTimeout, fileConfig.Server.ReadTimeout),
			WriteTimeout: getEnvOrDefaultInt(envConfig.Server.WriteTimeout, fileConfig.Server.WriteTimeout),
			IdleTimeout:  getEnvOrDefaultInt(envConfig.Server.IdleTimeout, fileConfig.Server.IdleTimeout),
		},
		Agent: AgentConfig{
			MaxIterations: getEnvOrDefaultInt(envConfig.Agent.MaxIterations, fileConfig.Agent.MaxIterations),
			Timeout:      getEnvOrDefaultDuration(envConfig.Agent.Timeout, fileConfig.Agent.Timeout),
			Debug:        envConfig.Agent.Debug || fileConfig.Agent.Debug,
		},
		Models:   fileConfig.Models, //模型配置通常在配置文件中定义
		Database: fileConfig.Database,
		Logging:  fileConfig.Logging,
		Features: FeaturesConfig{
			EnableRAG:     envConfig.Features.EnableRAG || fileConfig.Features.EnableRAG,
			EnableTools:   envConfig.Features.EnableTools || fileConfig.Features.EnableTools,
			EnableSSE:     envConfig.Features.EnableSSE || fileConfig.Features.EnableSSE,
			EnableMetrics: envConfig.Features.EnableMetrics || fileConfig.Features.EnableMetrics,
		},
	}
	
	return merged, nil
}

// getEnvOrDefault 获取环境变量值或默认值
func getEnvOrDefault(envValue, defaultValue string) string {
	if envValue != "" {
		return envValue
	}
	return defaultValue
}

// getEnvOrDefaultInt 获取环境变量整数值或默认值
func getEnvOrDefaultInt(envValue, defaultValue int) int {
	if envValue > 0 {
		return envValue
	}
	return defaultValue
}

// getEnvOrDefaultDuration 获取环境变量时间值或默认值
func getEnvOrDefaultDuration(envValue, defaultValue time.Duration) time.Duration {
	if envValue > 0 {
		return envValue
	}
	return defaultValue
}