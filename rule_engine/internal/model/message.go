package model

// Message 消息结构
type Message struct {
	Event string `json:"event"` // 事件类型
	Data  any    `json:"data"`  // 事件数据
}
