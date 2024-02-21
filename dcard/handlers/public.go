package handlers

import (
	"dcard/storage" // 用於存取廣告數據
	"fmt"
	"net/http" // 處理 HTTP 請求與響應的工具
	"strconv"  // 資料轉換

	// 用於字符串和其他類型之間的轉換
	"github.com/gin-gonic/gin" // Gin Web 框架
)

// ListAds 處理 GET 請求，用於列出符合條件的廣告
func ListAds(c *gin.Context) {
	queryParams := c.Request.URL.Query() // 獲取查詢參數

	// 定義允許的參數名稱
	allowedParams := map[string]bool{
		"age":      true,
		"gender":   true,
		"country":  true,
		"platform": true,
		"limit":    true,
		"offset":   true,
	}

	// 檢查是否有不允許的參數名稱
	for param := range queryParams {
		if _, exists := allowedParams[param]; !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid parameter name: '%s'", param)})
			return
		}
	}

	// 驗證 'age' 查詢參數
	if ageVals, ok := queryParams["age"]; ok {
		if len(ageVals[0]) > 0 {
			age, err := strconv.Atoi(ageVals[0])
			if err != nil || age < 1 || age > 100 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'age' parameter, must be an integer between 1 and 100"})
				return
			}
		}
	}

	// 驗證 'gender' 查詢參數
	if genderVals, ok := queryParams["gender"]; ok {
		if genderVals[0] != "M" && genderVals[0] != "F" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'gender' parameter, must be 'M' or 'F'"})
			return
		}
	}

	if countryVals, ok := queryParams["country"]; ok && len(countryVals[0]) > 0 {
		// 嘗試將 country 參數轉換為整數，以檢查是否為純數字字符串
		if _, err := strconv.Atoi(countryVals[0]); err == nil {
			// strconv.Atoi 沒有返回錯誤，表示它是純數字字符串
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'country' parameter, must be a non-numeric string"})
			return
		}
		// 如果 strconv.Atoi 返回錯誤，則不做任何事情，因為這表示 countryVals[0] 不是純數字字符串
	}

	// 驗證 'platform' 查詢參數
	if platformVals, ok := queryParams["platform"]; ok {
		if platformVals[0] != "ios" && platformVals[0] != "web" && platformVals[0] != "android" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'platform' parameter, must be 'ios', 'web' or 'android'"})
			return
		}
	}

	// 驗證 'limit' 查詢參數
	if limitVals, ok := queryParams["limit"]; ok {
		if len(limitVals[0]) > 0 {
			limit, err := strconv.Atoi(limitVals[0])
			if err != nil || limit <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'limit' parameter, must be an integer greater than 0"})
				return
			}
		}
	}

	// 驗證 'offset' 查詢參數
	if offsetVals, ok := queryParams["offset"]; ok {
		if len(offsetVals[0]) > 0 {
			offset, err := strconv.Atoi(offsetVals[0])
			if err != nil || offset < 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'offset' parameter, must be an integer greater than or equal to 0"})
				return
			}
		}
	}

	ads, err := storage.GetAds(queryParams) // 從 storage 中根據查詢參數獲取廣告列表
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}) // 如果出現錯誤，返回 500 內部服務器錯誤和錯誤信息
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": ads}) // 返回 200 狀態碼和符合條件的廣告列表
}
