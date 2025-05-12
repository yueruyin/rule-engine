package file

import (
	"time"
	"zenith.engine.com/engine/internal/repository"
)

type FileRuleDesign struct {
	Id         uint64    `json:"id"`
	Code       string    `json:"code"`
	Name       string    `json:"name" binding:"required"`
	Desc       string    `json:"desc"`
	Design     string    `json:"design" binding:"required"`
	Version    string    `json:"version"`
	Uri        string    `json:"uri" binding:"required"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

func (*FileRuleDesign) Create(req *repository.RuleDesign) error {
	return nil
}

func (*FileRuleDesign) Delete(id uint64) error {
	return nil
}

func (*FileRuleDesign) DeleteIds(id []uint64) error {
	return nil
}

func (*FileRuleDesign) GetById(id uint64) (repository.RuleDesign, error) {
	return repository.RuleDesign{}, nil
}

func (*FileRuleDesign) GetByIds(id []uint64) ([]repository.RuleDesign, error) {
	return []repository.RuleDesign{}, nil
}

func (*FileRuleDesign) Update(req *repository.RuleDesign) error {
	return nil
}

func (*FileRuleDesign) UpdatePublish(req *repository.RuleDesign) error {

	return nil
}

func (*FileRuleDesign) GetByCode(code string, version string) (repository.RuleDesign, error) {
	return repository.RuleDesign{}, nil
}

func (*FileRuleDesign) CheckDesignExist(code string, version string) bool {
	return false
}

func (*FileRuleDesign) ListDesign(page int, pageSize int, name string, code string, groupId []string, types int) (error, []repository.RuleDesign, int64) {
	return nil, nil, 0
}

// 通过表名获取字段

func (*FileRuleDesign) GetDataDesign(table string) ([]repository.Result, error) {
	return []repository.Result{}, nil
}
func (*FileRuleDesign) GetByUri(uri string, version string) (repository.RuleDesign, error) {
	return repository.RuleDesign{}, nil
}
func (*FileRuleDesign) CheckDesignExistUri(uri string, version string) bool {
	return false
}

func (*FileRuleDesign) PublishDesign(id uint64) bool {

	return true
}

func (*FileRuleDesign) DeleteByGroupId(groupId uint64) error {

	return nil
}

func (*FileRuleDesign) UpdateGroupId(groupId uint64) error {
	return nil
}

func (*FileRuleDesign) ListByCode(code string) ([]string, error) {

	return nil, nil
}

func (*FileRuleDesign) GetPublishDesignByCode(code string, version string, env string) (repository.RuleDesign, error) {

	return repository.RuleDesign{}, nil
}

// GetDesignByCodeAndVersion 通过版本和编码获取设计信息
func (*FileRuleDesign) GetDesignByCodeAndVersion(code string, version string) repository.RuleDesign {

	return repository.RuleDesign{}
}

func (*FileRuleDesign) ListByGroupId(groupId uint64, types int) []repository.RuleDesign {
	return nil
}

func (*FileRuleDesign) GetLastVersionByCode(code string) string {
	return ""
}

func (*FileRuleDesign) CheckCodeExist(code string, id string) bool {
	return true
}
