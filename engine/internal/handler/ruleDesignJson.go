package handler

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"zenith.engine.com/engine/internal"
	"zenith.engine.com/engine/pkg/e"
	"zenith.engine.com/engine/pkg/util"
)

type PreviewFormatJsonParam struct {
	JsonModel string                      `json:"jsonModel,omitempty"`
	Result    []util.ExecuteMappingResult `json:"result,omitempty"`
	SetResult []util.ExecuteSetResult     `json:"setResult,omitempty"`
}

// PreviewFormatJson 预览处理json结果
func PreviewFormatJson(ctx *gin.Context) {
	var pj PreviewFormatJsonParam
	err := ctx.BindJSON(&pj)
	if err != nil {
		internal.Fail(ctx, nil, err.Error())
		return
	}
	// 验证是否是json格式字符串
	if !json.Valid([]byte(pj.JsonModel)) {
		internal.Fail(ctx, nil, e.GetMsg(e.JsonFormatError))
		return
	}
	// 获取json中值
	mapping := make(map[string]interface{})
	for _, result := range pj.Result {
		if !result.Selected {
			continue
		}
		if gjson.Get(pj.JsonModel, result.Path).Exists() {
			mapping[result.Title] = gjson.Get(pj.JsonModel, result.Path).String()
		} else {
			mapping[result.Title] = result.Def
		}
	}
	// 设置json中值
	for _, sr := range pj.SetResult {
		if !sr.Selected {
			continue
		}
		if gjson.Valid(sr.Value.(string)) {
			var vm interface{}
			err = json.Unmarshal([]byte(sr.Value.(string)), &vm)
			pj.JsonModel, err = sjson.Set(pj.JsonModel, sr.Key, vm)
		} else {
			pj.JsonModel, err = sjson.Set(pj.JsonModel, sr.Key, sr.Value)
		}
	}
	var origin interface{}
	err = json.Unmarshal([]byte(pj.JsonModel), &origin)
	if err != nil {
		internal.Fail(ctx, nil, e.GetMsg(e.JsonFormatError))
		return
	}
	internal.Success(ctx, SendHttpResult{Origin: origin, Mapping: mapping}, e.GetMsg(e.SUCCESS))
}
