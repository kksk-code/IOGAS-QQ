package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

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
	hbody := handleMarkdown(body)
	class, date, time, _title, err := extractIssueParams(title)
	if err != nil {
		fmt.Println("Error extractIssueParams:", err)
		return
	}

	filename := cleanFileName(_title) + ".png"

	// 构建消息
	var titlet string
	var bodyt string
	var ititle string
	titlet = "[" + class + "]" + _title
	switch class {
	case "茶话会":
		bodyt = "日期：" + date + " " + "时长：" + time + "\n" + hbody
		ititle = "## " + _title + "\r\n" + "日期：" + date + " " + "时长：" + time + "\r\n"
	default:
		bodyt = ""
		timeinit()
		ititle = "## " + class + _title + "\r\n" + "日期：" + currentDate + "\r\n"
	}

	// 获取图片
	_, err = getimg(ititle+body, filename)
	if err != nil {
		fmt.Println("Error getimg:", err)
		return
	}

	msg := map[string]string{
		"body":     bodyt,
		"title":    titlet,
		"username": username,
		"url":      url,
		"date":     date,
	}
	//群号
	group_id := config.GroupID
	// 通过 WebSocket 向 QQ 机器人发送消息
	// test := true
	if action == "opened" {
		message_id, err := sendMessageToWebSocket(filename, int64(group_id))

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

// 加载配置文件
func loadConfig(filename string) (*Config, error) {
	// 打开文件
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 读取文件内容
	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// 解析 JSON
	var config Config
	if err := json.Unmarshal(bytes, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
