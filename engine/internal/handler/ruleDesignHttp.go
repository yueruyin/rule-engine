package handler

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"net/http"
	"sort"
	"strings"
	"zenith.engine.com/engine/config"
	"zenith.engine.com/engine/internal"
	"zenith.engine.com/engine/internal/adapter"
	"zenith.engine.com/engine/internal/repository"
	"zenith.engine.com/engine/pkg/e"
	"zenith.engine.com/engine/pkg/util"
)

type SendHttpParam struct {
	System    string                      `json:"system"`
	Method    string                      `json:"method"`
	Url       string                      `json:"url"`
	Query     []HttpParams                `json:"query"`
	Header    []HttpParams                `json:"header"`
	Body      string                      `json:"body"`
	Arguments string                      `json:"arguments"`
	Result    []util.ExecuteMappingResult `json:"result"`
}

type SendHttpResult struct {
	Origin  interface{} `json:"origin"`
	Mapping interface{} `json:"mapping"`
}

type HttpParams struct {
	Selected bool   `json:"selected"`
	Key      string `json:"key"`
	Value    string `json:"value"`
	Desc     string `json:"desc"`
}

//SystemList 获取可用第三方系统
func SystemList(ctx *gin.Context) {
	var systemList []string
	for _, h := range config.Conf.Http {
		systemList = append(systemList, h.System)
	}
	sort.Strings(systemList)
	internal.Success(ctx, systemList, e.GetMsg(e.SUCCESS))
}

//SendHttp 测试http请求
func SendHttp(ctx *gin.Context) {
	var sendHttpParam SendHttpParam
	err := ctx.BindJSON(&sendHttpParam)
	if err != nil {
		internal.Fail(ctx, nil, e.GetMsg(e.InvalidParams))
		return
	}
	var systemInfo config.Http
	h, _ := repository.SystemMap.Load(sendHttpParam.System)
	systemInfo = h.(config.Http)

	// 处理arguments
	//sendHttpParam.Arguments = util.FormatHasArgumentsJson(sendHttpParam.Arguments)
	argumentsMap := make(map[string]interface{})
	if len(sendHttpParam.Arguments) > 0 {
		err = json.Unmarshal([]byte(sendHttpParam.Arguments), &argumentsMap)
		if err != nil {
			internal.Fail(ctx, nil, e.GetMsg(e.InvalidParams))
			return
		}
	}

	// 处理body
	sendHttpParam.Body = util.FormatHasArgumentsJson(sendHttpParam.Body)
	if len(sendHttpParam.Body) > 0 && strings.Contains(sendHttpParam.Body, util.ArgumentsDefiner) {
		bodyMap := make(map[string]interface{})
		err = json.Unmarshal([]byte(sendHttpParam.Body), &bodyMap)
		if err != nil {
			internal.Fail(ctx, nil, e.GetMsg(e.InvalidParams))
			return
		}
		replaceBody := make(map[string]interface{})
		for k, v := range bodyMap {
			k = util.HasArguments(argumentsMap, k)
			v = util.HasArguments(argumentsMap, util.StrVal(v))
			replaceBody[k] = v
		}
		bodyByte, _ := json.Marshal(replaceBody)
		sendHttpParam.Body = string(bodyByte)
	}
	sendHttpParam.Url = formatPathParam(argumentsMap, sendHttpParam.Url)
	req, err := http.NewRequest(sendHttpParam.Method, systemInfo.Host+sendHttpParam.Url, strings.NewReader(sendHttpParam.Body))
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	// query 参数添加
	query := req.URL.Query()
	for _, param := range sendHttpParam.Query {
		if param.Selected {
			param.Key = util.HasArguments(argumentsMap, param.Key)
			param.Value = util.HasArguments(argumentsMap, param.Value)
			query.Add(param.Key, param.Value)
		}
	}
	req.URL.RawQuery = query.Encode()

	// headers 参数添加
	for _, header := range sendHttpParam.Header {
		if header.Selected {
			header.Key = util.HasArguments(argumentsMap, header.Key)
			header.Value = util.HasArguments(argumentsMap, header.Value)
			req.Header.Add(header.Key, header.Value)
		}
	}
	// 没有开启认证直接返回
	if systemInfo.Oauth.Enable {
		_, err = adapter.GetOauthType(systemInfo.Oauth.Type).GetToken(systemInfo, req)
		if err != nil {
			internal.Fail(ctx, nil, e.GetMsg(e.ErrorAuthCheckTokenFail))
			return
		}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		internal.Fail(ctx, nil, e.GetMsg(e.ERROR))
		return
	}
	data, err := ResponseHandle(resp, err)
	var variableMap interface{}
	err = json.Unmarshal(data, &variableMap)

	mapping := make(map[string]interface{})
	for _, rs := range sendHttpParam.Result {
		if rs.Selected {
			rc := gjson.Get(string(data), rs.Path)
			if rc.Exists() && len(rc.String()) > 0 {
				mapping[rs.Title] = rc.String()
			} else {
				mapping[rs.Title] = rs.Def
			}
		}
	}
	internal.Success(ctx, SendHttpResult{
		Origin:  util.ReplaceLong(variableMap),
		Mapping: mapping,
	}, e.GetMsg(e.SUCCESS))
}

// formatPathParam 处理路径参
func formatPathParam(arguments map[string]interface{}, url string) string {
	if strings.ContainsAny(url, util.ArgumentsDefiner) {
		i := strings.Index(url, util.ArgumentsDefiner)
		var v string
		if strings.ContainsAny(url[i:], "/") {
			ix := strings.Index(url[i:], "/")
			v = util.HasArguments(arguments, url[i:][1:ix]) + url[i:][ix:]
		} else {
			v = util.HasArguments(arguments, url[i+1:])
		}
		return formatPathParam(arguments, url[0:i]+v)
	}
	return url
}
