package handlers

import (
	"dcard/storage"
	"net/http"
	"time"

	"github.com/gin-gonic/gin" // Gin Web 框架
)

// ResetAdCreationCounter 重置廣告創建計數器
func ResetAdCreationCounter(c *gin.Context) {
	// 使用請求的上下文來進行操作
	ctx := c.Request.Context()

	// 調用 storage 包中的 Redis 來設置計數器
	err := storage.GetRedisClient().Set(ctx, storage.DailyAdCreateLimitKey, 0, 24*time.Hour).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset ad creation counter"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Ad creation counter has been reset"})
}
