package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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

func main() {
	// 初始化 Gin 路由
	router := gin.Default()

	// GitHub Webhook 路由
	router.POST("/webhook", handleWebhook)

	// WebSocket 路由
	router.GET("/ws", handleWebSocket)

	// 运行服务器
	router.Run(":8080")
}

// 处理 GitHub Webhook 请求
func handleWebhook(c *gin.Context) {
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 打印收到的 GitHub Webhook 数据
	fmt.Printf("Received Webhook: %v\n", payload)

	// 处理 Webhook 数据（如提取 Issue 信息）
	action, _ := payload["action"].(string)
	issue, _ := payload["issue"].(map[string]interface{})
	title, _ := issue["title"].(string)
	user, _ := issue["user"].(map[string]interface{})
	username, _ := user["login"].(string)
	body, _ := issue["body"].(string)
	url, _ := issue["html_url"].(string)
	// 构建要发送的消息
	msg := map[string]string{
		"body":     body,
		"title":    title,
		"username": username,
		"url":      url,
	}
	//群号
	group_id := 914590482
	// 通过 WebSocket 向 QQ 机器人发送消息

	if action == "opened" {
		message_id, err := sendMessageToWebSocket(msg, int64(group_id))

		if err != nil {
			fmt.Println("Error sending group message:", err)
			return
		}

		//设置为精华消息
		err = setEssenceMsg(message_id)
		if err != nil {
			fmt.Println("Error setting essence message:", err)
			return
		}

		//设置为群公告
		err = setGroupNotice(msg, int64(group_id))
		if err != nil {
			fmt.Println("Error setting essence message:", err)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "Webhook received"})
}

// 处理 WebSocket 连接
func handleWebSocket(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade connection"})
		return
	}
	defer ws.Close()

	for {
		// 读取 WebSocket 消息
		_, msg, err := ws.ReadMessage()
		if err != nil {
			fmt.Println("Read Error:", err)
			break
		}
		fmt.Printf("Received from WebSocket: %s\n", msg)
	}
}

// 向 WebSocket 发送消息
func sendMessageToWebSocket(message map[string]string, group_id int64) (int32, error) {
	// 连接到 QQ 机器人的 WebSocket 服务器
	ws, _, err := websocket.DefaultDialer.Dial("ws://localhost:3001/", nil) // 替换为实际 QQ 机器人的 WebSocket 地址
	if err != nil {
		fmt.Println("Dial Error:", err)
		return 0, err
	}
	defer ws.Close()

	// 将github issue中的消息利用接口发送到qq

	body := "标题：" + message["title"] + "\n" + "宣讲人：" + message["username"] + "\n" + "内容：" + message["body"] + "\n" + "链接：" + message["url"]
	msg := PrivateMessage{
		Action: "send_group_msg",
		Params: MessageParams{
			Group_id: group_id, // 替换为实际的 QQ 用户ID
			Message:  body,     // 要发送的消息内容
		},
		Echo: "send_msg",
	}

	jsonData, err := json.Marshal(msg)

	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return 0, err
	}

	// 通过 WebSocket 发送消息到 OneBot
	err = ws.WriteMessage(websocket.TextMessage, jsonData)
	if err != nil {
		fmt.Println("Write Error:", err)
		return 0, err
	}

	// 读取响应
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			return 0, fmt.Errorf("read error: %v", err)
		}

		// 解析响应
		var response map[string]interface{}
		if err := json.Unmarshal(msg, &response); err != nil {
			return 0, fmt.Errorf("JSON unmarshal error: %v", err)
		}

		// 检查是否为发送消息的响应
		if response["echo"] == "send_msg" && response["retcode"].(float64) == 0 {
			messageID := int32(response["data"].(map[string]interface{})["message_id"].(float64))
			fmt.Println("Message sent successfully, message ID:", messageID)
			return messageID, nil
		}
	}

}

// 向 WebSocket 发送精华消息设置请求
func setEssenceMsg(messageID int32) error {
	// 连接到 QQ 机器人的 WebSocket 服务器
	ws, _, err := websocket.DefaultDialer.Dial("ws://localhost:3001/", nil) // 替换为实际 QQ 机器人的 WebSocket 地址
	if err != nil {
		fmt.Println("Dial Error:", err)
		return err
	}
	defer ws.Close()

	req := EssenceMessage{
		Action: "set_essence_msg",
		Params: eMessageParams{
			Message_id: messageID,
		},
	}

	jsonData, err := json.Marshal(req)

	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return err
	}

	// 发送请求
	if err := ws.WriteMessage(websocket.TextMessage, jsonData); err != nil {
		fmt.Println("Error setting essence message:", err)
		return err
	}

	fmt.Println("Essence set successfully!")

	// 保持连接 1 秒
	time.Sleep(1 * time.Second)

	return nil
}

// 向 WebSocket 发送群公告设置请求
func setGroupNotice(message map[string]string, groupID int64) error {
	// 连接到 QQ 机器人的 WebSocket 服务器
	ws, _, err := websocket.DefaultDialer.Dial("ws://localhost:3001/", nil) // 替换为实际 QQ 机器人的 WebSocket 地址
	if err != nil {
		fmt.Println("Dial Error:", err)
		return err
	}
	defer ws.Close()

	body := "标题：" + message["title"] + "\n" + "宣讲人：" + message["username"] + "\n" + "内容：" + message["body"] + "\n" + "链接：" + message["url"]

	req := Group_notice{
		Action: "_send_group_notice",
		Params: gMessageParams{
			Group_id: groupID,
			Content:  body,
		},
		Echo: "send_notice",
	}

	jsonData, err := json.Marshal(req)

	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return err
	}

	// 发送请求
	if err := ws.WriteMessage(websocket.TextMessage, jsonData); err != nil {
		fmt.Println("Error setting group notice:", err)
		return err
	}

	fmt.Println("Group_notice set successfully!")

	// 保持连接 1 秒
	time.Sleep(1 * time.Second)

	return nil
}
