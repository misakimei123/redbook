package ioc

import (
	"os"

	"github.com/misakimei123/redbook/internal/service/oauth2/wechat"
)

func InitWechatService() wechat.Service {
	appID, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		panic("找不到环境变量 WECHAT_APP_ID")
	}
	appSecret, _ := os.LookupEnv("WECHAT_SECRET")
	return wechat.NewWechatService(appID, appSecret)
}
