package repository

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"zenith.engine.com/engine/config"
)

// OauthPassword 纵横用户密码登陆方式
type OauthPassword struct {
}

func (*OauthPassword) GetToken(systemInfo config.Http, r *http.Request) (string, error) {
	ssoUrl := systemInfo.Host + systemInfo.Oauth.Url
	resp, err := http.Post(ssoUrl, "application/json", strings.NewReader("{'account':'"+systemInfo.Oauth.UserName+"','password':'"+systemInfo.Oauth.Password+"'}"))
	if err != nil {
		return "", err
	}
	var result map[string]interface{}
	b, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(b, &result)
	data := result["data"].(map[string]interface{})
	token := data["token"].(string)
	r.Header.Add("token", token)
	return token, err
}
