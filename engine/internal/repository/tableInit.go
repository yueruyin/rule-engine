package repository

import (
	"os"
	"zenith.engine.com/engine/pkg/util"
)

func migration() {
	//自动迁移模式
	err := DB.Set("gorm:rule_design_version", "charset=utf8mb4").AutoMigrate(&RuleDesignVersion{})
	err = DB.Set("gorm:rule_design", "charset=utf8mb4").AutoMigrate(&RuleDesign{})
	err = DB.Set("gorm:rule_info", "charset=utf8mb4").AutoMigrate(&RuleInfo{})
	err = DB.Set("gorm:rule_group", "charset=utf8mb4").AutoMigrate(&RuleGroup{})
	err = DB.Set("gorm:user", "charset=utf8mb4").AutoMigrate(&User{})
	err = DB.Set("gorm:role", "charset=utf8mb4").AutoMigrate(&Role{})
	err = DB.Set("gorm:role_group", "charset=utf8mb4").AutoMigrate(&RoleGroup{})
	if err != nil {
		util.Log.Infoln("register init table fail")
		os.Exit(0)
	}

	// 创建默认用户
	user := User{UserName: "rule", Password: "rule"}
	if !user.ExistsUser(user.UserName) {
		user, err = user.CreateUser(user.UserName, user.Password, user.OrgId)
		if err != nil {
			util.Log.Infoln("init user fail")
			os.Exit(0)
		}
	}

	// 创建默认角色
	role := Role{RoleName: "ROLE_ADMIN", Manager: 1}
	if !role.ExistsRole(role.RoleName) {
		role, err = role.CreateRole(role.RoleName, role.Manager, role.OrgId)
		if err != nil {
			util.Log.Infoln("init role fail")
			os.Exit(0)
		}
		// 绑定角色
		user.BindRole(user.Id, role.Id)
		// 创建权限
		roleGroup := RoleGroup{RoleId: role.Id, Action: "RW", GroupId: 0}
		_, err = roleGroup.CreateRoleGroup(roleGroup.RoleId, roleGroup.GroupId, "RW", roleGroup.OrgId)
		if err != nil {
			util.Log.Infoln("init roleGroup fail")
			os.Exit(0)
		}
	}
	util.Log.Infoln("register table  success")
}
