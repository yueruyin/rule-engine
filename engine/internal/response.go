package internal

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"zenith.engine.com/engine/pkg/e"
)

func Response(ctx *gin.Context, httpStatus int, code int, data interface{}, message string) {
	ctx.JSON(httpStatus, gin.H{"code": code, "data": data, "message": message})
}

func Success(ctx *gin.Context, data interface{}, message string) {
	Response(ctx, http.StatusOK, e.SUCCESS, data, message)
}

func Fail(ctx *gin.Context, data interface{}, message string) {
	Response(ctx, http.StatusOK, e.ERROR, data, message)
}

func Unauthorized(ctx *gin.Context, message string) {
	Response(ctx, http.StatusUnauthorized, http.StatusUnauthorized, nil, message)
}

func UnauthorizedError(ctx *gin.Context) {
	Response(ctx, http.StatusUnauthorized, http.StatusUnauthorized, nil, "安全认证失败")
}
