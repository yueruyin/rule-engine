# 规则引擎 (Rule Engine)

基于Golang实现的高性能规则引擎系统，主要技术栈包括gin + gorm + redis + mysql + gengine。

## 项目特点

- 高性能：基于Go语言开发，性能卓越
- 易扩展：模块化设计，便于功能扩展
- 可配置：支持多种数据存储方式，灵活配置规则

## 主要依赖

- gin: Web框架
- gorm: ORM框架
- go-redis: Redis客户端
- bilibili/gengine: 规则引擎核心库
- logrus: 日志管理
- yaml: 配置管理
- jwt: 认证授权
- panjf2000/ants: 高性能协程池

## 项目结构

```
engine/
├── cmd                   // 程序启动入口
├── config                // 配置文件
│   └── rules             // 规则配置
├── internal              // 业务逻辑（不对外暴露）
│   ├── handler           // 处理器层
│   ├── cache             // 缓存层
│   ├── file              // 文件存储
│   ├── repository        // 数据持久层
│   └── adapter           // 适配器
├── logs                  // 日志文件
├── pkg                   // 公共包
│   ├── e                 // 错误码定义
│   ├── middlewares       // 中间件
│   ├── util              // 工具函数
│   └── sql               // SQL相关
├── routes                // 路由定义
└── vendor                // 依赖包
```

## 配置文件说明

在`engine/config/config.yml`文件中设置项目配置信息：

```yaml
server:
  name: engine            # 服务名称
  port: 8080              # HTTP服务端口
  version: 1.0            # 服务版本
  grpcAddress: "127.0.0.1:10001"  # gRPC服务地址

mysql:
  driverName: mysql       # 数据库驱动
  host: localhost         # 数据库主机
  port: 3306              # 数据库端口
  database: rule-engine   # 数据库名
  username: root          # 用户名
  password:               # 密码
  charset: utf8mb4        # 字符集
  maxIdleConn: 20         # 最大空闲连接数
  maxOpenConn: 200        # 最大打开连接数
  connMaxLifetime: 300    # 连接最大生命周期(秒)

redis:
  address: 127.0.0.1:6379 # Redis地址
  password:               # Redis密码
  db: 0                   # Redis数据库
  poolSize: 1000          # 连接池大小
  maxIdleConn: 1000       # 最大空闲连接数
  minIdleConn: 50         # 最小空闲连接数
  poolTimeout: 4          # 连接池超时时间(秒)

rule:
  storage: db             # 规则存储类型[db,file]
  data: memory            # 数据存储类型[memory,redis]
  log: file               # 日志存储类型[file,es]
  poolMax: 10             # 规则引擎最大协程数
  poolMin: 5              # 规则引擎最小协程数
  debugSecond: 3000       # 调试模式过期时间(秒)
```

## 快速开始

### 环境要求

- Go 1.18+
- MySQL 5.7+
- Redis 6.0+

### 安装依赖

```bash
cd engine
go mod tidy
```

### 启动服务

开发环境：

```bash
cd engine
go run cmd/main.go
```

### 打包部署

```bash
# 设置目标平台
export GOARCH=amd64  # 或其他目标架构：arm64, 386等
export GOOS=linux    # 或其他目标系统：darwin, windows等

# 编译
cd engine
go build -o rule-server ./cmd/main.go
```

### Docker部署

```bash
docker build -t rule-engine .
docker run -p 8080:8080 rule-engine
```

## 许可证

本项目采用Apache License 2.0许可证。详见[LICENSE](License)文件。