package main

import (
	"dcard/handlers" // 處理 HTTP 請求的 handlers 套件
	"dcard/storage"  // 資料儲存相關的 storage 套件

	"github.com/gin-gonic/gin" // Gin Web 框架
)

func main() {

	storage.InitClient() // 初始化 Redis 連線

	router := gin.Default() // 創建一個默認的 Gin router

	// 處理 POST 請求，用於創建新的廣告
	router.POST("/api/v1/ad", handlers.CreateAd)
	// 處理 GET 請求，用於列出符合條件的廣告
	router.GET("/api/v1/ad", handlers.ListAds)

	// 處理 POST 請求，用於重置廣告創建計數器(僅供測試使用)
	router.POST("/api/v1/admin/reset-ad-counter", handlers.ResetAdCreationCounter)

	router.Run(":8080") // 在 8080 端口啟動服務
}
