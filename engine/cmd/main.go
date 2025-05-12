package main

import (
	"zenith.engine.com/engine/config"
	"zenith.engine.com/engine/internal/cache"
	"zenith.engine.com/engine/internal/handler"
	"zenith.engine.com/engine/internal/repository"
	"zenith.engine.com/engine/routes"
)

func init() {
	// 读取配置
	config.InitConfig()
	// 数据库
	repository.InitDB()
	// redis
	cache.InitRedis()
	// 初始化默认规则
	handler.InitDefaultRule()
	// 初始化data中db
	repository.InitData()
	// 初始化http中可用第三方系统
	repository.InitSystem()
}

func main() {
	ginRouter := routes.NewRouter()
	//port := viper.GetString("server.port")
	port := config.Conf.Server.Port
	if port != "" {
		panic(ginRouter.Run(":" + port))
	}
	panic(ginRouter.Run()) // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

//func main() {
//	ginRouter := routes.NewRouter()
//	//ginRouter.LoadHTMLGlob("dist/*.html")
//	//ginRouter.LoadHTMLFiles("dist/js/*")               // 添加资源路径
//	//ginRouter.LoadHTMLFiles("dist/css/*")              // 添加资源路径
//	//ginRouter.LoadHTMLFiles("dist/img/*")              // 添加资源路径
//	//ginRouter.Static("/js", "dist/js")               // 添加资源路径
//	//ginRouter.Static("/css", "dist/css")             // 添加资源路径
//	//ginRouter.Static("/img", "dist/img")             // 添加资源路径
//	//ginRouter.Static("/fonts", "dist/fonts")         // 添加资源路径
//	//ginRouter.Static("/config.js", "dist/config.js") // 添加资源路径
//	//ginRouter.Static("//", "dist/index.html")
//	//ginRouter.StaticFile("/hello", "dist/index.html") //前端接口
//
//	//port := viper.GetString("server.port")
//	port := config.Conf.Server.Port
//	if port != "" {
//		panic(ginRouter.Run(":" + port))
//	}
//	panic(ginRouter.Run()) // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
//}
