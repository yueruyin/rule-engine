package repository

import (
	"net/http"
	"zenith.engine.com/engine/config"
)

type OauthClient struct {
}

func (*OauthClient) GetToken(systemInfo config.Http, r *http.Request) (string, error) {

	return "", nil
}
