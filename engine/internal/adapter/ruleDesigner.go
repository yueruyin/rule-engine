package adapter

import (
	"zenith.engine.com/engine/config"
	"zenith.engine.com/engine/internal/file"
	"zenith.engine.com/engine/internal/repository"
)

type RuleDesigner interface {
	Create(req *repository.RuleDesign) error //new rule design...
	Delete(id uint64) error
	DeleteIds(ids []uint64) error
	DeleteByGroupId(groupId uint64) error    //designed according to ID deletion rules
	Update(req *repository.RuleDesign) error //update rule design...

	UpdatePublish(req *repository.RuleDesign) error // update rule publish design

	UpdateGroupId(groupId uint64) error
	ListDesign(page int, pageSize int, name string, code string, groupId []string, types int) (error, []repository.RuleDesign, int64) // list rule design...

	GetById(id uint64) (repository.RuleDesign, error)
	GetByIds(id []uint64) ([]repository.RuleDesign, error)                                         //get rule design through ID
	GetByCode(code string, version string) (repository.RuleDesign, error)                          //get rule design through code
	GetPublishDesignByCode(code string, version string, env string) (repository.RuleDesign, error) // 获取最新已发布版本设计信息
	ListByCode(code string) ([]string, error)
	ListByGroupId(groupId uint64, types int) []repository.RuleDesign
	GetDesignByCodeAndVersion(code string, version string) repository.RuleDesign

	GetDataDesign(table string) ([]repository.Result, error)
	GetByUri(uri string, version string) (repository.RuleDesign, error)

	CheckDesignExistUri(uri string, version string) bool //verify whether the rule uri exist
	CheckDesignExist(code string, version string) bool   //verify whether the rule code and version exist
	CheckCodeExist(code string, id string) bool

	PublishDesign(id uint64) bool // publish rule design
	GetLastVersionByCode(code string) string
}

func GetStorage() RuleDesigner {
	var ruleDesigner RuleDesigner
	switch config.Conf.Rule.Storage {
	case "file":
		ruleDesigner = &file.FileRuleDesign{}
	case "db":
		ruleDesigner = &repository.RuleDesign{}
	default:
		ruleDesigner = &repository.RuleDesign{}
	}
	return ruleDesigner
}
