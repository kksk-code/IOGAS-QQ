package main

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// WebSocket 升级器配置
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源的连接，实际使用中应加以限制
	},
}

type Config struct {
	ServerPort   string `json:"server_port"`
	GroupID      int64  `json:"group_id"`
	WebSocketURL string `json:"websocket_url"`
}

type PrivateMessage struct {
	Action string        `json:"action"`
	Params MessageParams `json:"params"`
	Echo   string        `json:"echo"`
}

type MessageParams struct {
	Group_id int64  `json:"group_id"`
	Message  string `json:"message"`
}

type eMessageParams struct {
	Message_id int32 `json:"message_id"`
}

type gMessageParams struct {
	Group_id int64  `json:"group_id"`
	Content  string `json:"content"`
}

type EssenceMessage struct {
	Action string         `json:"action"`
	Params eMessageParams `json:"params"`
}

type Group_notice struct {
	Action string         `json:"action"`
	Params gMessageParams `json:"params"`
	Echo   string         `json:"echo"`
}

var config *Config
