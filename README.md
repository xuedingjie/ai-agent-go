# AI Agent 🤖

[![Go Report Card](https://goreportcard.com/badge/github.com/xuedingjie/ai-agent-go)](https://goreportcard.com/report/github.com/xuedingjie/ai-agent-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/xuedingjie/ai-agent-go)](https://golang.org/doc/go1.19)

一个用Go语言实现的企业级智能AI代理系统，具备自主决策的Think-Execute循环、可扩展工具调用框架、RAG向量检索、多模型切换架构和SSE实时推送功能。

## 🌟 核心特性

- **🧠 自主思考决策**：基于Think-Execute循环的智能决策系统
- **🔧 可扩展工具框架**：支持自定义工具开发和动态注册
- **📚 RAG向量检索**：基于pgvector的语义搜索和知识库管理
- **🔄 多模型支持**：统一接口支持多种大语言模型
- **📡 实时状态推送**：SSE技术实现执行过程可视化
- **🛡️ 智能错误恢复**：自动错误检测和恢复机制
- **🔌 RESTful API**：完整的HTTP服务接口

## 🚀 核心功能详解

### 🧠 Think-Execute自主决策循环
- **智能分析**：基于上下文的深度问题分析
- **计划生成**：自动生成可执行的步骤计划
- **迭代优化**：多轮思考和执行优化
- **错误恢复**：智能错误检测和自动恢复
- **相关性验证**：确保执行计划与目标一致

### 🛠️ 可扩展工具调用框架
- **动态注册**：运行时注册和管理工具
- **内置工具**：网络搜索、计算器、天气查询等
- **参数验证**：自动参数类型和格式验证
- **结果缓存**：智能缓存机制提升性能
- **错误处理**：完善的异常处理和恢复机制

### 📚 RAG向量检索系统
- **语义搜索**：基于向量相似度的智能检索
- **知识管理**：文档的增删改查和版本控制
- **多模态支持**：文本、代码等多种内容类型
- **实时索引**：文档变更时自动更新索引
- **相似度排序**：按相关性智能排序检索结果

### 🔄 多模型统一架构
- **注册表模式**：统一的模型管理接口
- **动态切换**：运行时无缝切换不同模型
- **负载均衡**：多模型实例的智能调度
- **成本优化**：根据任务特点选择最优模型
- **监控统计**：模型使用情况和性能监控

### 📡 SSE实时状态推送
- **事件驱动**：基于事件的实时状态更新
- **多客户端**：支持同时连接多个客户端
- **状态可视化**：执行过程的详细状态展示
- **连接管理**：自动处理客户端连接和断开
- **消息广播**：统一的消息分发机制

## 🏗️ 系统架构

```
AI Agent
├── Core Layer (核心层)
│   ├── Agent (Think-Execute自主决策引擎)
│   ├── Execution Plan (智能执行计划)
│   └── Error Recovery (智能错误恢复)
├── Model Layer (模型层)
│   ├── Model Interface (统一模型接口)
│   ├── Model Registry (模型注册表)
│   └── Model Implementations (模型实现)
├── Tool Layer (工具层)
│   ├── Tool Interface (工具接口)
│   ├── Tool Registry (工具注册表)
│   └── Tool Implementations (工具实现)
├── RAG Layer (检索层)
│   ├── Vector Engine (向量引擎)
│   ├── Document Management (文档管理)
│   └── Embedding Models (嵌入模型)
├── SSE Layer (推送层)
│   ├── Event Broker (事件代理)
│   └── Client Management (客户端管理)
└── HTTP Layer (服务层)
    ├── API Endpoints (RESTful API)
    ├── Middleware (中间件)
    └── Authentication (认证授权)
```

### 📊 架构特点

- **模块化设计**：各组件独立开发和测试
- **松耦合**：通过接口降低组件间依赖
- **可扩展性**：支持插件式功能扩展
- **高可用性**：内置错误恢复和监控机制
- **性能优化**：缓存、连接池等优化策略

## 🚀 快速开始

### 📋 环境要求

- Go 1.19+
- PostgreSQL 12+ (启用pgvector扩展)
- 至少4GB内存
- 网络连接（用于外部API调用）

### 🛠️ 安装步骤

#### 1. 克隆项目

```bash
git clone https://github.com/xuedingjie/ai-agent-go.git
cd ai-agent-go
```

#### 2. 安装依赖

```bash
go mod tidy
```

#### 3. 数据库初始化

```bash
# 安装pgvector扩展
psql -U postgres -c "CREATE EXTENSION IF NOT EXISTS vector;"

# 创建数据库
createdb aigent
```

#### 4. 配置环境

创建 `config.json`配置文件：

```json
{
  "server": {
    "port": "8080",
    "host": "localhost"
  },
  "agent": {
    "max_iterations": 10,
    "timeout": 300000000000,
    "debug": false
  },
  "models": [
    {
      "name": "openai-gpt35",
      "type": "gpt-3.5-turbo",
      "api_key": "your-openai-api-key",
      "max_tokens": 2000,
      "temperature": 0.7,
      "timeout": 300,
      "enabled": true
    }
  ],
  "database": {
    "host": "localhost",
    "port": 5432,
    "user": "postgres",
    "password": "your-password",
    "database": "aigent",
    "ssl_mode": "disable"
  },
  "features": {
    "enable_rag": true,
    "enable_tools": true,
    "enable_sse": true,
    "enable_metrics": false
  }
}
```

### 🚀 启动应用

#### 开发模式启动

```bash
# 基本启动
./aigent

# 指定配置文件
./aigent -config ./config.json

# 启用调试模式
./aigent -debug

# 查看版本信息
./aigent version
```

#### 生产环境部署

```bash
# 编译优化版本
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o aigent .

# 使用systemd管理服务
sudo systemctl start aigent
sudo systemctl enable aigent
```

#### Docker部署

```bash
# 构建镜像
docker build -t aigent .

# 运行容器
docker run -d -p 8080:8080 \
  -e DATABASE_URL=postgresql://user:pass@host:5432/aigent \
  -e OPENAI_API_KEY=your-api-key \
  --name aigent aigent
```

### ✅ 验证安装

```bash
# 检查服务状态
curl http://localhost:8080/health

# 测试基本功能
curl http://localhost:8080/api/v1/tools
```

## 📡 API接口文档

### 🤖 Agent核心接口

#### 执行智能任务
```bash
# 基本任务执行
curl -X POST http://localhost:8080/api/v1/agent/execute \
  -H "Content-Type: application/json" \
  -d '{
    "query": "分析2023年AI发展趋势",
    "model_name": "gpt-3.5-turbo",
    "max_tokens": 2000,
    "temperature": 0.7,
    "timeout": 300
  }'

# 带上下文的复杂任务
curl -X POST http://localhost:8080/api/v1/agent/execute \
  -H "Content-Type: application/json" \
  -d '{
    "query": "基于搜索结果总结Go语言最佳实践",
    "model_name": "gpt-4",
    "max_tokens": 3000,
    "temperature": 0.3
  }'
```

#### 查看Agent状态
```bash
curl http://localhost:8080/api/v1/agent/status
```

### 🛠️ 工具管理接口

#### 获取工具列表
```bash
curl http://localhost:8080/api/v1/tools

# 响应示例
{
  "count": 3,
  "tools": [
    {
      "name": "web_search",
      "description": "执行网络搜索，获取最新网络信息",
      "parameters": {
        "query": {"type": "string", "description": "搜索查询词"},
        "max_results": {"type": "integer", "default": 5}
      }
    }
  ]
}
```

#### 执行工具调用
```bash
curl -X POST http://localhost:8080/api/v1/tools/execute \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "calculator",
    "input": "{\"expression\": \"2**10 + 100\"}"
  }'
```

### 🔄 模型管理接口

#### 获取可用模型
```bash
curl http://localhost:8080/api/v1/models

# 响应示例
{
  "count": 4,
  "models": ["gpt-3.5-turbo", "gpt-4", "qwen-turbo", "llama2"]
}
```

#### 动态注册模型
```bash
curl -X POST http://localhost:8080/api/v1/models \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-openai-model",
    "model_id": "gpt-4-turbo",
    "api_key": "your-api-key",
    "max_tokens": 2000,
    "temperature": 0.7
  }'
```

### 📚 RAG检索接口

#### 添加文档
```bash
curl -X POST http://localhost:8080/api/v1/rag/documents \
  -H "Content-Type: application/json" \
  -d '{
    "id": "doc_001",
    "content": "Go语言是一门开源的编程语言...",
    "metadata": {"category": "programming", "language": "Go"}
  }'
```

#### 执行语义搜索
```bash
curl "http://localhost:8080/api/v1/rag/search?query=Go语言最佳实践&top_k=5"
```

### 📡 SSE实时事件

#### 连接事件流
```bash
curl -H "Accept: text/event-stream" http://localhost:8080/api/v1/events

# 事件流示例
id: connect
event: connected
data: {"clientId": "client_123", "timestamp": 1700000000, "message": "已成功连接到SSE服务器"}

id: think_1
event: agent
data: {"status": "thinking", "message": "第1轮思考中...", "timestamp": 1700000001}

id: step_1_start
event: agent
data: {"status": "executing", "message": "执行步骤1: search_tool", "timestamp": 1700000002}
```

### 🏥 健康检查接口

```bash
# 服务健康检查
curl http://localhost:8080/health

# 服务就绪检查
curl http://localhost:8080/ready
```

## ⚙️ 配置说明

### 📁 配置文件结构

```json
{
  "server": {
    "port": "8080",
    "host": "localhost",
    "read_timeout": 30,
    "write_timeout": 30,
    "idle_timeout": 120
  },
  "agent": {
    "max_iterations": 10,
    "timeout": 300000000000,
    "debug": false
  },
  "models": [
    {
      "name": "openai-gpt35",
      "type": "gpt-3.5-turbo",
      "api_key": "your-openai-api-key",
      "api_endpoint": "https://api.openai.com/v1",
      "max_tokens": 2000,
      "temperature": 0.7,
      "timeout": 300,
      "enabled": true
    }
  ],
  "database": {
    "url": "postgresql://user:pass@localhost:5432/aigent",
    "host": "localhost",
    "port": 5432,
    "user": "postgres",
    "password": "your-password",
    "database": "aigent",
    "ssl_mode": "disable"
  },
  "logging": {
    "level": "info",
    "format": "json",
    "output": "stdout",
    "max_size": 100,
    "max_age": 30,
    "max_backups": 3,
    "compress": true
  },
  "features": {
    "enable_rag": true,
    "enable_tools": true,
    "enable_sse": true,
    "enable_metrics": false
  }
}
```

### 🌐 环境变量配置

```bash
# 🏢 服务器配置
SERVER_PORT=8080
SERVER_HOST=localhost
SERVER_READ_TIMEOUT=30
SERVER_WRITE_TIMEOUT=30

# 🤖 Agent配置
AGENT_MAX_ITERATIONS=10
AGENT_TIMEOUT=300
AGENT_DEBUG=true

# 🗄️ 数据库配置
DATABASE_URL=postgresql://user:pass@localhost:5432/aigent
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=postgres
DATABASE_PASSWORD=your-password
DATABASE_NAME=aigent
DATABASE_SSL_MODE=disable

# 🧠 模型配置
OPENAI_API_KEY=your-openai-api-key
QWEN_API_KEY=your-qwen-api-key
LLAMA_API_ENDPOINT=http://localhost:8000

# ⚙️ 功能开关
ENABLE_RAG=true
ENABLE_TOOLS=true
ENABLE_SSE=true
ENABLE_METRICS=false

# 📝 日志配置
LOG_LEVEL=info
LOG_FORMAT=json
LOG_OUTPUT=stdout
```

### 🤖 支持的模型类型

#### OpenAI系列
- **GPT-3.5 Turbo**: `gpt-3.5-turbo`
- **GPT-4**: `gpt-4`
- **GPT-4 Turbo**: `gpt-4-turbo`

#### 阿里云通义千问
- **Qwen Turbo**: `qwen-turbo`
- **Qwen Plus**: `qwen-plus`
- **Qwen Max**: `qwen-max`

#### 本地大模型
- **LLaMA 2**: `llama2`
- **LLaMA 3**: `llama3`
- **自定义模型**: `custom-model`

### 🔧 配置优先级

配置按以下优先级加载：
1. **环境变量** (最高优先级)
2. **配置文件** 
3. **默认值** (最低优先级)

### 📊 性能调优建议

```json
{
  "agent": {
    "max_iterations": 5,
    "timeout": 180000000000
  },
  "models": [
    {
      "max_tokens": 1500,
      "temperature": 0.3
    }
  ],
  "logging": {
    "level": "warn"
  }
}
```

## 🛠️ 开发指南

### 🧰 开发环境搭建

```bash
# 克隆项目
git clone https://github.com/xuedingjie/ai-agent-go.git
cd ai-agent-go

# 安装开发依赖
go mod tidy

# 运行测试
go test -v ./...

# 代码格式化
go fmt ./...

# 静态检查
go vet ./...
```

### 🔧 自定义工具开发

#### 基础工具实现

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "aigent/internal/tool"
)

type WeatherTool struct {
    APIKey string
}

func (t *WeatherTool) Name() string {
    return "weather"
}

func (t *WeatherTool) Description() string {
    return "查询指定城市的天气信息"
}

func (t *WeatherTool) Parameters() map[string]interface{} {
    return map[string]interface{}{
        "city": map[string]interface{}{
            "type": "string",
            "description": "城市名称",
        },
        "country": map[string]interface{}{
            "type": "string",
            "description": "国家代码（可选）",
            "default": "",
        },
    }
}

func (t *WeatherTool) Execute(ctx context.Context, input string) (string, error) {
    var params struct {
        City    string `json:"city"`
        Country string `json:"country"`
    }
    
    if err := json.Unmarshal([]byte(input), &params); err != nil {
        return "", fmt.Errorf("解析参数失败: %w", err)
    }
    
    // 实现天气查询逻辑
    weatherInfo := fmt.Sprintf("天气信息 - %s: 晴朗, 温度22°C", params.City)
    return weatherInfo, nil
}

// 注册工具到全局管理器
func init() {
    tool.RegisterToolFactory("weather", func() tool.Tool {
        return &WeatherTool{APIKey: "your-weather-api-key"}
    })
}
```

#### 高级工具特性

```go
// 支持异步执行的工具
type AsyncTool struct{}

func (t *AsyncTool) ExecuteAsync(ctx context.Context, input string) (<-chan string, <-chan error) {
    resultChan := make(chan string, 1)
    errorChan := make(chan error, 1)
    
    go func() {
        // 异步执行逻辑
        select {
        case <-ctx.Done():
            errorChan <- ctx.Err()
        case resultChan <- "异步执行结果":
        }
    }()
    
    return resultChan, errorChan
}
```

### 🤖 自定义模型集成

#### 实现模型接口

```go
package main

import (
    "context"
    "fmt"
    "aigent/internal/model"
)

type CustomModel struct {
    config model.ModelConfig
}

func NewCustomModel(config model.ModelConfig) (model.Model, error) {
    return &CustomModel{config: config}, nil
}

func (m *CustomModel) Generate(ctx context.Context, prompt string) (string, error) {
    // 实现具体的模型调用逻辑
    response := fmt.Sprintf("基于提示词 \"%s\" 的智能响应", prompt)
    return response, nil
}

func (m *CustomModel) Name() string {
    return m.config.Name
}

func (m *CustomModel) Config() model.ModelConfig {
    return m.config
}

// 注册模型到全局注册表
func init() {
    model.RegisterModel("custom-model", NewCustomModel)
}
```

#### 模型池化管理

```go
// 支持连接池的模型实现
type PooledModel struct {
    config model.ModelConfig
    pool   chan *ModelConnection
}

func (m *PooledModel) Generate(ctx context.Context, prompt string) (string, error) {
    // 从连接池获取连接
    conn := <-m.pool
    defer func() { m.pool <- conn }() // 归还连接
    
    // 使用连接执行推理
    return conn.Inference(prompt)
}
```

### 📚 RAG系统扩展

#### 自定义嵌入模型

```go
type CustomEmbeddingModel struct {
    modelName string
}

func (m *CustomEmbeddingModel) Embed(ctx context.Context, text string) ([]float32, error) {
    // 实现文本向量化逻辑
    vector := make([]float32, 1536)
    // ... 向量化算法
    return vector, nil
}

func (m *CustomEmbeddingModel) Name() string {
    return m.modelName
}

// 在RAG引擎中使用
func setupRAG() *rag.Engine {
    config := rag.Config{
        DatabaseURL:    "postgresql://localhost:5432/aigent",
        EmbeddingModel: &CustomEmbeddingModel{modelName: "my-embedding"},
        Dimensions:     1536,
    }
    
    engine, err := rag.NewEngine(config)
    if err != nil {
        panic(err)
    }
    return engine
}
```

### 📡 SSE事件扩展

#### 自定义事件类型

```go
// 定义自定义事件
type CustomEvent struct {
    EventType string      `json:"event_type"`
    Data      interface{} `json:"data"`
    Timestamp int64       `json:"timestamp"`
}

// 发送自定义事件
func SendCustomEvent(broker *sse.Broker, eventType string, data interface{}) {
    event := CustomEvent{
        EventType: eventType,
        Data:      data,
        Timestamp: time.Now().Unix(),
    }
    
    broker.Broadcast("custom", event)
}
```

### 🔌 插件系统

#### 动态加载插件

```go
// 插件接口
type Plugin interface {
    Name() string
    Initialize(config map[string]interface{}) error
    Execute(data interface{}) (interface{}, error)
}

// 插件管理器
type PluginManager struct {
    plugins map[string]Plugin
}

func (pm *PluginManager) LoadPlugin(pluginPath string) error {
    // 动态加载插件逻辑
    return nil
}
```

## 📁 项目结构

```
aigent/
├── internal/
│   ├── config/          # 🛠️ 配置管理
│   │   └── config.go    # 配置加载和验证
│   ├── core/            # 🧠 核心Agent逻辑
│   │   ├── agent.go     # Think-Execute主引擎
│   │   └── plan.go      # 执行计划解析
│   ├── model/           # 🤖 模型接口和实现
│   │   ├── registry.go  # 模型注册表
│   │   └── models.go    # 具体模型实现
│   ├── tool/            # 🛠️ 工具框架
│   │   ├── registry.go  # 工具注册表
│   │   └── tools.go     # 内置工具实现
│   ├── rag/             # 📚 RAG向量检索
│   │   └── engine.go    # 向量引擎核心
│   ├── sse/             # 📡 SSE实时推送
│   │   └── broker.go    # 事件代理
│   └── http/            # 🔌 HTTP服务
│       └── server.go    # RESTful API服务
├── main.go              # 🚀 应用入口
├── main_test.go         # 🧪 测试用例
├── config.json          # ⚙️ 配置文件
├── go.mod               # 📦 Go模块文件
├── go.sum               # 🔒 依赖校验
├── Dockerfile           # 🐳 容器配置
├── .gitignore           # 📝 Git忽略文件
└── README.md            # 📖 项目文档
```

## 📊 性能基准测试

### 🚀 基准测试结果

```bash
# 运行基准测试
go test -bench=. -benchmem ./...

# 测试结果示例
BenchmarkAgentExecute-8           1000    1234567 ns/op    45678 B/op    123 allocs/op
BenchmarkToolExecution-8         10000     123456 ns/op     4567 B/op     45 allocs/op
BenchmarkRAGSearch-8              5000     234567 ns/op    12345 B/op     67 allocs/op
```

### 📈 性能优化建议

- **并发处理**：启用Goroutine池优化并发性能
- **连接复用**：数据库和API连接池配置
- **缓存策略**：结果缓存和预热机制
- **内存管理**：及时释放大对象避免内存泄漏

## 🔒 安全考虑

### 🛡️ 安全最佳实践

```go
// API密钥安全存储
func loadSecureConfig() *Config {
    // 从加密配置文件或密钥管理服务加载
    return config
}

// 请求验证和限流
func setupMiddleware() {
    // 添加认证、授权、限流中间件
}
```

### 📋 安全检查清单

- [ ] API密钥加密存储
- [ ] 请求频率限制
- [ ] 输入参数验证
- [ ] 输出内容过滤
- [ ] 数据库连接安全
- [ ] 日志敏感信息脱敏

## 🤝 贡献指南

### 📝 贡献流程

1. **Fork项目**
2. **创建功能分支** (`git checkout -b feature/AmazingFeature`)
3. **提交更改** (`git commit -m "Add some AmazingFeature"`)
4. **推送分支** (`git push origin feature/AmazingFeature`)
5. **开启Pull Request**

### 🎯 代码规范

```bash
# 代码格式化
go fmt ./...

# 静态检查
go vet ./...

# 运行测试
go test -v ./...
```

### 📋 提交信息规范

```
feat: 添加新功能
fix: 修复bug
docs: 更新文档
style: 代码格式调整
refactor: 代码重构
perf: 性能优化
test: 添加测试
chore: 构建过程或辅助工具的变动
```

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

```
MIT License

Copyright (c) 2024 AI Agent Project

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:
```

## 🙏 致谢

- 感谢所有贡献者和支持者
- 特别感谢开源社区的宝贵建议
- 感谢各大AI平台提供的API支持

## 📞 联系方式

- **项目主页**: [https://github.com/xuedingjie/ai-agent-go](https://github.com/xuedingjie/ai-agent-go)
- **问题反馈**: [Issues](https://github.com/xuedingjie/ai-agent-go/issues)
- **讨论交流**: [Discussions](https://github.com/xuedingjie/ai-agent-go/discussions)
- **邮箱**: your-email@example.com

---

<p align="center">
  Made with ❤️ by the xuedingjie
</p>
