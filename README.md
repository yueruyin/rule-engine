# zenith-rules

gin+grpc+gorm+mysql+redis 规则引擎


# 项目主要依赖
- gin
- gorm
- grpc
- logrus
- viper
- protobuf
- gengine
- go-redis
- go-mysql

# 项目结构

## 1. engine 规则引擎部分

```
engine/
├── cmd                   // 启动入口
├── config                // 配置文件
├── discovery             // etcd服务注册、keep-alive、获取服务信息等等
├── internal              // 业务逻辑（不对外暴露）
│   ├── handler           // 视图层
│   ├── cache             // 缓存
│   ├── file              // 文件存储
│   ├── repository        // 持久层
│   └── service           // 服务层
│       └──pb             // 放置生成的pb文件
├── logs                  // 放置打印日志模块
├── pkg                   // 各种包
│   ├── e                 // 统一错误状态码
│   ├── res               // 统一response接口返回
│   └── util              // 各种工具、JWT、Logger等等..
└── routes                // http路由模块

```
# 项目文件配置

各模块下的`config/config.yml`文件

```yaml
server:
# 模块
  name: engine
  port: 8080
  # 模块名称
  version: 1.0
  # 模块版本
  grpcAddress: "127.0.0.1:10001"
  # grpc地址

datasource:
# mysql数据源
  driverName: mysql
  host: localhost
  port: 3306
  database: rule-engine
  # 数据库名
  username: root
  password: 
  charset: utf8mb4
  maxIdleConn: 20
  maxOpenConn: 100

redis:
# redis 配置
  address: 127.0.0.1:6379
  password:
  db: 0

rule:
  # 规则引擎 配置
  storage: db # 规则存储类型[db,file]
  data: memory #数据存储类型 [memory,redis]
  log: file # 日志存储类型[file,es]
  poolMax: 10
  poolMin: 5
```

# 项目启动
- 在各模块下进行

```go
go mod tidy
```

- 在各模块下目录

```go
go run cmd/main.go
```

- 打包 GOARCH/GOOS 替换成自己需要环境的值

```go
export GOARCH=arm64
export GOOS=darwin
go build -o rule-server ./cmd/main.go 
```

```License```
zenith-rules is licensed under the Apache License 2.0. See the [LICENSE](License) file.