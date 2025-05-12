package handler

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
	"zenith.engine.com/engine/internal"
	"zenith.engine.com/engine/internal/adapter"
	"zenith.engine.com/engine/internal/repository"
	"zenith.engine.com/engine/pkg/e"
	"zenith.engine.com/engine/pkg/util"
)

type RuleGroupResult struct {
	Id         uint64             `json:"id"`
	Code       string             `json:"code"`
	Name       string             `json:"name"`
	ParentId   uint64             `json:"parentId"`
	CreateTime time.Time          `json:"create_time"`
	UpdateTime time.Time          `json:"update_time"`
	Deleted    int64              `json:"deleted"`
	Children   []*RuleGroupResult `json:"children"`
}

type RuleGroupRequest struct {
	Id         uint64    `json:"id"`
	Code       string    `json:"code"`
	Name       string    `json:"name" binding:"required"`
	ParentId   uint64    `json:"parentId"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
	Deleted    int64     `json:"deleted"`
	RoleIds    []uint64  `json:"roleIds"`
	Type       int64     `json:"type"`
}

// SaveGroup 添加分组
func SaveGroup(ctx *gin.Context) {
	var request *RuleGroupRequest
	err := ctx.BindJSON(&request)
	if err != nil {
		internal.Fail(ctx, nil, err.Error())
		return
	}
	rug := repository.RuleGroup{}
	util.SimpleCopyProperties(&rug, request)
	err = rug.CreateGroup(&rug)
	if err != nil {
		internal.Fail(ctx, nil, err.Error())
		return
	}
	// 分组角色权限
	if len(request.RoleIds) > 0 {
		var roleGroup repository.RoleGroup
		for _, id := range request.RoleIds {
			roleGroup.CreateRoleGroup(id, rug.Id, "RW", 0)
		}
	}
	internal.Success(ctx, nil, e.GetMsg(e.SUCCESS))
}

// GetGroupInfo 获取分组信息
func GetGroupInfo(ctx *gin.Context) {
	var rg repository.RuleGroup
	id := ctx.Query("id")
	internal.Success(ctx, rg.GetGroupInfo(id), e.GetMsg(e.SUCCESS))
}

// ListGroup 分组列表查询
func ListGroup(ctx *gin.Context) {
	var ruleGroupResult []*RuleGroupResult
	var ruleGroupAll []*RuleGroupResult
	var rg repository.RuleGroup
	// 获取当前用户信息
	userinfo := util.CurrentUser(ctx)
	name := ctx.Query("name")
	parentId := ctx.Query("parentId")
	var pid uint64
	pid, _ = strconv.ParseUint(parentId, 10, 64)
	roleGroups := repository.GetRoleGroupByRoleId(userinfo.RoleId)
	rgList := rg.ListGroup(pid, name, 0)
	//过滤掉无权限规则分组
	rgList = filterGroup(roleGroups, rgList)
	for _, group := range rgList {
		ruleGroupAll = append(ruleGroupAll,
			&RuleGroupResult{
				Id:         group.Id,
				Code:       group.Code,
				Name:       group.Name,
				CreateTime: group.CreateTime,
				UpdateTime: group.UpdateTime,
				ParentId:   group.ParentId,
				Deleted:    group.Deleted,
				Children:   getTreeRecursive(ruleGroupAll, group.Id),
			})
	}
	for _, result := range ruleGroupAll {
		isShow := true
		// 处理2级规则组
		for _, groupResult := range ruleGroupAll {
			if result.ParentId == groupResult.Id {
				isShow = false
				break
			}
		}
		if result.ParentId == 0 || isShow {
			ruleGroupResult = append(ruleGroupResult, result)
		}

	}
	internal.Success(ctx, ruleGroupResult, e.GetMsg(e.SUCCESS))
}

// filterGroup 过滤掉没有权限的分组
func filterGroup(groups []repository.RoleGroup, rg []repository.RuleGroup) []repository.RuleGroup {
	var frg []repository.RuleGroup
	isManager := false
	for _, result := range rg {
		for _, group := range groups {
			if group.GroupId == 0 {
				isManager = true
				break
			}
			if result.Id == group.GroupId {
				frg = append(frg, result)
				break
			}
		}
		if isManager {
			break
		}
	}
	if isManager {
		return rg
	}
	return frg
}

// UpdateGroup 修改分组信息
func UpdateGroup(ctx *gin.Context) {
	var request *RuleGroupRequest
	err := ctx.BindJSON(&request)
	if err != nil {
		internal.Fail(ctx, nil, err.Error())
		return
	}
	rug := repository.RuleGroup{}
	util.SimpleCopyProperties(&rug, request)
	err = rug.UpdateGroup(&rug)
	if err != nil {
		internal.Fail(ctx, err, e.GetMsg(e.ERROR))
		return
	}
	var roleGroup repository.RoleGroup
	repository.DeleteRoleGroupByGroupId(request.Id)
	for _, id := range request.RoleIds {
		roleGroup.CreateRoleGroup(id, request.Id, "RW", 0)
	}
	internal.Success(ctx, nil, e.GetMsg(e.SUCCESS))
}

type DeleteGroupParam struct {
	Id        uint64 `json:"id" binding:"required"`
	IsContent bool   `json:"isContent"`
}

// DeleteGroup 删除分组
func DeleteGroup(ctx *gin.Context) {
	var rg repository.RuleGroup
	var dp DeleteGroupParam
	err := ctx.BindJSON(&dp)
	if err != nil {
		internal.Fail(ctx, nil, err.Error())
		return
	}
	repository.DeleteRoleGroupByGroupId(dp.Id)
	// 处理子级分组和规则
	FDeleteByParentId(dp.Id, dp.IsContent, rg)
	// 删除当前分组
	err = rg.DeleteGroup(dp.Id)
	if dp.IsContent {
		// 通过分组删除规则
		adapter.GetStorage().DeleteByGroupId(dp.Id)
	} else {
		// 将规则放在最外层
		adapter.GetStorage().UpdateGroupId(dp.Id)
	}
	if err != nil {
		internal.Fail(ctx, nil, e.GetMsg(e.ERROR))
		return
	}
	internal.Success(ctx, nil, e.GetMsg(e.SUCCESS))
}

// FDeleteByParentId 删除分组
func FDeleteByParentId(uid uint64, rb bool, rg repository.RuleGroup) {
	rgList := rg.ListGroup(uid, "", 0)
	for _, r := range rgList {
		rg.DeleteByParentId(uid)
		FDeleteByParentId(r.Id, rb, rg)
		if rb {
			// 通过分组删除规则
			adapter.GetStorage().DeleteByGroupId(r.Id)
		} else {
			// 将规则放在最外层
			adapter.GetStorage().UpdateGroupId(r.Id)
		}
	}
}

// getTreeRecursive 获取数据子节点
func getTreeRecursive(list []*RuleGroupResult, parentId uint64) []*RuleGroupResult {
	res := make([]*RuleGroupResult, 0)
	for _, v := range list {
		if v.ParentId == parentId {
			v.Children = getTreeRecursive(list, v.Id)
			res = append(res, v)
		}
	}
	return res
}
