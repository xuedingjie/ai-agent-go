// Package http 实现HTTP服务端点
package http

import (
	"context"
	"net/http"
	"time"

	"aigent/internal/core"
	"aigent/internal/model"
	"aigent/internal/sse"
	"aigent/internal/tool"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Server HTTP服务
type Server struct {
	router    *gin.Engine
	agent     *core.Agent
	sseBroker *sse.Broker
	logger    *logrus.Logger
	port      string
}

// Config服务器配置
type Config struct {
	Port      string
	Debug     bool
	Agent     *core.Agent
	SSEBroker *sse.Broker
}

// NewServer创建新的HTTP服务器
func NewServer(config Config) *Server {
	logger := logrus.New()
	if config.Debug {
		logger.SetLevel(logrus.DebugLevel)
	}

	server := &Server{
		agent:     config.Agent,
		sseBroker: config.SSEBroker,
		logger:    logger,
		port:      config.Port,
	}

	server.setupRouter()
	return server
}

// setupRouter设置路由
func (s *Server) setupRouter() {
	// 设置gin为发布模式（生产环境）
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	//中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	
	// API路由组
	api := r.Group("/api/v1")
	{
		// Agent相关接口
		api.POST("/agent/execute", s.handleAgentExecute)
		api.GET("/agent/status", s.handleAgentStatus)

		//模型相关接口
		api.GET("/models", s.handleListModels)
		api.POST("/models", s.handleCreateModel)

		//工具相关接口
		api.GET("/tools", s.handleListTools)
		api.POST("/tools/execute", s.handleExecuteTool)

		// RAG相关接口
		api.POST("/rag/documents", s.handleAddDocument)
		api.GET("/rag/search", s.handleRAGSearch)
		api.GET("/rag/documents", s.handleListDocuments)

		// SSE接口
		api.GET("/events", gin.WrapH(sse.Handler(s.sseBroker)))
	}

	//健检查
	r.GET("/health", s.handleHealthCheck)
	r.GET("/ready", s.handleReadyCheck)

	s.router = r
}

// handleAgentExecute处理Agent执行请求
type AgentExecuteRequest struct {
	Query       string  `json:"query"`
	ModelName   string  `json:"model_name"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
	Timeout     int     `json:"timeout"`
}

func (s *Server) handleAgentExecute(c *gin.Context) {
	var req AgentExecuteRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		s.writeError(c, http.StatusBadRequest, "无效的请求体", err)
		return
	}

	if req.Query == "" {
		s.writeError(c, http.StatusBadRequest, "查询不能为空", nil)
		return
	}

	//设置默认值
	if req.MaxTokens <= 0 {
		req.MaxTokens = 2000
	}
	if req.Temperature <= 0 {
		req.Temperature = 0.7
	}
	if req.Timeout <= 0 {
		req.Timeout = 300
	}

	//创建模型配置
	modelConfig := model.ModelConfig{
		Name:        req.ModelName,
		ModelID:     req.ModelName,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Timeout:     req.Timeout,
	}

	//创建模型实例
	llm, err := model.CreateModel(modelConfig)
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, "创建模型失败", err)
		return
	}

	//配置Agent
	agentConfig := core.AgentConfig{
		ModelName:     req.ModelName,
		MaxIterations: 10,
		Timeout:       time.Duration(req.Timeout) * time.Second,
		Debug:         s.logger.GetLevel() == logrus.DebugLevel,
	}

	// 更新Agent配置
	agent := core.NewAgent(agentConfig).
		WithModel(llm).
		WithToolManager(tool.GlobalManager).
		WithSSE(s.sseBroker)

	if s.agent != nil {
		// 如果已有RAG引擎，复用它
		//这里需要获取现有的RAG引擎引用
	}

	//在后台执行
	go func() {
		ctx := context.Background()
		result, err := agent.Execute(ctx, req.Query)
		if err != nil {
			s.sseBroker.Broadcast("agent_error", map[string]interface{}{
				"error": err.Error(),
				"query": req.Query,
			})
			return
		}

		s.sseBroker.Broadcast("agent_result", map[string]interface{}{
			"result": result,
			"query":  req.Query,
		})
	}()

	c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Agent执行已启动",
		"query":   req.Query,
	})
}

// handleAgentStatus处理Agent状态查询
func (s *Server) handleAgentStatus(c *gin.Context) {
	status := map[string]interface{}{
		"status":        "running",
		"clients_count": s.sseBroker.GetClientsCount(),
		"client_ids":    s.sseBroker.GetClientIDs(),
		"timestamp":     time.Now().Unix(),
	}

	c.JSON(http.StatusOK, status)
}

// handleListModels处理模型列表查询
func (s *Server) handleListModels(c *gin.Context) {
	models := model.GlobalRegistry.ListModels()

	c.JSON(http.StatusOK, map[string]interface{}{
		"models": models,
		"count":  len(models),
	})
}

// CreateModelRequest创建模型请求
type CreateModelRequest struct {
	Name        string  `json:"name"`
	ModelID     string  `json:"model_id"`
	APIKey      string  `json:"api_key"`
	APIEndpoint string  `json:"api_endpoint"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
	Timeout     int     `json:"timeout"`
}

// handleCreateModel处理模型创建
func (s *Server) handleCreateModel(c *gin.Context) {
	var req CreateModelRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		s.writeError(c, http.StatusBadRequest, "无效的请求体", err)
		return
	}

	config := model.ModelConfig{
		Name:        req.Name,
		ModelID:     req.ModelID,
		APIKey:      req.APIKey,
		APIEndpoint: req.APIEndpoint,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Timeout:     req.Timeout,
	}

	if config.Timeout <= 0 {
		config.Timeout = 300
	}

	llm, err := model.CreateModel(config)
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, "创建模型失败", err)
		return
	}

	c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "模型创建成功",
		"model": map[string]interface{}{
			"name": llm.Name(),
			"type": llm.Config().ModelID,
		},
	})
}

// handleListTools处理工具列表查询
func (s *Server) handleListTools(c *gin.Context) {
	tools := tool.GlobalManager.ListTools()

	c.JSON(http.StatusOK, map[string]interface{}{
		"tools": tools,
		"count": len(tools),
	})
}

// ExecuteToolRequest执行工具请求
type ExecuteToolRequest struct {
	ToolName string `json:"tool_name"`
	Input    string `json:"input"`
}

// handleExecuteTool处理工具执行
func (s *Server) handleExecuteTool(c *gin.Context) {
	var req ExecuteToolRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		s.writeError(c, http.StatusBadRequest, "无效的请求体", err)
		return
	}

	if req.ToolName == "" {
		s.writeError(c, http.StatusBadRequest, "工具名称不能为空", nil)
		return
	}

	result, err := tool.GlobalManager.ExecuteTool(c.Request.Context(), req.ToolName, req.Input)
	if err != nil {
		s.writeError(c, http.StatusInternalServerError, "执行工具失败", err)
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"result": result,
		"tool":   req.ToolName,
	})
}

// handleAddDocument处理添加文档
func (s *Server) handleAddDocument(c *gin.Context) {
	//这个接口需要RAG引擎实例
	//这里返回未实现，实际应用中需要注入RAG引擎
	s.writeError(c, http.StatusNotImplemented, "RAG功能需要额外配置", nil)
}

// handleRAGSearch处理RAG检索
func (s *Server) handleRAGSearch(c *gin.Context) {
	//这个接口需要RAG引擎实例
	s.writeError(c, http.StatusNotImplemented, "RAG功能需要额外配置", nil)
}

// handleListDocuments处理文档列表
func (s *Server) handleListDocuments(c *gin.Context) {
	//这个接口需要RAG引擎实例
	s.writeError(c, http.StatusNotImplemented, "RAG功能需要额外配置", nil)
}

// handleHealthCheck处理健康检查
func (s *Server) handleHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
	})
}

// handleReadyCheck处理就绪检查
func (s *Server) handleReadyCheck(c *gin.Context) {
	status := "ready"
	if s.agent == nil {
		status = "not_ready"
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"status":    status,
		"timestamp": time.Now().Unix(),
	})
}

// writeError写入错误响应
func (s *Server) writeError(c *gin.Context, statusCode int, message string, err error) {
	response := map[string]interface{}{
		"error":   message,
		"status":  statusCode,
		"message": message,
	}

	if err != nil && s.logger.GetLevel() == logrus.DebugLevel {
		response["details"] = err.Error()
	}

	c.JSON(statusCode, response)
}

// Start启动服务器
func (s *Server) Start() error {
	s.logger.Infof("HTTP服务器启动在端口 %s", s.port)

	return http.ListenAndServe(":"+s.port, s.router)
}

// StartWithContext启动服务器（支持上下文取消）
func (s *Server) StartWithContext(ctx context.Context) error {
	server := &http.Server{
		Addr:    ":" + s.port,
		Handler: s.router,
	}

	go func() {
		<-ctx.Done()
		s.logger.Info("正在关闭HTTP服务器...")
		server.Shutdown(context.Background())
	}()

	s.logger.Infof("HTTP服务器启动在端口 %s", s.port)
	return server.ListenAndServe()
}

// Router返回路由器（用于测试）
func (s *Server) Router() *gin.Engine {
	return s.router
}
