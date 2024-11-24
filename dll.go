package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func getimg(input, filename string) (string, error) {
	body := ImageInput{Input: input}
	url := config.MdToImgURL
	data, err := json.Marshal(body)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return "", err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("error reading response body: %v", err)
		}

		// 创建路径
		date := time.Now().Format("2006-01-02")
		path := "./images/" + date
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return "", fmt.Errorf("error creating directory: %v", err)
		}

		err = ioutil.WriteFile(path+"/"+filename, body, 0644)
		if err != nil {
			return "", fmt.Errorf("error saving image: %v", err)
		}

		return filename, nil
	} else {
		return "", fmt.Errorf("error response: %v", resp.Status)
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
	ws, _, err := websocket.DefaultDialer.Dial(config.WebSocketURL, nil) // 替换为实际 QQ 机器人的 WebSocket 地址
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

// 清理 Markdown 文本，处理 # 和 -
func cleanMarkdown(markdown string) string {
	// 处理标题部分，去除开头的 # 符号
	reTitle := regexp.MustCompile(`^#+\s*`)
	markdown = reTitle.ReplaceAllString(markdown, "")

	// 将 - 替换为 ·
	markdown = strings.ReplaceAll(markdown, "-", "·")

	return markdown
}

// 向 WebSocket 发送消息
func sendMessageToWebSocket(imgfile string, group_id int64) (int32, error) {
	// 连接到 QQ 机器人的 WebSocket 服务器
	ws, _, err := websocket.DefaultDialer.Dial(config.WebSocketURL, nil) // 替换为实际 QQ 机器人的 WebSocket 地址
	if err != nil {
		fmt.Println("Dial Error:", err)
		return 0, err
	}
	defer ws.Close()

	// 将 github issue中的消息利用接口发送到qq

	date := time.Now().Format("2006-01-02")
	//body := message["title"] + "\n" + message["body"] + "\n" + "链接：" + message["url"] + "\n感兴趣请加入后援群 291694149 交流"
	msg := PrivateMessage{
		Action: "send_group_msg",
		Params: MessageParams{
			Group_id: group_id,                                                                    // 替换为实际的 QQ 用户 ID
			Message:  "[CQ:image,file=http://localhost:8080/images/" + date + "/" + imgfile + "]", // 要发送的消息内容
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

// 清理文件名，去除非法字符
func cleanFileName(fileName string) string {
	// 定义非法字符的正则表达式
	re := regexp.MustCompile(`[<>:"/\\|?*]`)
	// 替换非法字符为空字符串
	cleanedFileName := re.ReplaceAllString(fileName, "")
	// 去除文件名开头和结尾的空格
	cleanedFileName = strings.TrimSpace(cleanedFileName)
	return cleanedFileName
}
