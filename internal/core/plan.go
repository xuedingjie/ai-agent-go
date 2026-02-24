package core

import (
	"encoding/json"
	"fmt"
)

// ExecutionPlan执行计划
type ExecutionPlan struct {
	Thought string      `json:"thought"`
	Steps   []*PlanStep `json:"steps"`
}

// PlanStep计划步骤
type PlanStep struct {
	Action        string                 `json:"action"`
	Parameters    map[string]interface{} `json:"parameters"`
	ShouldContinue bool                  `json:"should_continue"`
}

// ParseExecutionPlan解析执行计划JSON
func ParseExecutionPlan(response string) (*ExecutionPlan, error) {
	var plan ExecutionPlan
	
	//尝直接解析JSON
	if err := json.Unmarshal([]byte(response), &plan); err != nil {
		// 如果直接解析失败，尝试从代码块中提取JSON
		jsonStr := extractJSONFromResponse(response)
		if jsonStr == "" {
			return nil, fmt.Errorf("无法从响应中提取JSON: %w", err)
		}
		
		if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
			return nil, fmt.Errorf("解析执行计划JSON失败: %w", err)
		}
	}
	
	//验证计划的有效性
	if err := validatePlan(&plan); err != nil {
		return nil, fmt.Errorf("执行计划验证失败: %w", err)
	}
	
	return &plan, nil
}

// extractJSONFromResponse从响应中提取JSON内容
func extractJSONFromResponse(response string) string {
	// 查找代码块中的JSON
	start := -1
	end := -1
	
	// 查找 ```json 开始标记
	jsonStart := "```json"
	if idx := findSubstring(response, jsonStart); idx != -1 {
		start = idx + len(jsonStart)
	} else {
		// 查找 ``` 开始标记
		codeStart := "```"
		if idx := findSubstring(response, codeStart); idx != -1 {
			start = idx + len(codeStart)
		}
	}
	
	// 查找结束标记
	if start != -1 {
		endMarkers := []string{"```", "\n\n"}
		for _, marker := range endMarkers {
			if idx := findSubstring(response[start:], marker); idx != -1 {
				end = start + idx
				break
			}
		}
	}
	
	// 如果没有找到结束标记，使用到字符串末尾
	if start != -1 && end == -1 {
		end = len(response)
	}
	
	if start != -1 && end != -1 && start < end {
		return response[start:end]
	}
	
	return ""
}

// findSubstring查找子字符串，忽略前后空格
func findSubstring(text, substr string) int {
	//简单的字符串查找实现
	for i := 0; i <= len(text)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if text[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// validatePlan验证执行计划的有效性
func validatePlan(plan *ExecutionPlan) error {
	if plan.Thought == "" {
		return fmt.Errorf("执行计划缺少思考过程")
	}
	
	if len(plan.Steps) == 0 {
		return fmt.Errorf("执行计划缺少步骤")
	}
	
	for i, step := range plan.Steps {
		if step.Action == "" {
			return fmt.Errorf("步骤 %d缺少执行动作", i+1)
		}
		
		if step.Parameters == nil {
			return fmt.Errorf("步骤 %d缺少参数", i+1)
		}
		
		//验证特定动作的必需参数
		switch step.Action {
		case "search_tool":
			if _, ok := step.Parameters["tool_name"]; !ok {
				return fmt.Errorf("工具调用步骤缺少tool_name参数")
			}
			if _, ok := step.Parameters["input"]; !ok {
				return fmt.Errorf("工具调用步骤缺少input参数")
			}
		case "rag_search":
			if _, ok := step.Parameters["query"]; !ok {
				return fmt.Errorf("RAG检索步骤缺少query参数")
			}
		case "reason":
			if _, ok := step.Parameters["prompt"]; !ok {
				return fmt.Errorf("推理步骤缺少prompt参数")
			}
		default:
			return fmt.Errorf("未知的执行动作: %s", step.Action)
		}
	}
	
	return nil
}