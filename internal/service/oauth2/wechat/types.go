package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/misakimei123/redbook/internal/domain"
)

var redirectURL = url.PathEscape("https://misaki.com/oauth2/wechat/callback")

type Service interface {
	AuthURL(ctx context.Context, state string) string
	VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error)
}

type service struct {
	appID      string
	appSecret  string
	httpClient *http.Client
}

type Result struct {
	AccessToken string `json:"access_token"`
	// access_token接口调用凭证超时时间，单位（秒）
	ExpiresIn int64 `json:"expires_in"`
	// 用户刷新access_token
	RefreshToken string `json:"refresh_token"`
	// 授权用户唯一标识
	OpenId string `json:"openid"`
	// 用户授权的作用域，使用逗号（,）分隔
	Scope string `json:"scope"`
	// 当且仅当该网站应用已获得该用户的userinfo授权时，才会出现该字段。
	UnionId string `json:"unionid"`

	// 错误返回
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (s *service) AuthURL(ctx context.Context, state string) string {
	// 微信扫码登录页面url
	const authURLPattern = `https://open.weixin.qq.com/connect/qrconnect?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_login&state=%s#wechat_redirect`
	return fmt.Sprintf(authURLPattern, s.appID, redirectURL, state)
}

func (s *service) VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error) {
	accessTokenUrl := fmt.Sprintf(`https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code`,
		s.appID, s.appSecret, code)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, accessTokenUrl, nil)
	if err != nil {
		return domain.WechatInfo{}, err
	}
	response, err := s.httpClient.Do(request)
	if err != nil {
		return domain.WechatInfo{}, err
	}
	var result Result
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return domain.WechatInfo{}, err
	}
	if result.ErrCode != 0 {
		return domain.WechatInfo{}, fmt.Errorf("调用微信接口失败 errcode %d, errmsg %s", result.ErrCode, result.ErrMsg)
	}
	return domain.WechatInfo{OpenId: result.OpenId, UnionId: result.UnionId}, nil
}

func NewWechatService(appID, appSecret string) Service {
	return &service{appID: appID, appSecret: appSecret, httpClient: http.DefaultClient}
}
