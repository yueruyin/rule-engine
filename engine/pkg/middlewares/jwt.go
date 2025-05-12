package middlewares

import (
	"github.com/gin-gonic/gin"
	"zenith.engine.com/engine/internal"
	"zenith.engine.com/engine/pkg/e"
	"zenith.engine.com/engine/pkg/util"
)

// JWTAuthMiddleware  基于jwt认证
func JWTAuthMiddleware() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		// 客户端携带Token有三种方式 1.放在请求头 2.放在请求体 3.放在URI
		// 这里假设Token放在URI的Authorization中，并使用Bearer开头

		//authHeader := c.Request.Header.Get("Authorization")
		auth := ctx.Request.Header.Get("token")
		if len(auth) == 0 || auth == "token" {
			auth = ctx.Request.Header.Get("TOKEN")
		}
		authHeader := auth
		if len(authHeader) == 0 {
			internal.UnauthorizedError(ctx)
			ctx.Abort()
			return
		}
		info, err := util.ParseToken(auth)
		if err != nil {
			internal.Unauthorized(ctx, e.GetMsg(e.ErrorAuthCheckTokenFail))
			ctx.Abort()
			return
		}
		// 这里考虑是否续期token 延长活跃token寿命
		//info.StandardClaims.ExpiresAt = time.Now().Add(util.TokenExpireDuration).Unix()
		// 将当前请求的username信息保存到请求的上下文ctx上
		ctx.Set(util.USERINFO, info)
		ctx.Next()
	}
}
