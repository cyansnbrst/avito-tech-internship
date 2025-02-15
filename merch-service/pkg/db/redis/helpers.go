package redis

import "fmt"

func GetUserInfoCacheKey(userID int64) string {
	return fmt.Sprintf("user:%d:info", userID)
}
