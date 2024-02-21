package handlers

import (
	"dcard/models"  // 廣告模板
	"dcard/storage" // 儲存邏輯，負責與資料庫互動
	"net/http"      // 處理 HTTP 請求與響應的工具

	"github.com/biter777/countries" // 引入國家代碼處理套件，用於驗證國家代碼是否有效
	"github.com/gin-gonic/gin"      // Gin Web 框架套件
)

// CreateAd 處理創建廣告的請求
func CreateAd(c *gin.Context) {
	var ad models.Ad
	// 確保請求的 JSON 格式正確
	if err := c.ShouldBindJSON(&ad); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"}) // 如果請求體格式不正確，返回 400 狀態碼及錯誤信息
		return
	}

	// 如果設定了年齡範圍，驗證年齡範圍是否在 1 到 100 歲之間
	if (ad.Conditions.AgeStart != 0 || ad.Conditions.AgeEnd != 0) &&
		(ad.Conditions.AgeStart < 1 || ad.Conditions.AgeStart > 100 ||
			ad.Conditions.AgeEnd < 1 || ad.Conditions.AgeEnd > 100) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Age range must be between 1 and 100"}) // 年齡範圍不符合要求時，返回錯誤信息
		return
	}

	// 如果設定了性別，驗證性別是否為 "M" 或 "F"
	if ad.Conditions.Gender != "" && ad.Conditions.Gender != "M" && ad.Conditions.Gender != "F" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gender must be 'M' or 'F'"}) // 性別不符合要求時，返回錯誤信息
		return
	}

	// 如果設定了國家代碼，驗證國家代碼是否有效
	for _, country := range ad.Conditions.Country {
		if country != "" && !validateCountry(country) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid country code"}) // 國家代碼無效時，返回錯誤信息
			return
		}
	}

	// 如果設定了平台，驗證平台是否為 "android"、"ios" 或 "web"
	validPlatforms := map[string]bool{"android": true, "ios": true, "web": true}
	for _, platform := range ad.Conditions.Platform {
		if platform != "" && !validPlatforms[platform] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid platform"}) // 平台不符合要求時，返回錯誤信息
			return
		}
	}
	// 在存儲廣告之前，檢查每日創建限額
	allowed, err := storage.CheckAndIncrementAdCounter()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check daily ad creation limit"})
		return
	}
	if !allowed {
		c.JSON(http.StatusForbidden, gin.H{"error": "Daily ad creation limit reached"})
		return
	}

	// 將廣告存儲到資料庫
	if err := storage.SaveAd(&ad); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}) // 存儲失敗時，返回 500 狀態碼及錯誤信息
		return
	}

	// 廣告創建成功，返回 201 狀態碼及廣告信息
	c.JSON(http.StatusCreated, gin.H{"status": "success", "ad": ad})
}

// validateCountry 國家代碼驗證函數
func validateCountry(countryCode string) bool {
	// 使用 countries 來檢查國家代碼是否有效
	country := countries.ByName(countryCode) // 根據國家名稱獲取國家對象
	return country != countries.Unknown      // 如果返回值是 countries.Unknown，則國家代碼無效
}
