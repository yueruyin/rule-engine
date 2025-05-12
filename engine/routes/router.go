package routes

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"zenith.engine.com/engine/internal/adapter"
	. "zenith.engine.com/engine/internal/handler"
	"zenith.engine.com/engine/pkg/middlewares"
)

var ginRouter = gin.Default()

func init() {
	RegisterRouterEvent(AddRouter)
}

func Cors() gin.HandlerFunc {
	return func(context *gin.Context) {
		method := context.Request.Method

		context.Header("Access-Control-Allow-Origin", "*")
		context.Header("Access-Control-Allow-Credentials", "true")
		context.Header("Access-Control-Allow-Headers", "*")
		context.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		context.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")

		if method == "OPTIONS" {
			context.AbortWithStatus(http.StatusNoContent)
		}
		context.Next()
	}
}

func NewRouter(service ...interface{}) *gin.Engine {
	//ginRouter:= gin.Default()
	ginRouter.Use(Cors())
	v1 := ginRouter.Group("/v1")
	{
		v1.GET("ping", func(context *gin.Context) {
			context.JSON(200, "success")
		})
		// 无需登录保护
		{
			v1.POST("/user/register", Register)
			v1.POST("/user/login", Login)
			v1.POST("/refresh/token", RefreshToken)
		}
		// 需要登录保护
		authed := v1.Group("/")
		authed.Use(middlewares.JWTAuthMiddleware())
		{
			// 规则设计
			v1.GET("engine/list", ListDesign)
			// 获取规则信息
			//authed.GET("engine", GetDesign)
			// 通过id获取规则信息
			v1.GET("engine/:id", GetDesignById)
			// 保存规则
			v1.POST("engine", SaveDesign)
			// 发布规则
			v1.POST("engine/publish", PublishDesign)
			// 修改规则
			v1.POST("engine/update", UpdateDesign)
			// 批量删除
			v1.POST("engine/delete", DeleteDesigns)
			// 删除规则
			v1.POST("engine/delete/:id", DeleteDesign)
			// 验证规则是否存在
			v1.GET("engine/check", CheckDesign)
			// 验证规则编码是否存在
			v1.GET("engine/check/code", CheckCode)
			// 获取设计参数(无value)
			v1.GET("engine/param/define", GetParamDefine)
			// 拷贝规则
			v1.POST("engine/copy/:id", CopyDesign)
			// 根据编码获取全部版本
			authed.GET("engine/list/version", GetListVersion)
			// 根据版本编码获取设计信息
			authed.GET("engine/get/design", GetEngineDesign)
		}
		{
			// 预执行规则
			authed.POST("execute/rule/preview", PreviewExecuteRule)
			// 获取全部分组及规则
			authed.GET("rule/group/by/id", ListByGroupId)
			// 获取子规则所需执行参数
			authed.GET("rule/execute/arguments", RuleExecuteArguments)
		}
		{
			// 获取可选数据库
			authed.GET("engine/data/list", DataList)
			// 规则sql验证
			authed.POST("engine/sql/validate", ValidateSql)
			// 提前执行sql 预览结果
			authed.POST("engine/sql/preview", PreviewSql)
		}
		{
			// 获取可选三方系统
			authed.GET("engine/system/list", SystemList)
			// 测试请求Send
			authed.POST("engine/http/send", SendHttp)
		}
		{
			// json节点预览
			authed.POST("engine/json/preview", PreviewFormatJson)
		}
		{
			// 执行规则
			v1.POST("execute/rule", ExecuteRule)
			// 批量执行规则
			v1.POST("execute/rule/batch", ExecuteRuleBatch)
			// 批量执行规则(携程并行)
			v1.POST("execute/rule/batch/parallel", ExecuteRuleBatchParallel)
			// 通过id执行规则 (每次重新加载规则到内存,result返回单个值结果,后续有类型需求也可使用)
			v1.POST("execute/rule/id", ExecuteRuleId)
			// Debug 执行规则
			authed.POST("debug/execute/rule", DeBugExecuteRule)
			// 重新执行错误后规则
			authed.POST("retry/execute/rule", RetryExecuteRule)
			// 获取当前正在执行的规则数量
			//authed.GET("health/running", HealthRunning)
		}
		{
			// 规则分组
			authed.GET("rule/group/list", ListGroup)
			authed.GET("rule/group", GetGroupInfo)
			authed.POST("rule/group", SaveGroup)
			authed.POST("rule/group/update", UpdateGroup)
			authed.POST("rule/group/delete", DeleteGroup)
		}
		{
			// 模版
			authed.GET("rule/template/tag/list", ListTemplateTag)
			authed.GET("rule/template/tag/tree", TreeTemplate)
			//authed.GET("rule/template/list", ListTemplate)
		}
		{
			// 默认规则
			authed.GET("default/rule/list", ListDefaultInfo)
			authed.GET("default/rule", GetDefaultInfo)
			authed.POST("default/rule", SaveDefaultInfo)
			authed.POST("default/rule/update", UpdateDefaultInfo)
			authed.POST("default/rule/delete:id", DeleteDefaultInfo)
		}
		{
			// 用户相关
			authed.POST("/user/password", UpdatePassword)
			authed.GET("/user/info", UserInfo)
			//角色相关
			authed.GET("/role/list", RoleList)
			authed.GET("/role/by/group/id", RoleIdsByGroupId)
		}
	}
	ruleDesign, err1 := adapter.GetStorage().GetDataDesign("rule_design")
	if err1 == nil {
	}
	//获取数据库内路由uri
	for i := 0; i < len(ruleDesign); i++ {
		AddRouter(ruleDesign[i].Uri)
	}
	ginRouter.Use(middlewares.Logger())
	return ginRouter
}

/*AddRouter 自动添加路由方式：
 *POST+接口->发送内容:{"count":x,"businessId":"xxx","arguments":{}}
 */
func AddRouter(uri string) {
	if len(uri) == 0 {
		return
	}
	if CheckRouter(uri) {
		return
	}
	v1 := ginRouter.Group("/")
	{
		v1.POST(uri, RouterExecuteRule) //API接口
	}
}

// CheckRouter 验证路由是否存在
func CheckRouter(uri string) bool {
	rs := ginRouter.Routes()
	for _, r := range rs {
		if r.Path == uri {
			return true
		}
	}
	return false
}
