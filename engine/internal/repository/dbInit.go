package repository

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/godoes/gorm-dameng"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"strconv"
	"strings"
	"time"
	"zenith.engine.com/engine/config"
	"zenith.engine.com/engine/pkg/util"
)

var DB *gorm.DB

//InitDB 规则引擎主库初始化
func InitDB() {
	//host := viper.GetString("mysql.host")
	//port := viper.GetString("mysql.port")
	//database := viper.GetString("mysql.database")
	//username := viper.GetString("mysql.username")
	//password := viper.GetString("mysql.password")
	//charset := viper.GetString("mysql.charset")
	var err error
	host := config.Conf.Mysql.Host
	port := config.Conf.Mysql.Port
	database := config.Conf.Mysql.DataBase
	username := config.Conf.Mysql.UserName
	password := config.Conf.Mysql.PassWord
	charset := config.Conf.Mysql.Charset
	if config.Conf.Mysql.DriverName == "dameng" {
		options := map[string]string{
			"schema":         database,
			"appName":        database,
			"connectTimeout": "30000",
		}
		// 构建达梦连接URL
		url := ""
		if len(port) > 0 {
			portI, _ := strconv.Atoi(port)
			url = dameng.BuildUrl(username, password, host, portI, options)
		} else {
			url = "dm://" + username + ":" + password + "@" + host + "?connectTimeout=30000&schema=" + database + "&appName=" + database + "&compatibleMode=Mysql"
		}
		println(url)
		err = DatabaseDM(url)
	} else {
		dsn := strings.Join([]string{username, ":", password, "@tcp(", host, ":", port, ")/", database, "?charset=" + charset + "&parseTime=true"}, "")
		//dsn := strings.Join([]string{"root", ":", "root", "@tcp(", "127.0.0.1", ":", "3306", ")/", "basicInfo", "?charset=" + "utf8mb4" + "&parseTime=true"}, "")
		err = Database(dsn)
	}
	if err != nil {
		fmt.Println(err)
		util.Log.Error(err)
	}

}

//Database 规则引擎主库建立连接MYSQL
func Database(connString string) error {
	var ormLogger logger.Interface
	if gin.Mode() == "debug" {
		ormLogger = logger.Default.LogMode(logger.Info)
	} else {
		ormLogger = logger.Default
	}
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       connString, // DSN data source name
		DefaultStringSize:         256,        // string 类型字段的默认长度
		DisableDatetimePrecision:  true,       // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,       // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,       // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false,      // 根据版本自动配置
	}), &gorm.Config{
		Logger: ormLogger,
		//SkipDefaultTransaction: true,  // 关闭自动事务 性能提升30%左右
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		panic(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(config.Conf.Mysql.MaxIdleConn) //设置连接池，空闲
	sqlDB.SetMaxOpenConns(config.Conf.Mysql.MaxOpenConn) //打开
	sqlDB.SetConnMaxLifetime(time.Second * time.Duration(config.Conf.Mysql.ConnMaxLifetime))
	DB = db
	migration()
	return err
}

//DatabaseDM 规则引擎主库建立连接dameng
func DatabaseDM(connString string) error {
	var ormLogger logger.Interface
	if gin.Mode() == "debug" {
		ormLogger = logger.Default.LogMode(logger.Info)
	} else {
		ormLogger = logger.Default
	}
	db, err := gorm.Open(dameng.Open(connString), &gorm.Config{
		Logger: ormLogger, // DSN data source name
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		panic(err)
	}
	sqlDB, err := db.DB()
	sqlDB.SetMaxIdleConns(config.Conf.Mysql.MaxIdleConn) //设置连接池，空闲
	sqlDB.SetMaxOpenConns(config.Conf.Mysql.MaxOpenConn) //打开
	sqlDB.SetConnMaxLifetime(time.Second * time.Duration(config.Conf.Mysql.ConnMaxLifetime))
	DB = db
	var versionInfo []map[string]interface{}
	db.Table("SYS.V$VERSION").Find(&versionInfo)
	if err := db.Error; err == nil {
		versionBytes, _ := json.MarshalIndent(versionInfo, "", "  ")
		fmt.Printf("达梦数据库版本信息：\n%s\n", versionBytes)
	}
	return err
}
