package util

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"time"
)

// Token 返回struct
type Token struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
	TokenExp     int64  `json:"tokenExp"`
}

// MyClaims 用来生成token的struct
type MyClaims struct {
	Id       uint64 `json:"id"`
	UserName string `json:"username"`
	RoleId   uint64 `json:"roleId"`
	RoleName string `json:"rolename"`
	OrgId    uint64 `json:"orgId"`
	jwt.StandardClaims
}

//TokenExpireDuration token的过期时长
const TokenExpireDuration = time.Hour * 1
const RefreshTokenExpireDuration = time.Hour * 24

const USERINFO = "userinfo"

//MySecret ,签名时使用
var MySecret = []byte("rule")

// GenToken 创建token
func GenToken(id uint64, username string, roleId uint64, roleName string, orgId uint64) (string, string, int64, error) {
	tokenExp := time.Now().Add(TokenExpireDuration).Unix()
	c := MyClaims{
		id,
		username,
		roleId,
		roleName,
		orgId,
		// 自定义字段
		jwt.StandardClaims{
			ExpiresAt: tokenExp, // 过期时间
			Issuer:    "zenith", // 签发人
		},
	}

	r := MyClaims{
		id,
		username,
		roleId,
		roleName,
		orgId,
		// 自定义字段
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(RefreshTokenExpireDuration).Unix(), // 过期时间
			Issuer:    "zenith",                                          // 签发人
		},
	}

	// 使用指定的签名方法创建签名对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	// 使用指定的secret签名并获得完整的编码后的字符串token
	tk, err := token.SignedString(MySecret)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, r)
	rtk, err := refreshToken.SignedString(MySecret)
	if err != nil {
		return "", "", 0, err
	}
	return tk, rtk, tokenExp, nil
}

// ParseToken 解析token
func ParseToken(tokenString string) (*MyClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (i interface{}, err error) {
		return MySecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid { // 校验token
		//fmt.Println("token ok")
		//fmt.Println(claims.UserName)
		claims.ExpiresAt = time.Now().Add(TokenExpireDuration).Unix()
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshToken 刷新token
func RefreshToken(refreshToken string) (string, int64, error) {
	// 解析token
	claims, err := ParseToken(refreshToken)
	if err != nil {
		return "", 0, nil
	}
	tokenExp := time.Now().Add(TokenExpireDuration).Unix()
	// 刷新token
	c := MyClaims{
		claims.Id,
		claims.UserName,
		claims.RoleId,
		claims.RoleName,
		claims.OrgId,
		// 自定义字段
		jwt.StandardClaims{
			ExpiresAt: tokenExp, // 过期时间
			Issuer:    "zenith", // 签发人
		},
	}
	// 更新token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	// 使用指定的secret签名并获得完整的编码后的字符串token
	tk, err := token.SignedString(MySecret)
	return tk, tokenExp, err
}

// CurrentUser 获取当前用户信息
func CurrentUser(ctx *gin.Context) *MyClaims {
	user, _ := ctx.Get(USERINFO)
	if user == nil {
		return nil
	}
	return user.(*MyClaims)
}
