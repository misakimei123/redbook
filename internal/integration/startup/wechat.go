package startup

import (
	"github.com/misakimei123/redbook/internal/service/oauth2/wechat"
)

func InitWechatService() wechat.Service {
	return wechat.NewWechatService("appID", "appSecret")
}
