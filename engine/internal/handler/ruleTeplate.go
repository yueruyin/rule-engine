package handler

import (
	"github.com/gin-gonic/gin"
	"time"
	"zenith.engine.com/engine/internal"
	"zenith.engine.com/engine/internal/repository"
	"zenith.engine.com/engine/pkg/e"
)

type TemplateTagResult struct {
	Id         uint64    `json:"id"`
	Code       string    `json:"code"`
	Name       string    `json:"name"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
	Deleted    int64     `json:"deleted"`
}

type RuleTemplateResult struct {
	Id       uint64                `json:"id"`
	Name     string                `json:"name"`
	ParentId uint64                `json:"parentId"`
	Type     int64                 `json:"type"` //1 标签 2.模版
	Children []*RuleTemplateResult `json:"children"`
}

// ListTemplateTag 获取全部标签
func ListTemplateTag(ctx *gin.Context) {
	var templateTagResult []*TemplateTagResult
	var rg repository.RuleGroup
	name := ctx.Query("name")
	rgList := rg.ListGroup(0, name, 1)
	for _, tag := range rgList {
		templateTagResult = append(templateTagResult, &TemplateTagResult{
			Id:         tag.Id,
			Code:       tag.Code,
			Name:       tag.Name,
			CreateTime: tag.CreateTime,
			UpdateTime: tag.UpdateTime,
			Deleted:    tag.Deleted,
		})
	}
	internal.Success(ctx, templateTagResult, e.GetMsg(e.SUCCESS))
}

// TreeTemplate 获取树形模版
func TreeTemplate(ctx *gin.Context) {
	var ruleTemplateResult []*RuleTemplateResult
	var rt repository.RuleGroup
	var ruleTemplate repository.RuleDesign
	name := ctx.Query("name")
	// 查询全部标签
	rtList := rt.ListGroup(0, name, 1)
	for _, tag := range rtList {
		ruleTemplateResult = append(ruleTemplateResult, &RuleTemplateResult{
			Id:       tag.Id,
			Name:     tag.Name,
			Type:     0,
			Children: GetTemplateList(tag.Id),
		})
	}
	rdList := ruleTemplate.ListByGroupId(0, 1)
	for _, design := range rdList {
		ruleTemplateResult = append(ruleTemplateResult, &RuleTemplateResult{
			Id:   design.Id,
			Name: design.Name,
			Type: 1,
		})
	}
	internal.Success(ctx, ruleTemplateResult, e.GetMsg(e.SUCCESS))
}

func GetTemplateList(tagId uint64) []*RuleTemplateResult {
	var ruleTemplateResult []*RuleTemplateResult
	var ruleTemplate repository.RuleDesign
	rgList := ruleTemplate.ListByGroupId(tagId, 1)
	for _, design := range rgList {
		ruleTemplateResult = append(ruleTemplateResult, &RuleTemplateResult{
			Id:       design.Id,
			Name:     design.Name,
			Type:     1,
			ParentId: tagId,
		})
	}
	return ruleTemplateResult
}
