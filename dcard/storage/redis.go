package storage

import (
	"context"       // 用於創建和管理 go 程式的執行上下文
	"dcard/config"  // 連線資料庫會使用到
	"dcard/models"  // 廣告模板
	"encoding/json" // 用於 JSON 的編碼和解碼
	"fmt"
	"sort"    // 提供排序功能
	"strconv" // 用於字符串和其他類型的轉換
	"time"

	"github.com/go-redis/redis/v8" // go-redis ，用於操作 Redis
)

var ctx = context.Background() // 創建一個上下文
var rdb *redis.Client          // 宣告一個 Redis 變數

const DailyAdCreateLimitKey = "daily_ad_create_limit"
const DailyAdCreateLimit = 3000

// InitClient 初始化 Redis 客戶端
func InitClient() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,     // Redis 地址
		Password: config.RedisPassword, // Redis 密碼
		DB:       config.RedisDB,       // Redis 編號
	})
}

// CheckAndIncrementAdCounter 檢查當天創建的廣告數量並嘗試增加計數器
func CheckAndIncrementAdCounter() (bool, error) {
	currentCount, err := rdb.Get(ctx, DailyAdCreateLimitKey).Int()
	if err == redis.Nil {
		// 如果鍵不存在，初始化計數器
		err = rdb.Set(ctx, DailyAdCreateLimitKey, 1, 24*time.Hour).Err()
		if err != nil {
			return false, err
		}
		return true, nil
	} else if err != nil {
		return false, err
	}

	if currentCount >= DailyAdCreateLimit {
		// 如果達到每日限額，返回錯誤
		return false, nil
	}

	// 增加計數器
	err = rdb.Incr(ctx, DailyAdCreateLimitKey).Err()
	if err != nil {
		return false, err
	}
	return true, nil
}

// SaveAd 將廣告存到 Redis 中
func SaveAd(ad *models.Ad) error {
	// 將廣告轉換為 JSON 格式
	adJSON, err := json.Marshal(ad)
	if err != nil {
		return err // 轉換失敗返回錯誤
	}

	// 使用 ctx 作為操作的上下文，將轉換後的 JSON 保存到 Redis，並設定 24 小時銷毀資料
	err = rdb.Set(ctx, ad.Title, adJSON, 24*time.Hour).Err()
	if err != nil {
		return err // 保存失敗返回錯誤
	}

	return nil
}

// GetAds 從 Redis 中獲取符合條件的廣告
func GetAds(queryParams map[string][]string) ([]models.Ad, error) {
	var ads []models.Ad // 廣告列表
	now := time.Now()   // 當前時間

	iter := rdb.Scan(ctx, 0, "*", 0).Iterator() // 使用Iterator遍歷所有	key

	for iter.Next(ctx) { // 遍歷key
		var ad models.Ad
		val, err := rdb.Get(ctx, iter.Val()).Result() // 根據key獲取value
		if err != nil {
			continue // 獲取失敗則跳過
		}
		err = json.Unmarshal([]byte(val), &ad) // 將 JSON 字串解析
		if err != nil {
			continue // 解析失敗則跳過
		}

		// 過濾條件
		matches := true // 預設匹配條件

		// 檢查廣告是否發行中（開始時間 < 現在 < 結束時間）
		if ad.StartAt.After(now) || ad.EndAt.Before(now) {
			continue
		}

		// 根據年齡篩選
		if ageVals, ok := queryParams["age"]; ok && len(ageVals) > 0 {
			age, err := strconv.Atoi(ageVals[0])
			if err != nil {
				return nil, fmt.Errorf("invalid age parameter")
			}
			if age < ad.Conditions.AgeStart || age > ad.Conditions.AgeEnd {
				matches = false
			}
		}

		// 根據性別篩選
		if genderVals, ok := queryParams["gender"]; ok && len(genderVals) > 0 {
			if ad.Conditions.Gender != genderVals[0] {
				matches = false
			}
		}
		// 根據國家篩選
		if countryVals, ok := queryParams["country"]; ok && len(countryVals) > 0 {
			countryMatch := false
			for _, country := range ad.Conditions.Country {
				if country == countryVals[0] {
					countryMatch = true
					break
				}
			}
			if !countryMatch {
				matches = false
			}
		}

		// 根據platform篩選
		if platformVals, ok := queryParams["platform"]; ok && len(platformVals) > 0 {
			platformMatch := false
			for _, platform := range ad.Conditions.Platform {
				if platform == platformVals[0] {
					platformMatch = true
					break
				}
			}
			if !platformMatch {
				matches = false
			}
		}

		if matches {
			ads = append(ads, ad)
		}
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}

	// 按 EndAt 升冪排序
	sort.Slice(ads, func(i, j int) bool {
		return ads[i].EndAt.Before(ads[j].EndAt)
	})

	// 讀取limit和offset參數
	offset, limit := parsePaginationParams(queryParams)
	// print(offset, limit)
	// print(queryParams["offset"], queryParams["limit"])

	if offset > len(ads) {
		return []models.Ad{}, nil // 如果 offset 超出範圍，返回空列表
	}

	end := offset + limit // limit 過大時，讀取全部廣告
	if end > len(ads) {
		end = len(ads)
	}

	return ads[offset:end], nil
}

func parsePaginationParams(queryParams map[string][]string) (offset, limit int) {
	offsetStr, ok := queryParams["offset"]
	// print(offsetStr[0])

	if ok && len(offsetStr) > 0 {
		offset, _ = strconv.Atoi(offsetStr[0]) // 將字符串轉換為整數，錯誤處理被省略
	}

	limitStr, ok := queryParams["limit"]
	// print(limitStr, ok)

	if ok && len(limitStr) > 0 {
		limit, _ = strconv.Atoi(limitStr[0]) // 將字符串轉換為整數，錯誤處理被省略
	}

	if limit <= 0 || limit > 100 {
		limit = 5 // 如果未設置 limit 參數，或設置的值不合理（<=0 或 >100），則預設為 5
	}

	return // 返回計算得到的 offset 和 limit 值
}

// GetRedisClient 返回初始化的 Redis 客戶端
func GetRedisClient() *redis.Client {
	return rdb // 假設 rdb 是在此套件初始化的全局變數
}
