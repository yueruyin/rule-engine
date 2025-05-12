package repository

import (
	"errors"
	"time"
	"zenith.engine.com/engine/pkg/e"
)
import "golang.org/x/crypto/bcrypt"

type User struct {
	Id         uint64    `gorm:"primarykey" json:"id"`                          // 自增
	UserName   string    `gorm:"column:username;comment:'用户名'" json:"username"` // 唯一
	Password   string    `gorm:"column:password;comment:'密码'" json:"password"`
	CreateTime time.Time `gorm:"column:create_time;default:null;comment:'新增时间'" json:"create_time"`
	UpdateTime time.Time `gorm:"column:update_time;default:null;comment:'修改时间'" json:"update_time"`
	Enabled    int64     `gorm:"type:tinyint(1);column:enabled;default:0;comment:'是否禁用'"  json:"enabled"`
	Deleted    int64     `gorm:"type:tinyint(1);column:deleted;default:0;comment:'是否删除'"  json:"deleted"`
	RoleId     uint64    `gorm:"type:bigint(20);column:role_id;comment:'角色id'" json:"role_id"`
	OrgId      uint64    `gorm:"type:bigint(20);column:org_id;comment:'组织id'" json:"org_id"`
}

// CreateUser 创建用户
func (*User) CreateUser(userName string, origPassword string, orgId uint64) (User, error) {
	password, err := bcrypt.GenerateFromPassword([]byte(origPassword), bcrypt.DefaultCost)
	if err != nil {
		return User{}, errors.New(e.GetMsg(e.ErrorAuthCheckTokenFail))
	}
	user := User{UserName: userName, Password: string(password), CreateTime: time.Now(), UpdateTime: time.Now(), OrgId: orgId}
	tx := DB.Begin()
	tx.Create(&user)
	tx.Commit()
	return user, nil
}

// GetUserInfoById 通过id获取用户信息
func (*User) GetUserInfoById(id uint64) (user User) {
	DB.Model(User{}).Where("id", id).First(&user)
	return user
}

// GetUserInfoByUserName 通过username获取用户信息
func (*User) GetUserInfoByUserName(username string) (user User) {
	if DB.Config.Dialector.Name() == "dm" {
		DB.Debug().Raw(
			"select id,username,`password`,create_time, update_time,enabled,deleted,role_id,org_id from `user` where username = ? LIMIT 1",
			username).Scan(&user)
	} else {
		DB.Debug().Model(User{}).Where(User{UserName: username}).First(&user)
	}
	return user
}

// ExistsUser 验证用户是否存在
func (*User) ExistsUser(userName string) bool {
	var count int64
	DB.Debug().Model(User{}).Where("username", userName).Count(&count)
	return count > 0
}

// UpdatePassword 更新密码
func (*User) UpdatePassword(id uint64, password string) {
	DB.Model(User{}).Where("id", id).Update("password", password)
}

// BindRole 绑定角色
func (*User) BindRole(id uint64, roleId uint64) {
	DB.Model(User{}).Where("id", id).Update("role_id", roleId)
}
