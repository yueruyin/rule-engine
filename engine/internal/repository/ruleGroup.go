package repository

import (
	"strconv"
	"time"
	"zenith.engine.com/engine/pkg/e"
	"zenith.engine.com/engine/pkg/util"
)

type RuleGroup struct {
	Id         uint64    `gorm:"primarykey" json:"id"`
	Code       string    `gorm:"type:varchar(255);column:code;comment:'编码'" json:"code"`
	Name       string    `gorm:"type:varchar(255);column:name;comment:'分组名称'" json:"name" binding:"required"`
	ParentId   uint64    `gorm:"type:bigint(20);column:parent_id;default:0;comment:'上级id'" json:"parentId"`
	CreateTime time.Time `gorm:"column:create_time;default:null;comment:'新增时间'" json:"create_time"`
	UpdateTime time.Time `gorm:"column:update_time;default:null;comment:'修改时间'" json:"update_time"`
	Deleted    int64     `gorm:"type:tinyint(1);column:deleted;default:0;comment:'是否删除'"  json:"deleted"`
	Type       int64     `gorm:"type:tinyint(1);column:type;default:0;comment:'类型 0:分组 1:标签'" json:"type"`
}

func (*RuleGroup) CreateGroup(req *RuleGroup) error {
	var group RuleGroup
	tx := DB.Begin()
	group = RuleGroup{
		Name:       req.Name,
		Code:       req.Code,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
		ParentId:   req.ParentId,
		Deleted:    0,
		Type:       req.Type,
	}
	if err := tx.Create(&group).Error; err != nil {
		util.Log.Error(e.GetMsg(e.RuleGroupSaveError) + err.Error())
		tx.Rollback()
		return err
	}
	tx.Commit()
	req.Id = group.Id
	return nil
}

func (*RuleGroup) UpdateGroup(req *RuleGroup) error {
	err := DB.Model(RuleGroup{}).Where(RuleGroup{Id: req.Id}).
		UpdateColumns(req).Error
	if req.ParentId == 0 {
		DB.Model(RuleGroup{}).Where(RuleGroup{Id: req.Id}).
			Update("parent_id", 0)
	}
	return err
}

func (*RuleGroup) ListGroup(pid uint64, name string, types int64) []RuleGroup {
	ruleGroupList := make([]RuleGroup, 0)
	if DB.Config.Dialector.Name() == "dm" {
		sql := "SELECT id,code,`name`,parent_id,create_time,update_time,deleted,`type`  FROM rule_group WHERE deleted =0 AND `type` =?"
		if len(name) > 0 {
			sql += " AND name LIKE '%" + name + "%'"
		}
		if pid > 0 {
			sql += " AND parent_id =" + strconv.FormatUint(pid, 10) + ""
		}
		sql += " ORDER BY id DESC"
		DB.Debug().Raw(sql, types).Scan(&ruleGroupList)
	} else {
		tx := DB.Model(RuleGroup{}).Where("deleted =?", 0).Where("type =?", types)
		if len(name) > 0 {
			tx.Where("name LIKE ?", "%"+name+"%")
		}
		if pid > 0 {
			tx.Where("parent_id =?", pid)
		}
		if err := tx.Model(RuleGroup{}).Order("id desc").Find(&ruleGroupList).Error; err != nil {
			return nil
		}
	}
	return ruleGroupList
}

func (*RuleGroup) ParentGroup(pid uint64) []RuleGroup {
	ruleGroupList := make([]RuleGroup, 0)
	if DB.Config.Dialector.Name() == "dm" {
		sql := "SELECT id,code,`name`,parent_id,create_time,update_time,deleted,`type` FROM `rule_group` WHERE deleted =0 AND parent_id=? ORDER BY id desc"
		DB.Debug().Raw(sql, pid).Scan(&ruleGroupList)
	} else {
		if err := DB.Model(RuleGroup{}).Where("deleted =?", 0).Where("parent_id=?", pid).Order("id desc").Find(&ruleGroupList).Error; err != nil {
			return nil
		}
	}
	return ruleGroupList
}

func (*RuleGroup) GetGroupInfo(id string) *RuleGroup {
	var ruleGroup *RuleGroup
	tx := DB.Model(RuleGroup{}).Where("id", id)
	tx.First(&ruleGroup)
	return ruleGroup
}

func (*RuleGroup) DeleteGroup(id uint64) error {
	err := DB.Model(RuleGroup{}).Delete(RuleGroup{Id: id}).Error
	return err
}

func (*RuleGroup) DeleteByParentId(parentId uint64) error {
	var rg RuleGroup
	err := DB.Model(RuleGroup{}).Where(RuleGroup{ParentId: parentId}).Delete(&rg).Error
	return err
}
