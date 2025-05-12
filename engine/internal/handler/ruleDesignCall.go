package handler

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"strconv"
	"zenith.engine.com/engine/internal"
	"zenith.engine.com/engine/internal/adapter"
	"zenith.engine.com/engine/internal/repository"
	"zenith.engine.com/engine/pkg/e"
	"zenith.engine.com/engine/pkg/util"
)

type RuleGroupAndDesignResult struct {
	Id       uint64                     `json:"id"`
	Code     string                     `json:"code"`
	Name     string                     `json:"name"`
	Type     string                     `json:"type"`
	Children []RuleGroupAndDesignResult `json:"children"`
}

// ListByGroupId 通过分组id查询分组及规则设计(全加载)
func ListByGroupId(ctx *gin.Context) {
	id := ctx.Query("id")
	groupId, _ := strconv.ParseUint(id, 10, 64)
	internal.Success(ctx, ListByGroupIdFor(groupId), e.GetMsg(e.SUCCESS))
}

// ListByGroupIdFor 获取
func ListByGroupIdFor(groupId uint64) []RuleGroupAndDesignResult {
	var rg repository.RuleGroup
	var rd repository.RuleDesign
	ruleGroupAndDesignResult := make([]RuleGroupAndDesignResult, 0)
	rgl := rg.ParentGroup(groupId)
	rdl := rd.ListByGroupId(groupId, 0)
	for _, group := range rgl {
		ruleGroupAndDesignResult = append(ruleGroupAndDesignResult,
			RuleGroupAndDesignResult{Id: group.Id, Name: group.Name, Code: group.Code, Type: "group", Children: ListByGroupIdFor(group.Id)})
	}
	for _, design := range rdl {
		ruleGroupAndDesignResult = append(ruleGroupAndDesignResult,
			RuleGroupAndDesignResult{Id: design.Id, Name: design.Name, Code: design.Code, Type: "rule"})
	}
	return ruleGroupAndDesignResult
}

// PreviewExecuteRule 预执行规则
func PreviewExecuteRule(ctx *gin.Context) {
	var ep ExecuteRuleParam
	err := ctx.BindJSON(&ep)
	ep.Env = util.DEV
	if err != nil {
		internal.Fail(ctx, nil, err.Error())
		return
	}

	// 处理参数
	argumentsMap := make(map[string]interface{})
	for _, arg := range ep.CallArguments {
		if arg.Selected {
			key := util.HasArguments(ep.Arguments, arg.Key)
			value := util.HasArguments(ep.Arguments, arg.Value)
			if len(value) == 0 {
				value = arg.Def
			}
			argumentsMap[key] = value
		}
	}
	ep.Arguments = argumentsMap

	// 执行规则
	resultMap := execute(ep)
	r, _ := resultMap.Load("executeRuleResult")
	er := r.(ExecuteRuleResult)
	erJson, _ := json.Marshal(er)
	// 返回映射返回值
	mapping := make(map[string]interface{})
	for _, rs := range ep.Result {
		if rs.Selected {
			rc := gjson.Get(string(erJson), rs.Path)
			if rc.Exists() && len(rc.String()) > 0 {
				mapping[rs.Title] = rc.String()
			} else {
				mapping[rs.Title] = rs.Def
			}
		}
	}
	internal.Success(ctx, SendHttpResult{Origin: util.ReplaceLong(er), Mapping: mapping}, e.GetMsg(e.SUCCESS))
}

// RuleExecuteArguments 获取子规则所需执行参数
func RuleExecuteArguments(ctx *gin.Context) {
	code := ctx.Query("code")
	version := ctx.Query("version")
	rd, err := adapter.GetStorage().GetPublishDesignByCode(code, version, util.DEV)
	if err != nil {
		internal.Fail(ctx, nil, err.Error())
	}
	var designInfo util.DesignJson
	err = json.Unmarshal([]byte(rd.DesignJson), &designInfo)
	if err != nil {
		internal.Fail(ctx, nil, err.Error())
	}
	attrs := designInfo.Attrs.(map[string]interface{})
	internal.Success(ctx, attrs["ruleArguments"], e.GetMsg(e.SUCCESS))
}
