package repository

import (
	"net/http"
	"zenith.engine.com/engine/config"
)

type OauthToken struct {
}

func (*OauthToken) GetToken(systemInfo config.Http, r *http.Request) (string, error) {
	return "", nil
}
