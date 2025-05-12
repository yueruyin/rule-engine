package middlewares

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"time"
	"zenith.engine.com/engine/pkg/util"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		//开始时间
		start := time.Now()
		// 输出入参数
		param, _ := ioutil.ReadAll(c.Request.Body)
		// 重新赋值
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(param))
		//执行逻辑
		c.Next()
		//结束时间
		end := time.Now()
		//执行时间
		execTime := end.Sub(start)
		path := c.Request.URL.Path      //请求路径
		clientIP := c.ClientIP()        //请求IP
		method := c.Request.Method      //请求方式
		statusCode := c.Writer.Status() //请求状态
		util.Log.Infof("|%3d | %13v | %15s | %s %s | %s",
			statusCode,
			execTime,
			clientIP,
			method,
			path,
			string(param),
		)
	}
}
