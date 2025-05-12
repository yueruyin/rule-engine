package repository

import "time"

type RoleGroup struct {
	Id         uint64    `gorm:"primarykey" json:"id"` // 自增
	RoleId     uint64    `gorm:"type:bigint(20);column:role_id;comment:'角色id'" json:"role_id"`
	GroupId    uint64    `gorm:"type:bigint(20);column:group_id;comment:'规则分组id'" json:"group_id"`
	Action     string    `gorm:"type:varchar(255);column:action;comment:'动作(R:读 W:写 RW:读写)'" json:"action"`
	CreateTime time.Time `gorm:"column:create_time;default:null;comment:'新增时间'" json:"create_time"`
	UpdateTime time.Time `gorm:"column:update_time;default:null;comment:'修改时间'" json:"update_time"`
	Deleted    int64     `gorm:"type:tinyint(1);column:deleted;default:0;comment:'是否删除'"  json:"deleted"`
	OrgId      uint64    `gorm:"type:bigint(20);column:org_id;comment:'组织id'" json:"org_id"`
}

// CreateRoleGroup 创建权限
func (*RoleGroup) CreateRoleGroup(roleId uint64, groupId uint64, action string, orgId uint64) (RoleGroup, error) {
	roleGroup := RoleGroup{RoleId: roleId, GroupId: groupId, Action: action, OrgId: orgId, CreateTime: time.Now(), UpdateTime: time.Now()}
	tx := DB.Begin()
	tx.Create(&roleGroup)
	tx.Commit()
	return roleGroup, nil
}

// GetRoleGroupByRoleId 通过角色id获取权限
func GetRoleGroupByRoleId(roleId uint64) []RoleGroup {
	roleGroupList := make([]RoleGroup, 0)
	if DB.Config.Dialector.Name() == "dm" {
		DB.Debug().Raw(
			"SELECT id,role_id,group_id,`action`,create_time,update_time,deleted,org_id FROM role_group WHERE role_id = ? AND deleted = 0",
			roleId).Scan(&roleGroupList)
	} else {
		DB.Model(RoleGroup{}).Where("role_id", roleId).Where("deleted", 0).Find(&roleGroupList)
	}
	return roleGroupList
}

// GetRoleGroupByGroupId 通过分组id获取权限
func GetRoleGroupByGroupId(groupId uint64) []RoleGroup {
	roleGroupList := make([]RoleGroup, 0)
	if DB.Config.Dialector.Name() == "dm" {
		DB.Debug().Raw(
			"SELECT id,role_id,group_id,`action`,create_time,update_time,deleted,org_id FROM role_group WHERE group_id = ? AND deleted = 0",
			groupId).Scan(&roleGroupList)
	} else {
		DB.Model(RoleGroup{}).Where("group_id", groupId).Where("deleted", 0).Find(&roleGroupList)
	}
	return roleGroupList
}

// DeleteRoleGroupByGroupId 通过分组id删除权限
func DeleteRoleGroupByGroupId(groupId uint64) {
	rgs := GetRoleGroupByGroupId(groupId)
	for _, rg := range rgs {
		DB.Model(RoleGroup{}).Delete(&rg)
	}

}
