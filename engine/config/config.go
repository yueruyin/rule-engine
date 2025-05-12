package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var Conf *Config

type Config struct {
	Server   Server   `yaml:"server"`
	Mysql    Mysql    `yaml:"mysql"`
	Redis    Redis    `yaml:"redis"`
	Rule     Rule     `yaml:"rule"`
	Variable Variable `yaml:"variable"`
	Data     []Data   `yaml:"data"`
	Http     []Http   `yaml:"http"`
}

type Mysql struct {
	DriverName      string `yaml:"driverName"`
	Host            string `yaml:"host"`
	Port            string `yaml:"port"`
	DataBase        string `yaml:"database"`
	UserName        string `yaml:"username"`
	PassWord        string `yaml:"password"`
	Charset         string `yaml:"charset"`
	MaxIdleConn     int    `yaml:"maxIdleConn"`
	MaxOpenConn     int    `yaml:"maxOpenConn"`
	ConnMaxLifetime int    `yaml:"connMaxLifetime"`
}

type Redis struct {
	Address     string `yaml:"address"`
	PassWord    string `yaml:"password"`
	DB          int    `yaml:"db"`
	PoolSize    int    `yaml:"poolSize"`
	MaxIdleConn int    `yaml:"maxIdleConn"`
	MinIdleConn int    `yaml:"minIdleConn"`
	PoolTimeout int    `yaml:"poolTimeout"`
}

type Server struct {
	Name        string `yaml:"name"`
	Port        string `yaml:"port"`
	Version     string `yaml:"version"`
	GrpcAddress string `yaml:"grpcAddress"`
}

type Rule struct {
	Storage     string `yaml:"storage"`
	Data        string `yaml:"data"`
	Log         string `yaml:"log"`
	PoolMax     int64  `yaml:"poolMax"`
	PoolMin     int64  `yaml:"poolMin"`
	DebugSecond int64  `yaml:"debugSecond"`
}

type Variable struct {
	ApiAll    string `yaml:"apiAll"`
	ApiSingle string `yaml:"apiSingle"`
	ApiDefine string `yaml:"apiDefine"`
	Mock      bool   `yaml:"mock"`
}

type Data struct {
	Title           string `yaml:"title"`
	DriverName      string `yaml:"driverName"`
	Host            string `yaml:"host"`
	Port            string `yaml:"port"`
	DataBase        string `yaml:"database"`
	UserName        string `yaml:"username"`
	PassWord        string `yaml:"password"`
	Charset         string `yaml:"charset"`
	MaxIdleConn     int    `yaml:"maxIdleConn"`
	MaxOpenConn     int    `yaml:"maxOpenConn"`
	ConnMaxLifetime int    `yaml:"connMaxLifetime"`
}

type Http struct {
	System string `yaml:"system"`
	Host   string `yaml:"host"`
	Oauth  Oauth  `yaml:"oauth"`
}

type Oauth struct {
	Enable    bool   `yaml:"enable"`
	Type      string `yaml:"type"`
	Url       string `yaml:"url"`
	Token     string `yaml:"token"`
	AppKey    string `yaml:"appKey"`
	AppSecret string `yaml:"appSecret"`
	UserName  string `yaml:"userName"`
	Password  string `yaml:"password"`
}

func InitConfig() {
	//workDir, _ := os.Getwd()
	//viper.SetConfigName("config")
	//viper.SetConfigType("yml")
	//viper.AddConfigPath(workDir + "/config")
	//fmt.Println(viper.GetString("datasource.host"))
	//err := viper.ReadInConfig()
	//if err != nil {
	//	panic(err)
	//}

	//打包时需修改
	//f, err := os.Open(GetAppPath() + "/rule-config.yml")

	//绝对路径打开配置文件
	dir, _ := os.Getwd()
	f, err := os.Open(dir + "/config/config.yml")

	//_, filename, _, _ := runtime.Caller(0)
	//dir := path.Dir(path.Dir(filename))
	//f, err := os.Open(dir + "/config/config-qijiang.yml")

	if err != nil {
		println(fmt.Sprintf("读取配置文件错误: %+v", err))
	}
	b, _ := ioutil.ReadAll(f)
	if err = yaml.Unmarshal(b, &Conf); err != nil {
		return
	}
}

func GetAppPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	index := strings.LastIndex(path, string(os.PathSeparator))

	return path[:index]
}
