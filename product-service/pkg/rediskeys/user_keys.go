package rediskeys

import "fmt"

func UserInfoKey(id int64) string      { return fmt.Sprintf("user:info:%d", id) }
func DailyLoginKey(date string) string { return "user:login:" + date }
