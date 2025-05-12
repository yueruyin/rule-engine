package repository

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/godoes/gorm-dameng"
	oracle "github.com/godoes/gorm-oracle"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"strconv"
	"strings"
	"sync"
	"time"
	"zenith.engine.com/engine/config"
)

var DataMap = sync.Map{}
var SystemMap = sync.Map{}

//InitData 初始化SQL执行节点可用数据库
func InitData() {
	data := config.Conf.Data
	for _, datum := range data {
		if strings.ToLower(datum.DriverName) == "dameng" {
			ConnectDataDBDM(datum)
		} else if strings.ToLower(datum.DriverName) == "oracle" {
			ConnectDataDBORACLE(datum)
		} else {
			ConnectDataDB(datum)
		}

	}
}

//InitSystem 初始化HTTP执行节点可用第三方系统
func InitSystem() {
	http := config.Conf.Http
	for _, h := range http {
		SystemMap.Store(h.System, h)
	}
}

//ConnectDataDB 连接MYSQL数据库
func ConnectDataDB(datum config.Data) {
	host := datum.Host
	port := datum.Port
	database := datum.DataBase
	username := datum.UserName
	password := datum.PassWord
	charset := datum.Charset
	dsn := strings.Join([]string{username, ":", password, "@tcp(", host, ":", port, ")/", database, "?charset=" + charset + "&parseTime=true"}, "")
	var ormLogger logger.Interface
	if gin.Mode() == "debug" {
		ormLogger = logger.Default.LogMode(logger.Info)
	} else {
		ormLogger = logger.Default
	}
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,   // DSN data source name
		DefaultStringSize:         256,   // string 类型字段的默认长度
		DisableDatetimePrecision:  true,  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,  // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false, // 根据版本自动配置
	}), &gorm.Config{
		Logger: ormLogger,
		//SkipDefaultTransaction: true,  // 关闭自动事务 性能提升30%左右
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		DataMap.Store(datum.Title, nil)
		fmt.Println("数据库[" + datum.Title + "]链接失败!  error:" + err.Error())
		return
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(datum.MaxIdleConn) //设置连接池，空闲
	sqlDB.SetMaxOpenConns(datum.MaxOpenConn) //打开
	sqlDB.SetConnMaxLifetime(time.Second * time.Duration(datum.ConnMaxLifetime))
	DataMap.Store(datum.Title, db)
}

//ConnectDataDBDM 连接dameng数据库
func ConnectDataDBDM(datum config.Data) {
	host := datum.Host
	port := datum.Port
	database := datum.DataBase
	username := datum.UserName
	password := datum.PassWord
	url := ""
	options := map[string]string{
		"schema":         database,
		"appName":        datum.Title,
		"connectTimeout": "30000",
	}
	if len(port) > 0 {
		portI, _ := strconv.Atoi(port)
		url = dameng.BuildUrl(username, password, host, portI, options)
	} else {
		url = "dm://" + username + ":" + password + "@" + host + "?connectTimeout=30000&schema=" + database + "&appName=" + datum.Title + "&compatibleMode=Mysql"
	}
	println(url)
	var ormLogger logger.Interface
	if gin.Mode() == "debug" {
		ormLogger = logger.Default.LogMode(logger.Info)
	} else {
		ormLogger = logger.Default
	}
	db, err := gorm.Open(dameng.Open(url), &gorm.Config{
		Logger: ormLogger, // DSN data source name
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		DataMap.Store(datum.Title, nil)
		fmt.Println("数据库[" + datum.Title + "]链接失败!  error:" + err.Error())
		return
	}
	sqlDB, err := db.DB()
	sqlDB.SetMaxIdleConns(datum.MaxIdleConn) //设置连接池，空闲
	sqlDB.SetMaxOpenConns(datum.MaxOpenConn) //打开
	sqlDB.SetConnMaxLifetime(time.Second * time.Duration(datum.ConnMaxLifetime))
	DataMap.Store(datum.Title, db)
}

//ConnectDataDBORACLE 连接oracle数据库
func ConnectDataDBORACLE(datum config.Data) {
	host := datum.Host
	port := datum.Port
	database := datum.DataBase
	username := datum.UserName
	password := datum.PassWord
	portI, _ := strconv.Atoi(port)
	options := map[string]string{
		"CONNECTION TIMEOUT": "90",
		"LANGUAGE":           "SIMPLIFIED CHINESE",
		"TERRITORY":          "CHINA",
		"SSL":                "false",
	}
	url := oracle.BuildUrl(host, portI, database, username, password, options)
	dialector := oracle.New(oracle.Config{
		DSN:                     url,
		IgnoreCase:              false, // query conditions are not case-sensitive
		NamingCaseSensitive:     true,  // whether naming is case-sensitive
		VarcharSizeIsCharLength: true,  // whether VARCHAR type size is character length, defaulting to byte length

		// RowNumberAliasForOracle11 is the alias for ROW_NUMBER() in Oracle 11g, defaulting to ROW_NUM
		RowNumberAliasForOracle11: "ROW_NUM",
	})
	db, err := gorm.Open(dialector, &gorm.Config{
		SkipDefaultTransaction:                   true, // 是否禁用默认在事务中执行单次创建、更新、删除操作
		DisableForeignKeyConstraintWhenMigrating: true, // 是否禁止在自动迁移或创建表时自动创建外键约束
		// 自定义命名策略
		NamingStrategy: schema.NamingStrategy{
			NoLowerCase:         true, // 是否不自动转换小写表名
			IdentifierMaxLength: 30,
			SingularTable:       true,
		},
		PrepareStmt:     false, // 创建并缓存预编译语句，启用后可能会报 ORA-01002 错误
		CreateBatchSize: 50,    // 插入数据默认批处理大小
	})
	sqlDB, err := db.DB()
	if err != nil {
		DataMap.Store(datum.Title, nil)
		fmt.Println("数据库[" + datum.Title + "]链接失败!  error:" + err.Error())
		return
	}
	sqlDB.SetMaxIdleConns(datum.MaxIdleConn) //设置连接池，空闲
	sqlDB.SetMaxOpenConns(datum.MaxOpenConn) //打开
	sqlDB.SetConnMaxLifetime(time.Second * time.Duration(datum.ConnMaxLifetime))
	DataMap.Store(datum.Title, db)
}
