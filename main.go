// Package main AI Agent主应用
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"aigent/internal/config"
	"aigent/internal/core"
	"aigent/internal/http"
	"aigent/internal/model"
	"aigent/internal/rag"
	"aigent/internal/sse"
	"aigent/internal/tool"

	"github.com/sirupsen/logrus"
)

var (
	configFile = flag.String("config", "config.json", "配置文件路径")
	debug      = flag.Bool("debug", false, "启用调试模式")
	version    = "1.0.0"
)

func main() {
	flag.Parse()

	//显示版本信息
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("AI Agent v%s\n", version)
		return
	}

	// 初始化应用
	app, err := NewApp(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化应用失败: %v\n", err)
		os.Exit(1)
	}

	//启动应用
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "应用运行失败: %v\n", err)
		os.Exit(1)
	}
}

// App应用主结构
type App struct {
	config    *config.Config
	agent     *core.Agent
	sseBroker *sse.Broker
	server    *http.Server
	logger    *logrus.Logger
}

// NewApp 创建新的应用实例
func NewApp(configPath string) (*App, error) {
	// 加载配置
	cfg, err := config.MergeConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}

	// 设置日志
	logger := setupLogger(cfg)

	logger.Info("正在初始化AI Agent应用...")

	// 初始化组件
	sseBroker := sse.NewBroker()

	// 创建应用实例
	app := &App{
		config:    cfg,
		sseBroker: sseBroker,
		logger:    logger,
	}

	// 初始化模型
	if err := app.initModels(); err != nil {
		return nil, fmt.Errorf("初始化模型失败: %w", err)
	}

	// 初始化工具
	if err := app.initTools(); err != nil {
		return nil, fmt.Errorf("初始化工具失败: %w", err)
	}

	// 初始化RAG（如果启用）
	if err := app.initRAG(); err != nil {
		logger.WithError(err).Warn("初始化RAG失败，将禁用RAG功能")
		cfg.Features.EnableRAG = false
	}

	// 初始化Agent
	if err := app.initAgent(); err != nil {
		return nil, fmt.Errorf("初始化Agent失败: %w", err)
	}

	// 初始化HTTP服务器
	if err := app.initServer(); err != nil {
		return nil, fmt.Errorf("初始化HTTP服务器失败: %w", err)
	}

	logger.Info("应用初始化完成")
	return app, nil
}

// initModels 初始化模型
func (a *App) initModels() error {
	modelConfigs := a.config.GetModelConfigs()

	if len(modelConfigs) == 0 {
		a.logger.Warn("未配置任何模型，将使用默认模型")
		// 创建默认的模拟模型配置
		defaultConfig := model.ModelConfig{
			Name:        "default-mock",
			ModelID:     "mock-model",
			MaxTokens:   2000,
			Temperature: 0.7,
			Timeout:     300,
		}
		modelConfigs = append(modelConfigs, defaultConfig)
	}

	// 初始化所有配置的模型
	for _, modelConfig := range modelConfigs {
		a.logger.Infof("初始化模型: %s (%s)", modelConfig.Name, modelConfig.ModelID)

		//这里可以创建模型实例并测试连接
		// 实际应用中可能需要根据模型类型进行不同的初始化
	}

	return nil
}

// initTools 初始化工具
func (a *App) initTools() error {
	if !a.config.Features.EnableTools {
		a.logger.Info("工具功能已禁用")
		return nil
	}

	a.logger.Info("初始化工具框架...")

	// 注册默认工具
	tools := []tool.Tool{
		&tool.WebSearchTool{},
		&tool.CalculatorTool{},
		&tool.WeatherTool{},
	}

	for _, t := range tools {
		if err := tool.RegisterTool(t); err != nil {
			a.logger.WithError(err).Warnf("注册工具失败: %s", t.Name())
		} else {
			a.logger.Infof("已注册工具: %s", t.Name())
		}
	}

	return nil
}

// initRAG 初始化RAG
func (a *App) initRAG() error {
	if !a.config.Features.EnableRAG {
		a.logger.Info("RAG功能已禁用")
		return nil
	}

	a.logger.Info("初始化RAG引擎...")

	// 创建嵌入模型（这里使用模拟模型）
	embeddingModel := rag.NewMockEmbeddingModel()

	// 创建RAG配置
	ragConfig := rag.Config{
		DatabaseURL:    a.config.GetDatabaseURL(),
		EmbeddingModel: embeddingModel,
		Dimensions:     1536,
		TableName:      "documents",
	}

	// 创建RAG引擎
	_, err := rag.NewEngine(ragConfig)
	if err != nil {
		return fmt.Errorf("创建RAG引擎失败: %w", err)
	}

	a.logger.Info("RAG引擎初始化完成")
	return nil
}

// initAgent 初始化Agent
func (a *App) initAgent() error {
	a.logger.Info("初始化Agent...")

	// 创建Agent配置
	agentConfig := a.config.ToCoreAgentConfig()

	// 创建Agent实例
	a.agent = core.NewAgent(agentConfig).
		WithToolManager(tool.GlobalManager).
		WithSSE(a.sseBroker)

	// 注意：RAG引擎需要在initRAG中创建后传递给Agent
	//这里暂时不设置RAG引擎

	a.logger.Info("Agent初始化完成")
	return nil
}

// initServer 初始化HTTP服务器
func (a *App) initServer() error {
	a.logger.Info("初始化HTTP服务器...")

	// 创建服务器配置
	serverConfig := a.config.ToHTTPServerConfig()
	serverConfig.Agent = a.agent
	serverConfig.SSEBroker = a.sseBroker

	// 创建HTTP服务器
	a.server = http.NewServer(serverConfig)

	a.logger.Infof("HTTP服务器配置完成，监听端口: %s", a.config.Server.Port)
	return nil
}

// Run运行应用
func (a *App) Run() error {
	// 创建上下文用于优雅关闭
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//启HTTP服务器（在goroutine中）
	serverErr := make(chan error, 1)
	go func() {
		a.logger.Infof("启动HTTP服务器在端口 %s", a.config.Server.Port)
		serverErr <- a.server.StartWithContext(ctx)
	}()

	//等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	a.logger.Info("AI Agent服务已启动，按 Ctrl+C停服务")

	//等待信号或错误
	select {
	case <-sigChan:
		a.logger.Info("收到停止信号，正在关闭服务...")
	case err := <-serverErr:
		a.logger.WithError(err).Error("服务器错误")
		return err
	}

	// 优雅关闭
	a.shutdown(ctx)

	return nil
}

// shutdown 优雅关闭
func (a *App) shutdown(ctx context.Context) {
	a.logger.Info("开始优雅关闭...")

	// 设置关闭超时
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 关闭SSE代理
	if a.sseBroker != nil {
		a.sseBroker.Close()
		a.logger.Info("SSE代理已关闭")
	}

	// 关闭RAG引擎（如果存在）
	// ragEngine.Close()

	//等待所有连接关闭
	<-shutdownCtx.Done()

	a.logger.Info("服务已完全关闭")
}

// setupLogger 设置日志
func setupLogger(cfg *config.Config) *logrus.Logger {
	logger := logrus.New()

	// 设置日志级别
	level, err := logrus.ParseLevel(cfg.Logging.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// 设置日志格式
	if cfg.Logging.Format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	// 设置输出
	switch cfg.Logging.Output {
	case "stdout":
		logger.SetOutput(os.Stdout)
	case "stderr":
		logger.SetOutput(os.Stderr)
	default:
		//尝打开文件
		if file, err := os.OpenFile(cfg.Logging.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
			logger.SetOutput(file)
		} else {
			logger.SetOutput(os.Stdout)
			logger.WithError(err).Warn("无法打开日志文件，使用标准输出")
		}
	}

	return logger
}
