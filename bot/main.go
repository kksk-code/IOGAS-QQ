package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {

	// 加载配置文件
	var err error
	config, err = loadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化 Gin 路由
	router := gin.Default()

	// GitHub Webhook 路由
	router.POST("/webhook", handleWebhook)

	// WebSocket 路由
	router.GET("/ws", handleWebSocket)

	// 静态图片文件路由
	router.Static("/images", "./images")

	// 运行服务器
	router.Run(":" + config.ServerPort)
}
