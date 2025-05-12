package repository

import (
	"context"
	"github.com/go-redis/redis/v9"
	"time"
	"zenith.engine.com/engine/config"
	"zenith.engine.com/engine/pkg/util"
)

func InitRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Network: "tcp",
		//Addr:         viper.GetString("redis.address"),
		//Password:     viper.GetString("redis.password"),
		//DB:           viper.GetInt("redis.db"),
		Addr:         config.Conf.Redis.Address,
		Password:     config.Conf.Redis.PassWord,
		DB:           config.Conf.Redis.DB,
		PoolSize:     config.Conf.Redis.PoolSize, // 连接池最大socket连接数，默认为4倍CPU数， 4 * runtime.NumCPU
		MaxIdleConns: config.Conf.Redis.MaxIdleConn,
		MinIdleConns: config.Conf.Redis.MinIdleConn,                              //在启动阶段创建指定数量的Idle连接，并长期维持idle状态的连接数不少于指定数量；
		DialTimeout:  5 * time.Second,                                            //连接建立超时时间，默认5秒。
		ReadTimeout:  10 * time.Second,                                           //读超时，默认3秒， -1表示取消读超时
		WriteTimeout: 10 * time.Second,                                           //写超时，默认等于读超时
		PoolTimeout:  time.Duration(config.Conf.Redis.PoolTimeout) * time.Second, //当所有连接都处在繁忙状态时，客户端等待可用连接的最大等待时长，默认为读超时+1秒。
	})
	var err error
	_, err = rdb.Ping(context.Background()).Result()
	if err != nil {
		util.Log.Println("redis 连接失败....")
		return nil
	}
	return rdb
}

//func InitRedis() *redis.Pool {
//	// 建立连接池
//	rp := &redis.Pool{
//		MaxIdle:     1,
//		MaxActive:   1,
//		IdleTimeout: 300,
//		Wait:        true,
//		Dial: func() (redis.Conn, error) {
//			con, err := redis.Dial("tcp", config.Conf.Redis.Address,
//				redis.DialPassword(config.Conf.Redis.PassWord),
//				redis.DialDatabase(config.Conf.Redis.DB),
//				redis.DialConnectTimeout(4*time.Second),
//				redis.DialReadTimeout(3*time.Second),
//				redis.DialWriteTimeout(3*time.Second))
//			if err != nil {
//				return nil, err
//			}
//			return con, nil
//		},
//	}
//	return rp
//}
