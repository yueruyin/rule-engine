package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"reflect"
	"sync"
	"time"
	"zenith.engine.com/engine/pkg/e"
	"zenith.engine.com/engine/pkg/util"
)

var RuleDesignMap = sync.Map{}

// RuleDesign 规则设计主表
type RuleDesign struct {
	Id         uint64 `gorm:"primarykey" json:"id"`
	Code       string `gorm:"type:varchar(255);column:code" json:"code"`
	Name       string `gorm:"type:varchar(255);column:name" json:"name" binding:"required"`
	Desc       string `gorm:"type:varchar(512);column:desc" json:"desc"`
	Design     string `gorm:"type:text;column:design" json:"design"`
	DesignJson string `gorm:"type:longtext;column:design_json" json:"designJson"`
	Version    string `gorm:"type:varchar(255);column:version;default:0" json:"version"`
	//Uri         string    `gorm:"type:varchar(255);column:uri;uniqueIndex:uri:" json:"uri" `
	Uri         string    `gorm:"type:varchar(255);column:uri" json:"uri" `
	StorageType int64     `gorm:"type:tinyint(1);column:storage_type;default:0" json:"storageType"`
	Deleted     int64     `gorm:"type:tinyint(1);column:deleted;default:0"  json:"deleted"`
	Publish     int64     `gorm:"type:tinyint(1);column:publish;default:0"  json:"publish"`
	CreateTime  time.Time `gorm:"column:create_time;default:null" json:"create_time"`
	UpdateTime  time.Time `gorm:"column:update_time;default:null" json:"update_time"`
	GroupId     uint64    `gorm:"type:bigint(20);column:group_id;default:0" json:"group_id"`
	Templated   int64     `gorm:"type:tinyint(1);column:templated;default:0" json:"templated"`
	TemplateId  uint64    `gorm:"type:bigint(20);column:template_id;default:0" json:"template_id"`
	Type        int64     `gorm:"type:tinyint(1);column:type;default:0;comment:'类型 0:规则 1:模版'" json:"type"`
}

// RuleDesignVersion 规则设计版本表 存留临时版本
type RuleDesignVersion struct {
	Id           uint64 `gorm:"primarykey" json:"id"`
	RuleDesignId uint64 `gorm:"type:bigint(20);column:rule_design_id" json:"ruleDesignId"`
	Design       string `gorm:"type:text;column:design" json:"design"`
	DesignJson   string `gorm:"type:longtext;column:design_json" json:"designJson"`
	Version      string `gorm:"type:varchar(255);column:version;default:0" json:"version"`
	//	Uri          string    `gorm:"type:varchar(255);column:uri;uniqueIndex:uri:" json:"uri" `
	Uri        string    `gorm:"type:varchar(255);column:uri" json:"uri" `
	Publish    int64     `gorm:"type:tinyint(1);column:publish;default:0"  json:"publish"`
	CreateTime time.Time `gorm:"column:create_time;default:null" json:"create_time"`
	UpdateTime time.Time `gorm:"column:update_time;default:null" json:"update_time"`
}

func (*RuleDesign) BeforeCreate(tx *gorm.DB) (err error) {
	return
}

// Create 创建子表 创建主表
func (*RuleDesign) Create(req *RuleDesign) error {
	var design RuleDesign
	if design.CheckDesignExist(req.Code, req.Version) {
		return errors.New(e.GetMsg(e.RuleCodeExist))
	}
	if len(req.Uri) == 0 {
		req.Uri = util.GraterUri(req.Version, req.Code)
		UpdateCanvas(req)
	}
	//新增判断uri有的话返回提示
	if design.CheckDesignExistUri(req.Uri, req.Version) {
		return errors.New(e.GetMsg(e.RuleUriError))
	}
	tx := DB.Begin()
	design = RuleDesign{
		Name:        req.Name,
		Code:        req.Code,
		Desc:        req.Desc,
		Design:      req.Design,
		DesignJson:  req.DesignJson,
		Uri:         req.Uri,
		Version:     req.Version,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
		Deleted:     req.Deleted,
		StorageType: req.StorageType,
		Publish:     req.Publish,
		GroupId:     req.GroupId,
		Type:        req.Type,
		Templated:   req.Templated,
		TemplateId:  req.TemplateId,
	}
	if err := tx.Create(&design).Error; err != nil {
		util.Log.Error(e.GetMsg(e.RuleDesignSaveError) + err.Error())
		tx.Rollback()
		return err
	}
	designVersion := RuleDesignVersion{
		RuleDesignId: design.Id,
		Design:       req.Design,
		DesignJson:   req.DesignJson,
		Version:      req.Version,
		Uri:          req.Uri,
		Publish:      req.Publish,
		CreateTime:   time.Now(),
		UpdateTime:   time.Now(),
	}
	if err := tx.Create(&designVersion).Error; err != nil {
		util.Log.Error(e.GetMsg(e.RuleDesignSaveError) + err.Error())
		tx.Rollback()
		return err
	}
	tx.Commit()
	req.Id = design.Id
	return nil
}

func (*RuleDesign) Delete(id uint64) error {
	// (物理删除)
	//err := DB.Where(RuleDesign{Id: id}).Delete(RuleDesign{}).Error
	// (逻辑删除)
	err := DB.Model(RuleDesign{}).Where(RuleDesign{Id: id}).Update("deleted", 1).Error
	return err
}

func (*RuleDesign) DeleteIds(ids []uint64) error {
	// (逻辑删除)
	err := DB.Model(RuleDesign{}).Where("id in ?", ids).Update("deleted", 1).Error
	return err
}

func (*RuleDesign) DeleteByGroupId(groupId uint64) error {
	// (逻辑删除)
	err := DB.Model(RuleDesign{}).Where(RuleDesign{GroupId: groupId}).Updates(RuleDesign{Deleted: 1, GroupId: 0}).Error
	return err
}

// Update 修改主表 修改子表
func (*RuleDesign) Update(req *RuleDesign) error {
	if len(req.Uri) == 0 {
		req.Uri = util.GraterUri(req.Version, req.Code)
		UpdateCanvas(req)
	}
	var design = RuleDesign{
		Id:          req.Id,
		Name:        req.Name,
		Code:        req.Code,
		Version:     req.Version,
		Desc:        req.Desc,
		Design:      req.Design,
		DesignJson:  req.DesignJson,
		Uri:         req.Uri,
		StorageType: req.StorageType,
		Deleted:     req.Deleted,
		Publish:     req.Publish,
		UpdateTime:  time.Now(),
		GroupId:     req.GroupId,
	}
	tx := DB.Begin()
	if err := tx.Updates(&design).Error; err != nil {
		util.Log.Error(e.GetMsg(e.RuleDesignSaveError) + err.Error())
		return err
	}
	var designVersion = RuleDesignVersion{
		Design:     req.Design,
		DesignJson: req.DesignJson,
		Uri:        req.Uri,
		UpdateTime: time.Now(),
		Version:    req.Version,
		Publish:    req.Publish,
	}
	if err := tx.Where("version", req.Version).Where("rule_design_id", req.Id).Updates(&designVersion).Error; err != nil {
		util.Log.Error(e.GetMsg(e.RuleDesignSaveError) + err.Error())
		return err
	}
	tx.Commit()
	return nil
}

// UpdatePublish 修改主表创建子表
func (*RuleDesign) UpdatePublish(req *RuleDesign) error {
	req.Uri = util.GraterUri(req.Version, req.Code)
	UpdateCanvas(req)
	var design = RuleDesign{
		Id:          req.Id,
		Name:        req.Name,
		Code:        req.Code,
		Version:     req.Version,
		Desc:        req.Desc,
		Design:      req.Design,
		DesignJson:  req.DesignJson,
		Uri:         req.Uri,
		StorageType: req.StorageType,
		Deleted:     req.Deleted,
		Publish:     req.Publish,
		UpdateTime:  time.Now(),
		GroupId:     req.GroupId,
	}
	tx := DB.Begin()
	// 修改主表
	if err := tx.Updates(&design).Error; err != nil {
		util.Log.Error(e.GetMsg(e.RuleDesignSaveError) + err.Error())
		return err
	}
	// 修改发布状态
	if req.Publish == 0 {
		tx.Model(RuleDesign{}).Where("id", req.Id).Update("publish", 0)
	}
	designVersion := RuleDesignVersion{
		RuleDesignId: design.Id,
		Design:       req.Design,
		DesignJson:   req.DesignJson,
		Version:      req.Version,
		Uri:          req.Uri,
		Publish:      req.Publish,
		CreateTime:   time.Now(),
		UpdateTime:   time.Now(),
	}
	// 新增子表
	if err := tx.Create(&designVersion).Error; err != nil {
		util.Log.Error(e.GetMsg(e.RuleDesignSaveError) + err.Error())
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (*RuleDesign) UpdateGroupId(groupId uint64) error {
	err := DB.Model(RuleDesign{}).Where(RuleDesign{GroupId: groupId}).Update("group_id", 0).Error
	return err
}

func (*RuleDesign) GetById(id uint64) (RuleDesign, error) {
	var rud RuleDesign
	if DB.Config.Dialector.Name() == "dm" {
		DB.Debug().Raw("SELECT id,`code`,`name`,`desc`,design,design_json,`version`,uri,storage_type,deleted,publish,create_time,update_time,group_id,templated,template_id,`type` FROM rule_design WHERE rule_design.id = ? ORDER BY rule_design.id LIMIT 1", id).Scan(&rud)
	} else {
		DB.Where(RuleDesign{
			Id: id,
		}).First(&rud)
	}
	print(fmt.Sprintf("%v", rud))
	if rud.Id == 0 {
		return rud, errors.New(e.GetMsg(e.RuleNotExist))
	}
	return rud, nil
}

func (*RuleDesign) GetByIds(id []uint64) ([]RuleDesign, error) {
	var rud []RuleDesign
	DB.Where("id in ?", id).Find(&rud)
	return rud, nil
}

func (*RuleDesign) GetByCode(code string, version string) (RuleDesign, error) {
	var rud RuleDesign
	if rud.CheckDesignExist(code, version) {
		if len(version) > 0 {
			DB.Where("code = ? and version = ? and deleted = 0", code, version).First(&rud)
		} else {
			DB.Where("code = ? and deleted = 0", code).Last(&rud)
		}
		if rud.Id == 0 {
			return rud, errors.New(e.GetMsg(e.RuleNotExist))
		}
		return rud, nil
	}
	return rud, errors.New(e.GetMsg(e.RuleCodeNotExist))
}

func (*RuleDesign) GetPublishDesignByCode(code string, version string, env string) (RuleDesign, error) {
	var ruleDesign RuleDesign
	var ruleDesignVersion RuleDesignVersion
	if util.DEV == env {
		if DB.Config.Dialector.Name() == "dm" {
			DB.Debug().Raw("SELECT id,`code`,`name`,`desc`,design,design_json,`version`,uri,storage_type,deleted,publish,create_time,update_time,group_id,templated,template_id,`type` FROM rule_design WHERE rule_design.code = ? and rule_design.deleted = 0 ORDER BY rule_design.id LIMIT 1", code).Scan(&ruleDesign)
		} else {
			DB.Model(RuleDesign{}).Where("code", code).Where("deleted", 0).First(&ruleDesign)
		}
		if ruleDesign.Id == 0 {
			// 没有找到此规则
			return ruleDesign, errors.New(e.GetMsg(e.RuleNotExist))
		}
		if len(version) == 0 {
			// 没有传版本
			if DB.Config.Dialector.Name() == "dm" {
				DB.Debug().Raw("SELECT id,rule_design_id,design,design_json,`version`,uri,publish,create_time,update_time FROM rule_design_version WHERE rule_design_id = ? and publish = 1 ORDER BY id DESC LIMIT 1", ruleDesign.Id).Scan(&ruleDesignVersion)
			} else {
				DB.Model(RuleDesignVersion{}).Where("rule_design_id", ruleDesign.Id).Where("publish", 1).Last(&ruleDesignVersion)
			}
		} else {
			if DB.Config.Dialector.Name() == "dm" {
				DB.Debug().Raw("SELECT id,rule_design_id,design,design_json,`version`,uri,publish,create_time,update_time FROM rule_design_version WHERE rule_design_id = ? and `version` = ?  ORDER BY id LIMIT 1", ruleDesign.Id, version).Scan(&ruleDesignVersion)
			} else {
				DB.Model(RuleDesignVersion{}).Where("rule_design_id", ruleDesign.Id).Where("version", version).First(&ruleDesignVersion)
			}
		}
		if !reflect.DeepEqual(RuleDesignVersion{}, ruleDesignVersion) {
			ruleDesign.Design = ruleDesignVersion.Design
			ruleDesign.DesignJson = ruleDesignVersion.DesignJson
			ruleDesign.Version = ruleDesignVersion.Version
		}
		return ruleDesign, nil
	} else {
		cloumns := ""
		if DB.Config.Dialector.Name() == "dm" {
			cloumns = "rd.id AS id,rd.`code` AS `code`,rd.`name` AS `name`,rd.`desc` AS `desc`,rdv.design AS design,rdv.design_json AS design_json,rdv.version AS version,rdv.uri AS uri,rd.deleted AS deleted,rdv.publish AS publish,rd.create_time AS create_time,rd.update_time AS update_time,rd.group_id AS group_id "
		} else {
			cloumns = "rd.id AS id,rd.code AS code,rd.name AS name,rd.`desc` AS `desc`,rdv.design AS design,rdv.design_json AS design_json,rdv.version AS version,rdv.uri AS uri,rd.deleted AS deleted,rdv.publish AS publish,rd.create_time AS create_time,rd.update_time AS update_time,rd.group_id AS group_id "
		}
		// 正式环境查询逻辑
		DB.Table("rule_design AS rd").
			Select(cloumns).
			//Select("rd.id AS id,rd.`code` AS `code`,rd.`name` AS `name`,rd.`desc` AS `desc`,rdv.design AS design,rdv.design_json AS design_json,rdv.version AS version,rdv.uri AS uri,rd.deleted AS deleted,rdv.publish AS publish,rd.create_time AS create_time,rd.update_time AS update_time,rd.group_id AS group_id ").
			//Select("rd.id AS id,rd.code AS code,rd.name AS name,rd.\"desc\" AS \"desc\",rdv.design AS design,rdv.design_json AS design_json,rdv.version AS version,rdv.uri AS uri,rd.deleted AS deleted,rdv.publish AS publish,rd.create_time AS create_time,rd.update_time AS update_time,rd.group_id AS group_id ").
			Joins("LEFT JOIN rule_design_version AS rdv ON rd.id = rdv.rule_design_id").
			Where("rd.code", code).Where("rdv.publish", 1).Where("rd.deleted", 0).Order("rdv.version desc").Last(&ruleDesign)
		if ruleDesign.Id == 0 {
			// 没有找到此规则或者没有发布规则
			return ruleDesign, errors.New(e.GetMsg(e.RuleNotExistAndPublish))
		}
		return ruleDesign, nil
	}
}

func (*RuleDesign) CheckDesignExist(code string, version string) bool {
	var count int64
	if len(version) > 0 {
		DB.Debug().Model(RuleDesign{}).Where("code = ? and version = ? and deleted = 0", code, version).Count(&count)
	} else {
		DB.Debug().Model(RuleDesign{}).Where("code = ? and deleted = 0", code).Count(&count)
	}
	if count > 0 {
		return true
	}
	return false
}

type Result struct {
	Uri     string
	Version string
}

func (*RuleDesign) GetDataDesign(table string) ([]Result, error) {

	resultArr := make([]Result, 0)
	DB.Table(table).Where("deleted = 0").Select([]string{"uri", "version"}).Scan(&resultArr)

	return resultArr, errors.New(e.GetMsg(e.RuleTableError))
}

// GetByUri 通过Uri的数据查询整条rule
func (*RuleDesign) GetByUri(uri string, version string) (RuleDesign, error) {
	var rud RuleDesign
	var rudVersion RuleDesignVersion
	DB.Model(RuleDesignVersion{}).Where("uri", uri).First(&rudVersion)
	DB.Model(RuleDesign{}).Where("id", rudVersion.RuleDesignId).First(&rud)
	rud.Uri = rudVersion.Uri
	rud.DesignJson = rudVersion.DesignJson
	rud.Design = rudVersion.Design
	rud.Publish = rudVersion.Publish
	rud.Version = rudVersion.Version
	return rud, errors.New(e.GetMsg(e.RuleCodeNotExist))
}

func (*RuleDesign) ListDesign(page int, pageSize int, name string, code string, groupId []string, types int) (error, []RuleDesign, int64) {
	ruleDesignList := make([]RuleDesign, 0)
	var count int64
	tx := DB.Model(RuleDesign{}).Where("deleted =?", 0)
	tc := DB.Model(RuleDesign{}).Where("deleted =?", 0)
	if len(name) > 0 {
		tx.Where("name LIKE ?", "%"+name+"%")
		tc.Where("name LIKE ?", "%"+name+"%")
	}
	if len(code) > 0 {
		tx.Where("code LIKE ?", "%"+code+"%")
		tc.Where("code LIKE ?", "%"+code+"%")
	}
	if len(groupId) > 0 && !util.InSlice(groupId, "0") {
		tx.Where("group_id in ?", groupId)
		tc.Where("group_id in ?", groupId)
	}
	tx.Where("type = ?", types)
	tc.Where("type = ?", types)
	if DB.Config.Dialector.Name() == "dm" {
		if err := tx.Table("rule_design as rd").Select("MAX( rd.id ) AS id,rd.`code`,MAX( rd.`name` ) AS `name`,MAX( rd.`desc` ) AS `desc`,MAX( rdv.`version` ) AS `version`,rd.type,rd.templated,MAX( rdv.design ) AS design,MAX( rdv.design_json ) AS design_json,MAX( rdv.create_time ) AS create_time,MAX( rdv.update_time ) AS update_time,MAX( rdv.uri ) AS uri,MAX( deleted ) AS deleted,rd.publish AS publish,MAX( group_id ) AS group_id").
			Joins("LEFT JOIN rule_design_version AS rdv ON rd.id = rdv.rule_design_id ").Group("rd.code").Order("rd.id desc").Limit(pageSize).Offset((page - 1) * pageSize).Find(&ruleDesignList).Error; err != nil {
			return err, nil, 0
		}
	} else {
		if err := tx.Table("rule_design as rd").Select("MAX( rd.id ) AS id,rd.code,MAX( rd.name ) AS name,MAX( rd.desc ) AS \"desc\",MAX( rdv.version ) AS version,rd.type,rd.templated,MAX( rdv.design ) AS design,MAX( rdv.design_json ) AS design_json,MAX( rdv.create_time ) AS create_time,MAX( rdv.update_time ) AS update_time,MAX( rdv.uri ) AS uri,MAX( deleted ) AS deleted,rd.publish AS publish,MAX( group_id ) AS group_id").
			Joins("LEFT JOIN rule_design_version AS rdv ON rd.id = rdv.rule_design_id ").Group("rd.code").Order("rd.id desc").Limit(pageSize).Offset((page - 1) * pageSize).Find(&ruleDesignList).Error; err != nil {
			return err, nil, 0
		}
	}
	tc.Model(RuleDesign{}).Group("code").Count(&count)
	//countInt := strconv.FormatInt(count, 10)
	//id16, _ := strconv.Atoi(countInt)
	//pageNum := id16 / pageSize
	//if id16%pageSize != 0 {
	//	pageNum++
	//}
	//pageNum64, _ := strconv.ParseInt(strconv.Itoa(pageNum), 10, 64)
	return nil, ruleDesignList, count
}

func (*RuleDesign) CheckDesignExistUri(uri string, version string) bool {
	var count int64
	// 这里不清楚为什么必须要加model()?
	DB.Debug().Model(RuleDesign{}).Where("uri = ? and version = ? and deleted = 0", uri, version).Count(&count)
	if count > 0 {
		return true
	}
	return false
}

func (*RuleDesign) PublishDesign(id uint64) bool {
	DB.Model(RuleDesign{}).Where(RuleDesign{Id: id}).Update("publish", 1)
	return true
}

func (*RuleDesign) ListByCode(code string) ([]string, error) {
	ruleDesignVersionList := make([]string, 0)
	DB.Table("rule_design rd").
		Select("rdv.version as version").
		Joins("left join rule_design_version rdv on rd.id= rdv.rule_design_id").
		Where("code", code).
		Order("rdv.version desc").
		Find(&ruleDesignVersionList)
	return ruleDesignVersionList, nil
}

func (*RuleDesign) ListByGroupId(groupId uint64, types int) []RuleDesign {
	ruleDesignList := make([]RuleDesign, 0)
	if DB.Config.Dialector.Name() == "dm" {
		sql := "SELECT id,`code`,`name`,`desc`,design,design_json,`version`,uri,storage_type,deleted,publish,create_time,update_time,group_id,templated,template_id,`type` FROM rule_design  WHERE `group_id` = ? AND `deleted` = 0 AND `type` = ?"
		DB.Debug().Raw(sql, groupId, types).Scan(&ruleDesignList)
	} else {
		DB.Table("rule_design rd").Where("group_id", groupId).Where("deleted", 0).Where("type", types).Find(&ruleDesignList)
	}
	return ruleDesignList
}

// UpdateCanvas 修改设计画布信息
func UpdateCanvas(rud *RuleDesign) {
	var designInfo *util.DesignJson
	json.Unmarshal([]byte(rud.DesignJson), &designInfo)
	attrs := designInfo.Attrs.(map[string]interface{})
	attrs["ruleVersion"] = rud.Version
	attrs["ruleName"] = rud.Name
	attrs["ruleDesc"] = rud.Desc
	attrs["ruleCode"] = rud.Code
	attrs["ruleURI"] = util.GraterUri(rud.Version, rud.Code)
	rud.Uri = util.GraterUri(rud.Version, rud.Code)
	designInfo.Attrs = attrs
	newDesignJson, _ := json.Marshal(designInfo)
	rud.DesignJson = string(newDesignJson)
}

// GetDesignByCodeAndVersion 通过版本和编码获取设计信息
func (*RuleDesign) GetDesignByCodeAndVersion(code string, version string) RuleDesign {
	var ruleDesign RuleDesign
	cloumns := ""
	if DB.Config.Dialector.Name() == "dm" {
		cloumns = "rd.type, rd.desc, rdv.design_json, rd.id ,rd.code ,rd.name, rdv.version, rdv.uri, rdv.publish ,rd.deleted ,rd.group_id"
	} else {
		cloumns = "rd.type, rd.\"desc\", rdv.design_json, rd.id ,rd.code ,rd.name, rdv.version, rdv.uri, rdv.publish ,rd.deleted ,rd.group_id"
	}
	DB.Table("rule_design_version rdv").
		Select(cloumns).
		Joins("left join rule_design rd on rdv.rule_design_id = rd.id").
		Where("rd.code", code).Where("rdv.version", version).
		Where("deleted", 0).
		Order("rdv.id").Take(&ruleDesign)
	return ruleDesign
}

func (*RuleDesign) GetLastVersionByCode(code string) string {
	var lastVersion string
	DB.Table("rule_design rd").Select("rdv.version").
		Joins("left join rule_design_version rdv on rdv.rule_design_id = rd.id").
		Where("rd.code", code).Where("deleted", 0).Order("rdv.id desc").Limit(1).Scan(&lastVersion)
	return lastVersion
}

func (*RuleDesign) CheckCodeExist(code string, id string) bool {
	var count int64
	if len(id) > 0 {
		DB.Debug().Model(RuleDesign{}).Where("code = ? and id != ? and deleted = 0", code, id).Count(&count)
	} else {
		DB.Debug().Model(RuleDesign{}).Where("code = ? and deleted = 0", code).Count(&count)
	}
	if count > 0 {
		return true
	}
	return false
}
