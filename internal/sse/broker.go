// Package sse 实现SSE实时推送功能
package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Event SSE事件结构
type Event struct {
	ID    string      `json:"id,omitempty"`
	Event string      `json:"event,omitempty"`
	Data  interface{} `json:"data"`
	Retry int         `json:"retry,omitempty"`
}

// Client SSE客户端
type Client struct {
	ID     string
	events chan []byte
	done   chan struct{}
}

// NewClient创建新的客户端
func NewClient(id string) *Client {
	return &Client{
		ID:     id,
		events: make(chan []byte, 100),
		done:   make(chan struct{}),
	}
}

// Send 发送事件给客户端
func (c *Client) Send(event []byte) bool {
	select {
	case c.events <- event:
		return true
	case <-c.done:
		return false
	default:
		//缓区满了，丢弃事件
		return false
	}
}

// Close关闭客户端
func (c *Client) Close() {
	close(c.done)
	close(c.events)
}

// Broker SSE推送代理
type Broker struct {
	clients   map[string]*Client
	clientsMu sync.RWMutex
	events    chan Event
	register  chan *Client
	unregister chan string
	stop      chan struct{}
}

// NewBroker创建新的SSE代理
func NewBroker() *Broker {
	broker := &Broker{
		clients:    make(map[string]*Client),
		events:     make(chan Event, 1000),
		register:   make(chan *Client, 10),
		unregister: make(chan string, 10),
		stop:       make(chan struct{}),
	}
	
	go broker.run()
	return broker
}

// run运行代理主循环
func (b *Broker) run() {
	for {
		select {
		case client := <-b.register:
			b.clientsMu.Lock()
			b.clients[client.ID] = client
			b.clientsMu.Unlock()
			
		case clientID := <-b.unregister:
			b.clientsMu.Lock()
			if client, exists := b.clients[clientID]; exists {
				client.Close()
				delete(b.clients, clientID)
			}
			b.clientsMu.Unlock()
			
		case event := <-b.events:
			//序列化事件
			eventBytes, err := b.serializeEvent(event)
			if err != nil {
				continue
			}
			
			// 发送事件给所有客户端
			b.clientsMu.RLock()
			clients := make([]*Client, 0, len(b.clients))
			for _, client := range b.clients {
				clients = append(clients, client)
			}
			b.clientsMu.RUnlock()
			
			//异步发送
			for _, client := range clients {
				go func(c *Client) {
					c.Send(eventBytes)
				}(client)
			}
			
		case <-b.stop:
			// 清理所有客户端
			b.clientsMu.Lock()
			for _, client := range b.clients {
				client.Close()
			}
			b.clients = make(map[string]*Client)
			b.clientsMu.Unlock()
			return
		}
	}
}

// serializeEvent序化事件为SSE格式
func (b *Broker) serializeEvent(event Event) ([]byte, error) {
	var sseData string
	
	//构建SSE数据格式
	if event.ID != "" {
		sseData += fmt.Sprintf("id: %s\n", event.ID)
	}
	
	if event.Event != "" {
		sseData += fmt.Sprintf("event: %s\n", event.Event)
	}
	
	//序化数据为JSON
	jsonData, err := json.Marshal(event.Data)
	if err != nil {
		return nil, fmt.Errorf("序列化事件数据失败: %w", err)
	}
	
	//按行分割数据并添加data:前缀
	dataStr := string(jsonData)
	lines := splitLines(dataStr)
	for _, line := range lines {
		sseData += fmt.Sprintf("data: %s\n", line)
	}
	
	if event.Retry > 0 {
		sseData += fmt.Sprintf("retry: %d\n", event.Retry)
	}
	
	// 添加空行结束事件
	sseData += "\n"
	
	return []byte(sseData), nil
}

// splitLines按换行符分割字符串
func splitLines(s string) []string {
	lines := []string{}
	start := 0
	
	for i, char := range s {
		if char == '\n' {
			if start < i {
				lines = append(lines, s[start:i])
			}
			start = i + 1
		}
	}
	
	// 添加最后一行
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	
	return lines
}

// Subscribe订阅SSE事件
func (b *Broker) Subscribe(clientID string, w http.ResponseWriter, r *http.Request) {
	//设置SSE响应头
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	//创建客户端
	client := NewClient(clientID)
	
	//注册客户端
	b.register <- client
	defer func() {
		b.unregister <- clientID
	}()
	
	// 设置连接超时
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Minute)
	defer cancel()
	
	//监听客户端事件和请求取消
	go func() {
		<-ctx.Done()
		b.unregister <- clientID
	}()
	
	//发送连接成功事件
	successEvent := Event{
		ID:    "connect",
		Event: "connected",
		Data: map[string]interface{}{
			"clientId": clientID,
			"timestamp": time.Now().Unix(),
			"message": "已成功连接到SSE服务器",
		},
	}
	
	if eventBytes, err := b.serializeEvent(successEvent); err == nil {
		w.Write(eventBytes)
		w.(http.Flusher).Flush()
	}
	
	//主循环：发送事件给客户端
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-client.events:
			_, err := w.Write(event)
			if err != nil {
				return
			}
			w.(http.Flusher).Flush()
		case <-client.done:
			return
		}
	}
}

// Broadcast广事件给所有客户端
func (b *Broker) Broadcast(eventType string, data interface{}) {
	event := Event{
		ID:    generateEventID(),
		Event: eventType,
		Data:  data,
		Retry: 5000, // 5秒重试
	}
	
	//异步发送
	select {
	case b.events <- event:
	default:
		// 事件通道满了，丢弃事件
	}
}

// SendTo 发送事件给特定客户端
func (b *Broker) SendTo(clientID, eventType string, data interface{}) bool {
	b.clientsMu.RLock()
	client, exists := b.clients[clientID]
	b.clientsMu.RUnlock()
	
	if !exists {
		return false
	}
	
	event := Event{
		ID:    generateEventID(),
		Event: eventType,
		Data:  data,
	}
	
	eventBytes, err := b.serializeEvent(event)
	if err != nil {
		return false
	}
	
	return client.Send(eventBytes)
}

// generateEventID生成事件ID
func generateEventID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// GetClientsCount获取当前连接的客户端数
func (b *Broker) GetClientsCount() int {
	b.clientsMu.RLock()
	defer b.clientsMu.RUnlock()
	return len(b.clients)
}

// GetClientIDs获取所有客户端ID
func (b *Broker) GetClientIDs() []string {
	b.clientsMu.RLock()
	defer b.clientsMu.RUnlock()
	
	ids := make([]string, 0, len(b.clients))
	for id := range b.clients {
		ids = append(ids, id)
	}
	
	return ids
}

// Close 关闭代理
func (b *Broker) Close() {
	close(b.stop)
}

// Handler创建SSE处理函数
func Handler(broker *Broker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 从查询参数或路径参数获取客户端ID
		clientID := r.URL.Query().Get("client_id")
		if clientID == "" {
			//如果URL查询参数不存在，使用用户代理或其他唯一标识
			clientID = r.UserAgent()
			if clientID == "" {
				clientID = fmt.Sprintf("client_%d", time.Now().Unix())
			}
		}
		
		// 设置超时头
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		//验证请求方法
		if r.Method != "GET" {
			http.Error(w, "方法不支持", http.StatusMethodNotAllowed)
			return
		}
		
		// 如果浏览器有预检请求头，我们就不处理
		if upgrade := r.Header.Get("Upgrade"); upgrade != "" {
			if upgrade == "websocket" {
				http.Error(w, "不支持WebSocket", http.StatusBadRequest)
				return
			}
		}
		
		//订SSE事件
		broker.Subscribe(clientID, w, r)
	}
}

// EventLogger事件记录器
type EventLogger struct {
	broker *Broker
}

// NewEventLogger创建事件记录器
func NewEventLogger(broker *Broker) *EventLogger {
	return &EventLogger{broker: broker}
}

// Log记录事件
func (l *EventLogger) Log(eventType string, message string, data interface{}) {
	eventData := map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"message":   message,
		"data":      data,
	}
	
	l.broker.Broadcast(eventType, eventData)
}

// LogInfo记录信息事件
func (l *EventLogger) LogInfo(message string, data interface{}) {
	l.Log("info", message, data)
}

// LogError记录错误事件
func (l *EventLogger) LogError(message string, data interface{}) {
	l.Log("error", message, data)
}

// LogDebug记录调试事件
func (l *EventLogger) LogDebug(message string, data interface{}) {
	l.Log("debug", message, data)
}