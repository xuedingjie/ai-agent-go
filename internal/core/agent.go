// Package core 实现AI Agent的核心架构
package core

import (
	"context"
	"fmt"
	"strings"
	"time"

	"aigent/internal/model"
	"aigent/internal/tool"
	"aigent/internal/rag"
	"aigent/internal/sse"
	
	"github.com/sirupsen/logrus"
)

// AgentStatus表示Agent的执行状态
type AgentStatus string

const (
	StatusThinking   AgentStatus = "thinking"
	StatusPlanning   AgentStatus = "planning"
	StatusExecuting  AgentStatus = "executing"
	StatusCompleted  AgentStatus = "completed"
	StatusError      AgentStatus = "error"
)

// AgentEvent表示Agent执行过程中的事件
type AgentEvent struct {
	ID        string      `json:"id"`
	Timestamp time.Time   `json:"timestamp"`
	Status    AgentStatus `json:"status"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
}

// AgentConfig Agent配置
type AgentConfig struct {
	ModelName     string        `json:"model_name"`
	MaxIterations int           `json:"max_iterations"`
	Timeout       time.Duration `json:"timeout"`
	Debug         bool          `json:"debug"`
}

// Agent AI Agent核心实现
type Agent struct {
	config      AgentConfig
	model       model.Model
	toolManager *tool.Manager
	ragEngine   *rag.Engine
	sseBroker   *sse.Broker
	logger      *logrus.Logger
}

// NewAgent 创建新的Agent实例
func NewAgent(config AgentConfig) *Agent {
	logger := logrus.New()
	if config.Debug {
		logger.SetLevel(logrus.DebugLevel)
	}
	
	return &Agent{
		config: config,
		logger: logger,
	}
}

// WithModel 设置模型
func (a *Agent) WithModel(m model.Model) *Agent {
	a.model = m
	return a
}

// WithToolManager 设置工具管理器
func (a *Agent) WithToolManager(tm *tool.Manager) *Agent {
	a.toolManager = tm
	return a
}

// WithRAG 设置RAG引擎
func (a *Agent) WithRAG(engine *rag.Engine) *Agent {
	a.ragEngine = engine
	return a
}

// WithSSE 设置SSE推送
func (a *Agent) WithSSE(broker *sse.Broker) *Agent {
	a.sseBroker = broker
	return a
}

// Execute执行Think-Execute循环
func (a *Agent) Execute(ctx context.Context, query string) (string, error) {
	if a.model == nil {
		return "", fmt.Errorf("model not configured")
	}

	ctx, cancel := context.WithTimeout(ctx, a.config.Timeout)
	defer cancel()

	// 发送开始事件
	a.sendEvent("start", StatusThinking, "开始处理请求", nil)

	result, err := a.thinkExecuteLoop(ctx, query)
	if err != nil {
		a.sendEvent("error", StatusError, fmt.Sprintf("执行出错: %v", err), nil)
		return "", err
	}

	a.sendEvent("complete", StatusCompleted, "任务完成", map[string]interface{}{
		"result": result,
	})
	
	return result, nil
}

// thinkWithRetryInternal 带重试的思考函数
func (a *Agent) thinkWithRetryInternal(ctx context.Context, query string, retryCount int) (*ExecutionPlan, error) {
	// 构建重试提示词，包含错误信息
	prompt := a.buildRetryThinkPrompt(query, retryCount)
	
	a.logger.Debugf("重试思考提示词: %s", prompt)
	
	response, err := a.model.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("重试思考时模型生成失败: %w", err)
	}

	a.logger.Debugf("重试模型响应: %s", response)

	// 解析执行计划
	plan, err := ParseExecutionPlan(response)
	if err != nil {
		return nil, fmt.Errorf("重试解析执行计划失败: %w", err)
	}

	return plan, nil
}

// thinkExecuteLoop Think-Execute主循环
func (a *Agent) thinkExecuteLoop(ctx context.Context, query string) (string, error) {
	iteration := 0
	currentQuery := query
	
	for iteration < a.config.MaxIterations {
		iteration++
		a.logger.Debugf("执行第 %d-执行循环", iteration)
		
		// 1.思阶段 - 分析问题并制定计划
		a.sendEvent(fmt.Sprintf("think_%d", iteration), StatusThinking, 
			fmt.Sprintf("第 %d中...", iteration), nil)
		
		plan, err := a.think(ctx, currentQuery, iteration)
		if err != nil {
			return "", fmt.Errorf("思考阶段出错: %w", err)
		}
		
		a.sendEvent(fmt.Sprintf("plan_%d", iteration), StatusPlanning, 
			"制定执行计划", plan)

		// 2.执行阶段 -执行计划
		a.sendEvent(fmt.Sprintf("execute_%d", iteration), StatusExecuting, 
			"执行计划中...", plan)
		
		result, shouldContinue, err := a.execute(ctx, plan)
		if err != nil {
			return "", fmt.Errorf("执行阶段出错: %w", err)
		}

		if !shouldContinue {
			return result, nil
		}
		
		// 更新查询为执行结果，继续下一轮
		currentQuery = result
	}

	return "", fmt.Errorf("超过最大迭代次数 %d", a.config.MaxIterations)
}

// think思阶段 - 分析问题并制定执行计划
func (a *Agent) think(ctx context.Context, query string, iteration int) (*ExecutionPlan, error) {
	//构建思考提示词
	prompt := a.buildThinkPrompt(query, iteration)
	
	a.logger.Debugf("思考提示词: %s", prompt)
	
	response, err := a.model.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("模型生成失败: %w", err)
	}

	a.logger.Debugf("模型响应: %s", response)

	// 解析执行计划
	plan, err := ParseExecutionPlan(response)
	if err != nil {
		// 如果解析失败，尝试重新思考
		if iteration < 3 { // 最多重试3次
			a.logger.Warnf("解析执行计划失败，第%d次重试: %v", iteration, err)
			return a.thinkWithRetryInternal(ctx, query, iteration+1)
		}
		return nil, fmt.Errorf("解析执行计划失败: %w", err)
	}

	// 验证计划的合理性
	if err := a.validatePlan(plan, query); err != nil {
		return nil, fmt.Errorf("执行计划验证失败: %w", err)
	}

	return plan, nil
}

// execute执行阶段 -执行计划中的步骤
func (a *Agent) execute(ctx context.Context, plan *ExecutionPlan) (string, bool, error) {
	result := ""
	shouldContinue := false
	executionHistory := []string{} // 记录执行历史

	for i, step := range plan.Steps {
		a.logger.Debugf("执行步骤 %d: %s", i+1, step.Action)
		
		// 发送步骤执行事件
		a.sendEvent(fmt.Sprintf("step_%d_start", i+1), StatusExecuting, 
			fmt.Sprintf("执行步骤 %d: %s", i+1, step.Action), step)
		
		stepResult, err := a.executeStep(ctx, step)
		if err != nil {
			// 记录错误并尝试恢复
			errorMsg := fmt.Sprintf("执行步骤 %d失败: %v", i+1, err)
			a.logger.Errorf(errorMsg)
			
			// 发送错误事件
			a.sendEvent(fmt.Sprintf("step_%d_error", i+1), StatusError, errorMsg, nil)
			
			// 尝试错误恢复
			if recoveredResult, recoverErr := a.recoverFromError(ctx, step, err, executionHistory); recoverErr == nil {
				stepResult = recoveredResult
				a.sendEvent(fmt.Sprintf("step_%d_recovered", i+1), StatusExecuting, 
					"步骤执行已恢复", stepResult)
			} else {
				return "", false, fmt.Errorf("%s，恢复失败: %w", errorMsg, recoverErr)
			}
		}

		result = stepResult
		executionHistory = append(executionHistory, result)
		shouldContinue = step.ShouldContinue
		
		// 发送步骤完成事件
		a.sendEvent(fmt.Sprintf("step_%d_complete", i+1), StatusExecuting, 
			fmt.Sprintf("步骤 %d完成", i+1), map[string]interface{}{
				"result": stepResult,
				"should_continue": shouldContinue,
			})
		
		// 如果步骤要求继续且有后续步骤，继续执行
		if shouldContinue && i < len(plan.Steps)-1 {
			continue
		}
		
		break
	}

	return result, shouldContinue, nil
}

// executeStep执行单个步骤
func (a *Agent) executeStep(ctx context.Context, step *PlanStep) (string, error) {
	switch step.Action {
	case "search_tool":
		return a.executeToolStep(ctx, step)
	case "rag_search":
		return a.executeRAGStep(ctx, step)
	case "reason":
		return a.executeReasonStep(ctx, step)
	default:
		return "", fmt.Errorf("未知的执行动作: %s", step.Action)
	}
}

// executeToolStep执行工具调用步骤
func (a *Agent) executeToolStep(ctx context.Context, step *PlanStep) (string, error) {
	if a.toolManager == nil {
		return "", fmt.Errorf("工具管理器未配置")
	}

	toolName := step.Parameters["tool_name"].(string)
	toolInput := step.Parameters["input"].(string)
	
	result, err := a.toolManager.ExecuteTool(ctx, toolName, toolInput)
	if err != nil {
		return "", fmt.Errorf("工具调用失败 %s: %w", toolName, err)
	}

	return result, nil
}

// executeRAGStep执行RAG检索步骤
func (a *Agent) executeRAGStep(ctx context.Context, step *PlanStep) (string, error) {
	if a.ragEngine == nil {
		return "", fmt.Errorf("RAG引擎未配置")
	}

	query := step.Parameters["query"].(string)
	topK := 5
	if k, ok := step.Parameters["top_k"].(float64); ok {
		topK = int(k)
	}
	
	results, err := a.ragEngine.Search(ctx, query, topK)
	if err != nil {
		return "", fmt.Errorf("RAG检索失败: %w", err)
	}

	return formatRAGResults(results), nil
}

// executeReasonStep执行推理步骤
func (a *Agent) executeReasonStep(ctx context.Context, step *PlanStep) (string, error) {
	prompt := step.Parameters["prompt"].(string)
	
	response, err := a.model.Generate(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("推理失败: %w", err)
	}

	return response, nil
}

// validatePlan 验证执行计划的合理性
func (a *Agent) validatePlan(plan *ExecutionPlan, query string) error {
	if plan == nil {
		return fmt.Errorf("执行计划不能为空")
	}
	
	if len(plan.Steps) == 0 {
		return fmt.Errorf("执行计划必须包含至少一个步骤")
	}
	
	// 检查步骤的逻辑连贯性
	for i, step := range plan.Steps {
		// 检查工具调用步骤的参数
		if step.Action == "search_tool" {
			if a.toolManager == nil {
				return fmt.Errorf("步骤 %d需要工具调用，但工具管理器未配置", i+1)
			}
			
			toolName, ok := step.Parameters["tool_name"].(string)
			if !ok || toolName == "" {
				return fmt.Errorf("步骤 %d的工具调用缺少tool_name参数", i+1)
			}
			
			// 检查工具是否存在
			tools := a.toolManager.ListTools()
			toolExists := false
			for _, t := range tools {
				if t.Name == toolName {
					toolExists = true
					break
				}
			}
			if !toolExists {
				return fmt.Errorf("步骤 %d指定的工具 %s 不存在", i+1, toolName)
			}
		}
		
		// 检查RAG检索步骤
		if step.Action == "rag_search" {
			if a.ragEngine == nil {
				return fmt.Errorf("步骤 %d需要RAG检索，但RAG引擎未配置", i+1)
			}
			
			queryParam, ok := step.Parameters["query"].(string)
			if !ok || queryParam == "" {
				return fmt.Errorf("步骤 %d的RAG检索缺少query参数", i+1)
			}
		}
		
		// 检查推理步骤
		if step.Action == "reason" {
			if _, ok := step.Parameters["prompt"].(string); !ok {
				return fmt.Errorf("步骤 %d的推理缺少prompt参数", i+1)
			}
		}
	}
	
	// 检查计划的最终目标相关性
	if !a.isPlanRelevant(plan, query) {
		return fmt.Errorf("执行计划与用户查询的相关性不足")
	}
	
	return nil
}

// isPlanRelevant 检查计划与查询的相关性
func (a *Agent) isPlanRelevant(plan *ExecutionPlan, query string) bool {
	// 简单的相关性检查：计划思考内容应该包含查询关键词
	queryLower := strings.ToLower(query)
	thoughtLower := strings.ToLower(plan.Thought)
	
	// 检查查询中的关键词是否在思考内容中出现
	words := strings.Fields(queryLower)
	matchCount := 0
	
	for _, word := range words {
		if len(word) > 2 && strings.Contains(thoughtLower, word) {
			matchCount++
		}
	}
	
	// 如果至少30%的关键词匹配，则认为相关
	return float64(matchCount)/float64(len(words)) >= 0.3
}

// sendEvent 发送SSE事件
func (a *Agent) sendEvent(id string, status AgentStatus, message string, data interface{}) {
	if a.sseBroker != nil {
		event := &AgentEvent{
			ID:        id,
			Timestamp: time.Now(),
			Status:    status,
			Message:   message,
			Data:      data,
		}
		a.sseBroker.Broadcast("agent", event)
	}
	
	if a.logger != nil {
		a.logger.WithFields(logrus.Fields{
			"event_id": id,
			"status":   status,
			"message":  message,
		}).Info("Agent事件")
	}
}

// buildRetryThinkPrompt构建重试思考提示词
func (a *Agent) buildRetryThinkPrompt(query string, retryCount int) string {
	availableTools := []string{}
	if a.toolManager != nil {
		tools := a.toolManager.ListTools()
		for _, t := range tools {
			availableTools = append(availableTools, t.Name)
		}
	}

	template := `你是一个智能AI助手，之前的执行计划解析失败了，请重新分析用户问题并制定正确的执行计划。

用户问题: %s
重试次数: 第%d次

可用工具: %v

请重新分析问题并制定执行计划，使用以下JSON格式:

{
  "thought": "你的思考过程，需要更详细地分析问题",
  "steps": [
    {
      "action": "具体执行动作(search_tool/rag_search/reason)",
      "parameters": {
        "相关参数": "值"
      },
      "should_continue": true/false
    }
  ]
}

注意：
1. 请确保JSON格式正确
2. 思考过程要更详细和具体
3. 步骤要逻辑清晰，避免循环依赖
4. 确保计划能够解决用户的核心问题

请只返回JSON格式的计划，不要其他说明。`

	return fmt.Sprintf(template, query, retryCount, availableTools)
}

// buildThinkPrompt构建思考阶段的提示词
func (a *Agent) buildThinkPrompt(query string, iteration int) string {
	availableTools := []string{}
	if a.toolManager != nil {
		tools := a.toolManager.ListTools()
		for _, t := range tools {
			availableTools = append(availableTools, t.Name)
		}
	}

	template := `你是一个智能AI助手，需要分析用户问题并制定执行计划。

当前轮次: 第 %d用户问题: %s

可用工具: %v

请分析问题并制定执行计划，使用以下JSON格式:

{
  "thought": "你的思考过程",
  "steps": [
    {
      "action": "具体执行动作(search_tool/rag_search/reason)",
      "parameters": {
        "相关参数": "值"
      },
      "should_continue": true/false
    }
  ]
}

执行动作说明:
- search_tool:调用工具，参数包括tool_name, input
- rag_search:向检索，参数包括query, top_k
- reason:推分析，参数包括prompt

请只返回JSON格式的计划，不要其他说明。`

	return fmt.Sprintf(template, iteration, query, availableTools)
}

// recoverFromError 从错误中恢复
func (a *Agent) recoverFromError(ctx context.Context, step *PlanStep, err error, history []string) (string, error) {
	// 根据错误类型进行不同的恢复策略
	errorMsg := err.Error()
	
	// 工具调用错误恢复
	if step.Action == "search_tool" && strings.Contains(errorMsg, "工具调用失败") {
		return a.recoverToolError(ctx, step, history)
	}
	
	// RAG检索错误恢复
	if step.Action == "rag_search" && strings.Contains(errorMsg, "RAG检索失败") {
		return a.recoverRAGError(ctx, step, history)
	}
	
	// 推理错误恢复
	if step.Action == "reason" && strings.Contains(errorMsg, "推理失败") {
		return a.recoverReasonError(ctx, step, history)
	}
	
	// 默认恢复策略：使用历史信息进行推理
	return a.defaultRecovery(ctx, step, history, errorMsg)
}

// recoverToolError 工具调用错误恢复
func (a *Agent) recoverToolError(ctx context.Context, step *PlanStep, history []string) (string, error) {
	if a.toolManager == nil {
		return "", fmt.Errorf("工具管理器未配置，无法恢复")
	}
	
	toolName := step.Parameters["tool_name"].(string)
	
	// 尝试使用不同的参数重新调用
	alternativeInputs := a.generateAlternativeInputs(step, history)
	
	for _, input := range alternativeInputs {
		result, err := a.toolManager.ExecuteTool(ctx, toolName, input)
		if err == nil {
			return result, nil
		}
	}
	
	return "", fmt.Errorf("所有替代方案都失败")
}

// recoverRAGError RAG检索错误恢复
func (a *Agent) recoverRAGError(ctx context.Context, step *PlanStep, history []string) (string, error) {
	if a.ragEngine == nil {
		return "", fmt.Errorf("RAG引擎未配置，无法恢复")
	}
	
	query := step.Parameters["query"].(string)
	
	// 尝试修改查询语句
	alternativeQueries := a.generateAlternativeQueries(query, history)
	
	for _, altQuery := range alternativeQueries {
		results, err := a.ragEngine.Search(ctx, altQuery, 3)
		if err == nil && len(results) > 0 {
			return formatRAGResults(results), nil
		}
	}
	
	return "", fmt.Errorf("所有替代查询都失败")
}

// recoverReasonError 推理错误恢复
func (a *Agent) recoverReasonError(ctx context.Context, step *PlanStep, history []string) (string, error) {
	originalPrompt := step.Parameters["prompt"].(string)
	
	// 基于历史信息生成新的推理提示词
	recoveryPrompt := fmt.Sprintf("之前的推理过程出现了问题，请基于以下历史信息重新思考：\n\n历史执行结果: %v\n\n原始问题: %s\n\n请重新分析并给出合理的回答。", 
		strings.Join(history, "; "), originalPrompt)
	
	response, err := a.model.Generate(ctx, recoveryPrompt)
	if err != nil {
		return "", fmt.Errorf("恢复推理失败: %w", err)
	}
	
	return response, nil
}

// defaultRecovery 默认恢复策略
func (a *Agent) defaultRecovery(ctx context.Context, step *PlanStep, history []string, errorMsg string) (string, error) {
	recoveryPrompt := fmt.Sprintf("执行过程中遇到错误: %s\n\n历史执行情况: %v\n\n请基于现有信息给出一个合理的回答或解决方案。", 
		errorMsg, strings.Join(history, "; "))
	
	response, err := a.model.Generate(ctx, recoveryPrompt)
	if err != nil {
		return "", fmt.Errorf("默认恢复策略失败: %w", err)
	}
	
	return response, nil
}

// generateAlternativeInputs 生成替代输入
func (a *Agent) generateAlternativeInputs(step *PlanStep, history []string) []string {
	inputs := []string{}
	
	if input, ok := step.Parameters["input"].(string); ok {
		// 生成几种不同的输入变体
		inputs = append(inputs, input)
		
		// 基于历史生成替代输入
		if len(history) > 0 {
			_ = history[len(history)-1] // 获取最后结果但不使用，避免编译错误
			inputs = append(inputs, fmt.Sprintf("%s 基于之前的执行结果", input))
			inputs = append(inputs, fmt.Sprintf("%s 结合历史信息", input))
		}
	}
	
	return inputs
}

// generateAlternativeQueries 生成替代查询
func (a *Agent) generateAlternativeQueries(query string, history []string) []string {
	queries := []string{query}
	
	// 生成更通用的查询
	if len(query) > 10 {
		queries = append(queries, query[:len(query)-5]) // 去掉最后几个字符
	}
	
	// 基于历史生成相关查询
	if len(history) > 0 {
		lastResult := history[len(history)-1]
		words := strings.Fields(lastResult)
		if len(words) > 0 {
			queries = append(queries, fmt.Sprintf("%s %s", query, words[0]))
		}
	}
	
	return queries
}

// formatRAGResults格式化RAG检索结果
func formatRAGResults(results []rag.SearchResult) string {
	if len(results) == 0 {
		return "未找到相关结果"
	}
	
	response := "检索到以下相关信息:\n"
	for i, result := range results {
		response += fmt.Sprintf("%d. %s (相似度: %.2f)\n", 
			i+1, result.Document.Content, result.Similarity)
	}
	
	return response
}