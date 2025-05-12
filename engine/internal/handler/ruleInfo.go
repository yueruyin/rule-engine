package handler

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"zenith.engine.com/engine/internal"
	"zenith.engine.com/engine/internal/repository"
	"zenith.engine.com/engine/pkg/e"
)

func InitDefaultRule() {
	ruleList := repository.FindRuleInfoList()
	for _, info := range ruleList {
		pool := BuildEnginePool(info.Script)
		SetPool(info.Code, pool)
	}
}

func ListDefaultInfo(ctx *gin.Context) {
	ruleList := repository.FindRuleInfoList()
	internal.Success(ctx, ruleList, e.GetMsg(e.SUCCESS))
}

func GetDefaultInfo(ctx *gin.Context) {
	code := ctx.Query("code")
	var ri repository.RuleInfo
	internal.Success(ctx, ri.GetByCode(code), e.GetMsg(e.SUCCESS))
}

func SaveDefaultInfo(ctx *gin.Context) {
	var ri repository.RuleInfo
	err := ctx.BindJSON(&ri)
	if err != nil {
		internal.Fail(ctx, nil, err.Error())
		return
	}
	createError := ri.Create(&ri)
	if createError != nil {
		internal.Fail(ctx, nil, createError.Error())
		return
	}
	pool := BuildEnginePool(ri.Script)
	SetPool(ri.Code, pool)
	internal.Success(ctx, nil, e.GetMsg(e.SUCCESS))
}

func UpdateDefaultInfo(ctx *gin.Context) {
	var ri repository.RuleInfo
	err := ctx.BindJSON(&ri)
	if err != nil {
		internal.Fail(ctx, nil, err.Error())
		return
	}
	updateError := ri.Update(&ri)
	if updateError != nil {
		internal.Fail(ctx, nil, updateError.Error())
		return
	}
	DeleteDefaultPool(ri.Code)
	pool := BuildEnginePool(ri.Script)
	SetPool(ri.Code, pool)
	internal.Success(ctx, nil, e.GetMsg(e.SUCCESS))
}

func DeleteDefaultInfo(ctx *gin.Context) {
	id := ctx.Param("id")
	uid, _ := strconv.ParseUint(id, 10, 64)
	var ri repository.RuleInfo
	// 删掉内存中的默认规则
	info := ri.GetById(uid)
	DeleteDefaultPool(info.Code)
	err := ri.Delete(uid)
	if err != nil {
		internal.Fail(ctx, nil, err.Error())
		return
	}
	internal.Success(ctx, nil, e.GetMsg(e.SUCCESS))
}
