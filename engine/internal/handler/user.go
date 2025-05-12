package handler

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"strconv"
	"zenith.engine.com/engine/internal"
	"zenith.engine.com/engine/internal/repository"
	"zenith.engine.com/engine/pkg/e"
	"zenith.engine.com/engine/pkg/util"
)

type LoginRequest struct {
	UserName string `form:"username" binding:"required,min=3" json:"username"`
	PassWord string `form:"password" binding:"required,min=3" json:"password"`
}

type PasswordRequest struct {
	OriginPassWord  string `form:"originPassword" binding:"required,min=3" json:"originPassword"`
	NewPassword     string `form:"newPassword" binding:"required,min=3" json:"newPassword"`
	ConfirmPassword string `form:"confirmPassword" binding:"required,min=3" json:"confirmPassword"`
}

// Login 登录
func Login(ctx *gin.Context) {
	var request LoginRequest
	err := ctx.BindJSON(&request)
	if err != nil {
		internal.Fail(ctx, nil, e.GetMsg(e.InvalidParams))
		return
	}
	var user repository.User
	if !user.ExistsUser(request.UserName) {
		internal.Fail(ctx, nil, e.GetMsg(e.ErrorExistUser))
		return
	}
	user = user.GetUserInfoByUserName(request.UserName)
	// 对比密码是否正确
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.PassWord))
	if err != nil {
		print("错误信息:" + err.Error())
		internal.Fail(ctx, nil, e.GetMsg(e.PasswordError))
		return
	}
	var role repository.Role
	role = role.GetRoleById(user.RoleId)
	// 生产token
	token, refreshToken, tokenExp, _ := util.GenToken(user.Id, user.UserName, role.Id, role.RoleName, user.OrgId)
	internal.Success(ctx, util.Token{Token: token, RefreshToken: refreshToken, TokenExp: tokenExp}, e.GetMsg(e.SUCCESS))
}

// Register 注册
func Register(ctx *gin.Context) {
}

// RefreshToken 刷新token
func RefreshToken(ctx *gin.Context) {
	refreshToken := ctx.Query("refreshToken")
	_, err := util.ParseToken(refreshToken)
	if err != nil {
		internal.Fail(ctx, nil, e.GetMsg(e.ErrorAuth))
		return
	}
	token, tokenExp, _ := util.RefreshToken(refreshToken)
	internal.Success(ctx, util.Token{Token: token, RefreshToken: refreshToken, TokenExp: tokenExp}, e.GetMsg(e.SUCCESS))
}

// UpdatePassword 更改密码
func UpdatePassword(ctx *gin.Context) {
	var request PasswordRequest
	err := ctx.BindJSON(&request)
	if err != nil {
		internal.Fail(ctx, nil, e.GetMsg(e.InvalidParams))
		return
	}
	cu := util.CurrentUser(ctx)
	var user repository.User
	user = user.GetUserInfoByUserName(cu.UserName)
	// 验证原始密码正确性
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.OriginPassWord))
	if err != nil {
		internal.Fail(ctx, nil, e.GetMsg(e.PasswordError))
		return
	}
	// 新密码比较
	if request.NewPassword != request.ConfirmPassword {
		internal.Fail(ctx, nil, e.GetMsg(e.PasswordInconsistentError))
		return
	}
	// 更新密码
	nb, err := bcrypt.GenerateFromPassword([]byte(request.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		internal.Fail(ctx, nil, e.GetMsg(e.GenPasswordError))
		return
	}
	user.UpdatePassword(user.Id, string(nb))
	internal.Success(ctx, true, e.GetMsg(e.SUCCESS))
}

// UserInfo 当前用户信息
func UserInfo(ctx *gin.Context) {
	v, b := ctx.Get(util.USERINFO)
	if !b {
		internal.Fail(ctx, nil, e.GetMsg(e.ErrorExistUser))
		return
	}
	userinfo := v.(*util.MyClaims)
	internal.Success(ctx, userinfo, e.GetMsg(e.SUCCESS))
}

// RoleList 获取全部角色
func RoleList(ctx *gin.Context) {
	var role repository.Role
	internal.Success(ctx, role.ListRole(), e.GetMsg(e.SUCCESS))
}

// RoleIdsByGroupId 通过groupId查询角色id
func RoleIdsByGroupId(ctx *gin.Context) {
	id := ctx.Query("id")
	groupId, _ := strconv.ParseUint(id, 10, 64)
	rgs := repository.GetRoleGroupByGroupId(groupId)
	internal.Success(ctx, rgs, e.GetMsg(e.SUCCESS))
}
