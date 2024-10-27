package domain

import "time"

type User struct {
	Id         int64
	Email      string
	Password   string
	Nick       string
	AboutMe    string
	Birthday   time.Time
	Phone      string
	WechatInfo WechatInfo
}
