package repository

import (
	"time"
)

type Role struct {
	Id         uint64    `gorm:"primarykey" json:"id"`                          // 自增
	RoleName   string    `gorm:"column:rolename;comment:'角色名'" json:"rolename"` // 唯一
	CreateTime time.Time `gorm:"column:create_time;default:null;comment:'新增时间'" json:"create_time"`
	UpdateTime time.Time `gorm:"column:update_time;default:null;comment:'修改时间'" json:"update_time"`
	Enabled    int64     `gorm:"type:tinyint(1);column:enabled;default:0;comment:'是否禁用'"  json:"enabled"`
	Manager    int64     `gorm:"type:tinyint(1);column:manager;default:0;comment:'是否管理者'"  json:"manager"`
	Deleted    int64     `gorm:"type:tinyint(1);column:deleted;default:0;comment:'是否删除'"  json:"deleted"`
	OrgId      uint64    `gorm:"type:bigint(20);column:org_id;comment:'组织id'" json:"org_id"`
}

// CreateRole 创建角色
func (*Role) CreateRole(roleName string, manager int64, orgId uint64) (Role, error) {
	role := Role{RoleName: roleName, CreateTime: time.Now(), UpdateTime: time.Now(), Manager: manager, OrgId: orgId}
	tx := DB.Begin()
	tx.Create(&role)
	tx.Commit()
	return role, nil
}

// ExistsRole 验证角色是否存在
func (*Role) ExistsRole(roleName string) bool {
	var count int64
	DB.Model(Role{}).Where("rolename", roleName).Count(&count)
	return count > 0
}

// GetRoleByName 获取角色信息
func (*Role) GetRoleByName(roleName string) Role {
	var role Role
	DB.Model(Role{}).Where("rolename", roleName).First(&role)
	return role
}

// GetRoleById 通过id获取角色信息
func (*Role) GetRoleById(roleId uint64) (role Role) {
	if DB.Config.Dialector.Name() == "dm" {
		DB.Debug().Raw("SELECT id,rolename,create_time,update_time,enabled,manager,deleted,org_id FROM `role` WHERE id = ? ORDER BY `role`.id LIMIT 1", roleId).Scan(&role)
	} else {
		DB.Model(Role{}).Where("id", roleId).First(&role)
	}
	return role
}

// ListRole 获取全部角色信息
func (*Role) ListRole() []Role {
	var roles []Role
	DB.Model(Role{}).Find(&roles)
	return roles
}
