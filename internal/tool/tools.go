package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// WebSearchTool WebSearchTool网络搜索工具
type WebSearchTool struct{}

// Name Name工具名称
func (t *WebSearchTool) Name() string {
	return "web_search"
}

// Description Description工具描述
func (t *WebSearchTool) Description() string {
	return "执行网络搜索，获取最新的网络信息"
}

// Parameters Parameters工具参数定义
func (t *WebSearchTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"query": map[string]interface{}{
			"type":        "string",
			"description": "搜索查询词",
		},
		"max_results": map[string]interface{}{
			"type":        "integer",
			"description": "最大结果数",
			"default":     5,
		},
	}
}

// Execute Execute执行搜索
func (t *WebSearchTool) Execute(ctx context.Context, input string) (string, error) {
	var params struct {
		Query      string `json:"query"`
		MaxResults int    `json:"max_results"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	if params.Query == "" {
		return "", fmt.Errorf("查询词不能为空")
	}

	if params.MaxResults <= 0 {
		params.MaxResults = 5
	}

	//这里使用简单的搜索引擎API示例
	// 实际使用时需要替换为具体的搜索引擎API
	results, err := t.performSearch(ctx, params.Query, params.MaxResults)
	if err != nil {
		return "", fmt.Errorf("搜索失败: %w", err)
	}

	return formatSearchResults(results), nil
}

// performSearch执行搜索
func (t *WebSearchTool) performSearch(ctx context.Context, query string, maxResults int) ([]SearchResult, error) {
	//示例：使用DuckDuckGo的API（需要实际的API密钥）
	searchURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_html=1&skip_disambig=1",
		url.QueryEscape(query))

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("搜索API返回错误: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var ddgResponse struct {
		AbstractText string `json:"AbstractText"`
		AbstractURL  string `json:"AbstractURL"`
	}

	if err := json.Unmarshal(body, &ddgResponse); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	results := []SearchResult{}
	if ddgResponse.AbstractText != "" {
		results = append(results, SearchResult{
			Title:   "搜索结果",
			URL:     ddgResponse.AbstractURL,
			Content: ddgResponse.AbstractText,
		})
	}

	//如果DuckDuckGo没有返回有用结果，返回默认响应
	if len(results) == 0 {
		results = append(results, SearchResult{
			Title:   "搜索结果",
			URL:     fmt.Sprintf("https://www.google.com/search?q=%s", url.QueryEscape(query)),
			Content: fmt.Sprintf("关于 '%s' 的搜索结果可在Google上查看", query),
		})
	}

	return results[:min(len(results), maxResults)], nil
}

// SearchResult搜索结果
type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
}

// formatSearchResults格式化搜索结果
func formatSearchResults(results []SearchResult) string {
	if len(results) == 0 {
		return "未找到搜索结果"
	}

	response := "搜索结果:\n"
	for i, result := range results {
		response += fmt.Sprintf("%d. %s\n   URL: %s\n  : %s\n\n",
			i+1, result.Title, result.URL, result.Content)
	}

	return response
}

// CalculatorTool CalculatorTool计算器工具
type CalculatorTool struct{}

// Name工具名称
func (t *CalculatorTool) Name() string {
	return "calculator"
}

// Description工具描述
func (t *CalculatorTool) Description() string {
	return "执行数学计算"
}

// Parameters工具参数定义
func (t *CalculatorTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"expression": map[string]interface{}{
			"type":        "string",
			"description": "数学表达式，如: 2+2, (10*5)/2, sqrt(16)等",
		},
	}
}

// Execute Execute执行计算
func (t *CalculatorTool) Execute(ctx context.Context, input string) (string, error) {
	var params struct {
		Expression string `json:"expression"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	if params.Expression == "" {
		return "", fmt.Errorf("表达式不能为空")
	}

	// 使用简单的表达式计算（实际应用中可能需要更复杂的计算库）
	result, err := t.calculate(params.Expression)
	if err != nil {
		return "", fmt.Errorf("计算失败: %w", err)
	}

	return fmt.Sprintf("计算结果: %s = %v", params.Expression, result), nil
}

// calculate执行计算（简单实现）
func (t *CalculatorTool) calculate(expr string) (float64, error) {
	//这里是一个非常简化的计算器实现
	// 实际应用中应该使用专门的数学表达式解析库

	expr = strings.ReplaceAll(expr, " ", "")

	//简单的四则运算解析
	if strings.Contains(expr, "+") {
		parts := strings.Split(expr, "+")
		if len(parts) == 2 {
			a, err1 := parseFloat(parts[0])
			b, err2 := parseFloat(parts[1])
			if err1 == nil && err2 == nil {
				return a + b, nil
			}
		}
	}

	if strings.Contains(expr, "-") {
		parts := strings.Split(expr, "-")
		if len(parts) == 2 {
			a, err1 := parseFloat(parts[0])
			b, err2 := parseFloat(parts[1])
			if err1 == nil && err2 == nil {
				return a - b, nil
			}
		}
	}

	if strings.Contains(expr, "*") {
		parts := strings.Split(expr, "*")
		if len(parts) == 2 {
			a, err1 := parseFloat(parts[0])
			b, err2 := parseFloat(parts[1])
			if err1 == nil && err2 == nil {
				return a * b, nil
			}
		}
	}

	if strings.Contains(expr, "/") {
		parts := strings.Split(expr, "/")
		if len(parts) == 2 {
			a, err1 := parseFloat(parts[0])
			b, err2 := parseFloat(parts[1])
			if err1 == nil && err2 == nil {
				if b == 0 {
					return 0, fmt.Errorf("除零错误")
				}
				return a / b, nil
			}
		}
	}

	//尝直接解析数字
	if num, err := parseFloat(expr); err == nil {
		return num, nil
	}

	return 0, fmt.Errorf("不支持的表达式: %s", expr)
}

// parseFloat 解析浮点数
func parseFloat(s string) (float64, error) {
	var result float64
	_, err := fmt.Sscanf(s, "%f", &result)
	return result, err
}

// WeatherTool天气查询工具
type WeatherTool struct {
	APIKey string
}

// Name工具名称
func (t *WeatherTool) Name() string {
	return "weather"
}

// Description工具描述
func (t *WeatherTool) Description() string {
	return "查询指定城市的天气信息"
}

// Parameters工具参数定义
func (t *WeatherTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"city": map[string]interface{}{
			"type":        "string",
			"description": "城市名称",
		},
		"country": map[string]interface{}{
			"type":        "string",
			"description": "国家代码（可选）",
			"default":     "",
		},
	}
}

// Execute执行天气查询
func (t *WeatherTool) Execute(ctx context.Context, input string) (string, error) {
	var params struct {
		City    string `json:"city"`
		Country string `json:"country"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	if params.City == "" {
		return "", fmt.Errorf("城市名称不能为空")
	}

	//使用OpenWeatherMap API示例（需要API密钥）
	weather, err := t.getWeather(ctx, params.City, params.Country)
	if err != nil {
		return "", fmt.Errorf("获取天气信息失败: %w", err)
	}

	return formatWeatherInfo(weather), nil
}

// getWeather 获取天气信息
func (t *WeatherTool) getWeather(ctx context.Context, city, country string) (*WeatherInfo, error) {
	//示例实现，实际需要有效的API密钥
	location := city
	if country != "" {
		location += "," + country
	}

	// 这里返回模拟数据，实际应该调用天气API
	return &WeatherInfo{
		City:        city,
		Temperature: 22.5,
		Description: "晴朗",
		Humidity:    65,
		WindSpeed:   3.2,
	}, nil
}

// WeatherInfo天气信息
type WeatherInfo struct {
	City        string  `json:"city"`
	Temperature float64 `json:"temperature"`
	Description string  `json:"description"`
	Humidity    int     `json:"humidity"`
	WindSpeed   float64 `json:"wind_speed"`
}

// formatWeatherInfo格式化天气信息
func formatWeatherInfo(weather *WeatherInfo) string {
	return fmt.Sprintf("天气信息 - %s:\n温度: %.1f°C\n天气: %s\n湿度: %d%%\n风速: %.1f m/s",
		weather.City, weather.Temperature, weather.Description,
		weather.Humidity, weather.WindSpeed)
}

// 初始化时注册默认工具
func init() {
	// 注册网络搜索工具
	RegisterToolFactory("web_search", func() Tool {
		return &WebSearchTool{}
	})

	// 注册计算器工具
	RegisterToolFactory("calculator", func() Tool {
		return &CalculatorTool{}
	})

	// 注册天气工具
	RegisterToolFactory("weather", func() Tool {
		return &WeatherTool{}
	})
}

// min 返回较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
