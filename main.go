package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
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
	hbody := handleMarkdown(body)
	class, date, time, _title, err := extractIssueParams(title)
	if err != nil {
		fmt.Println("Error extractIssueParams:", err)
		return
	}

	var titlet string
	var bodyt string
	titlet = "[" + class + "]" + _title
	if class == "茶话会" {
		bodyt = "日期：" + date + " " + "时长：" + time + "\n" + hbody
	} else {
		bodyt = ""
	}
	msg := map[string]string{
		"body":     bodyt,
		"title":    titlet,
		"username": username,
		"url":      url,
	}
	//群号
	group_id := 12345678
	// 通过 WebSocket 向 QQ 机器人发送消息
	// test := true
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

	body := message["title"] + "\n" + message["body"] + "\n" + "链接：" + message["url"]
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

	body := message["title"] + "\n" + message["body"] + "\n" + "链接：" + message["url"]

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

// 清理 Markdown 文本，处理 # 和 -
func cleanMarkdown(markdown string) string {
	// 处理标题部分，去除开头的 # 符号
	reTitle := regexp.MustCompile(`^#+\s*`)
	markdown = reTitle.ReplaceAllString(markdown, "")

	// 将 - 替换为 ·
	markdown = strings.ReplaceAll(markdown, "-", "·")

	return markdown
}

func handleMarkdown(markdown string) string {
	// 分割 Markdown 文本
	lines := strings.Split(markdown, "\n")

	// 定义一个字符串来存储结果
	var result strings.Builder

	// 标记“内容类型”是否处理
	inContentType := false

	// 遍历每一行并进行处理
	for i := 0; i < len(lines); i++ {
		// 如果遇到 "关闭 Issue 前请先确认以下内容"，则停止处理
		if strings.Contains(lines[i], "关闭 Issue 前请先确认以下内容") {
			break
		}

		// 如果当前行是 "### 作者"
		if strings.TrimSpace(lines[i]) == "### 作者" {
			// 找到下一个非空行，将其视为“作者”的内容
			for j := i + 1; j < len(lines); j++ {
				if strings.TrimSpace(lines[j]) != "" {
					authorLine := "宣讲人：" + strings.TrimSpace(lines[j])
					result.WriteString(authorLine + "\n")
					i = j // 更新 i，跳过已处理的行
					break
				}
			}
			continue
		}

		// 如果当前行是 "### 内容类型"
		if strings.TrimSpace(lines[i]) == "### 内容类型" {
			inContentType = true
			continue
		}

		// 如果在 "内容类型" 部分，处理选项
		if inContentType {
			if strings.Contains(lines[i], "- [X] 技术类") {
				result.WriteString("内容类型：技术类\n")
				continue
			} else if strings.Contains(lines[i], "- [X] 其他") {
				result.WriteString("内容类型：其他\n")
				inContentType = false // 处理完内容类型，退出
				continue
			} else if strings.Contains(lines[i], "- [ ]") {
				// 忽略未选中的选项
				if !strings.Contains(lines[i+1], "- [X]") {
					inContentType = false
				}
				continue
			}
		}

		// 如果当前行是 "### 摘要/大纲"
		if strings.TrimSpace(lines[i]) == "### 摘要/大纲" {
			result.WriteString("摘要/大纲：\n")
			continue
		}

		// 如果当前行是 "### 补充说明"
		if strings.TrimSpace(lines[i]) == "### 补充说明" {
			// 找到下一个非空行，将其视为“作者”的内容
			for j := i + 1; j < len(lines); j++ {
				if strings.TrimSpace(lines[j]) != "" {
					if strings.TrimSpace(lines[j]) == "_No response_" {
						result.WriteString("补充说明：无\n")
						i = j // 更新 i，跳过已处理的行
						break
					} else {
						result.WriteString("补充说明：\n")
						break
					}
				}
			}
			continue
		}

		// 清理当前行
		cleaned := cleanMarkdown(lines[i])

		// 如果清理后的行不为空，添加到结果中
		if cleaned != "" {
			result.WriteString(cleaned + "\n")
		}
	}

	// 将最终结果返回为字符串
	return result.String()
}

// 提取 issue 标题中的参数
func extractIssueParams(title string) (string, string, string, string, error) {
	// 先匹配第一个方括号内的内容
	initialRe := regexp.MustCompile(`^\[(.*?)\]`)
	initialMatch := initialRe.FindStringSubmatch(title)

	// 只要有匹配结果就继续处理
	if len(initialMatch) > 0 {
		firstParam := initialMatch[1]

		// 如果第一个方括号内容是 "茶话会"，按照四个参数提取
		if firstParam == "茶话会" {
			// 匹配 [分类][日期][时长] 标题 格式
			fullRe := regexp.MustCompile(`\[(.*?)\]\s*\[(.*?)\]\s*\[(.*?)\]\s*(.*)`)
			matches := fullRe.FindStringSubmatch(title)

			if len(matches) == 5 {
				return matches[1], matches[2], matches[3], matches[4], nil
			}
			return "", "", "", "", fmt.Errorf("failed to extract parameters in '茶话会' format")
		}

		// 如果不是 "茶话会"，按照 [分类] 标题 格式提取
		simpleRe := regexp.MustCompile(`\[(.*?)\]\s*(.*)`)
		matches := simpleRe.FindStringSubmatch(title)

		if len(matches) == 3 {
			return matches[1], "", "", matches[2], nil
		}
		return "", "", "", "", fmt.Errorf("failed to extract parameters in '[分类] 标题' format")
	}

	return "", "", "", "", fmt.Errorf("failed to extract initial parameter")
}
