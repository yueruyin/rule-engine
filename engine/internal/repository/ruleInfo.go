package repository

import (
	"errors"
	"time"
	"zenith.engine.com/engine/pkg/e"
	"zenith.engine.com/engine/pkg/util"
)

type RuleInfo struct {
	Id         uint64    `gorm:"primarykey" json:"id"`
	Code       string    `gorm:"unique" binding:"required" json:"code"`
	Name       string    `gorm:"type:varchar(255);column:name" json:"name" binding:"required"`
	Desc       string    `gorm:"type:varchar(512);column:desc" json:"desc"`
	Script     string    `gorm:"type:text;column:script" json:"script" binding:"required"`
	CreateTime time.Time `gorm:"column:create_time;default:null" json:"create_time"`
	UpdateTime time.Time `gorm:"column:update_time;default:null" json:"update_time"`
}

func FindRuleInfoList() []RuleInfo {
	var ruleInfoList []RuleInfo
	DB.Find(&ruleInfoList)
	return ruleInfoList
}

func (*RuleInfo) Create(req *RuleInfo) error {
	var info RuleInfo
	if CheckRuleInfoExist(req.Code) {
		return errors.New(e.GetMsg(e.RuleCodeExist))
	}
	info = RuleInfo{
		Name:       req.Name,
		Code:       req.Code,
		Desc:       req.Desc,
		Script:     req.Script,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	if err := DB.Create(&info).Error; err != nil {
		util.Log.Error(e.GetMsg(e.RuleInfoSaveError) + err.Error())
		return err
	}
	return nil
}

func (*RuleInfo) Update(req *RuleInfo) error {
	var design RuleInfo
	design = RuleInfo{
		Id:         req.Id,
		Name:       req.Name,
		Code:       req.Code,
		Desc:       req.Desc,
		Script:     req.Script,
		UpdateTime: time.Now(),
	}
	if err := DB.Updates(&design).Error; err != nil {
		util.Log.Error(e.GetMsg(e.RuleInfoSaveError) + err.Error())
		return err
	}
	return nil
}

func (*RuleInfo) GetById(id uint64) RuleInfo {
	var info RuleInfo
	DB.Where(RuleInfo{Id: id}).First(&info)
	return info
}

func (*RuleInfo) GetByCode(code string) RuleInfo {
	var info RuleInfo
	DB.Where(RuleInfo{Code: code}).First(&info)
	return info
}

func (*RuleInfo) Delete(id uint64) error {
	err := DB.Where(RuleInfo{Id: id}).Delete(RuleInfo{}).Error
	return err
}

func (*RuleInfo) DeleteByCode(code string) error {
	err := DB.Where(RuleInfo{Code: code}).Delete(RuleInfo{}).Error
	return err
}

func CheckRuleInfoExist(code string) bool {
	var count int64
	DB.Model(RuleInfo{}).Where(RuleInfo{Code: code}).Count(&count)
	if count > 0 {
		return true
	}
	return false
}
