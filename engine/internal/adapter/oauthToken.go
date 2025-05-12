package adapter

import (
	"net/http"
	"zenith.engine.com/engine/config"
	"zenith.engine.com/engine/internal/repository"
)

type OauthToken interface {
	// GetToken 获取token方法,如需特定增加实现
	GetToken(systemInfo config.Http, r *http.Request) (string, error)
}

// GetOauthType 认证类型获取对应认证实现 如:
//  password账号密码模式
//  client客户端模式
//  token直接使用第三方提供token
func GetOauthType(t string) OauthToken {
	var oauthToken OauthToken
	// 开启了认证:自动获取token
	switch t {
	default:
		fallthrough
	case "password":
		oauthToken = &repository.OauthPassword{}
	case "client":
		oauthToken = &repository.OauthClient{}
	case "token":
		oauthToken = &repository.OauthToken{}
	}
	return oauthToken
}
