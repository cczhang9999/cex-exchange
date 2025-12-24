package data

import (
	"backend-v2/internal/conf"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewDB, NewRedis, NewExchangeClient)

type Data struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewData(db *gorm.DB, rdb *redis.Client) (*Data, func(), error) {
	d := &Data{
		db:  db,
		rdb: rdb,
	}
	return d, func() {
		fmt.Println("closing the data resources")
	}, nil
}

func NewDB(bc *conf.Bootstrap) *gorm.DB {
	dsn := bc.Data.Database.Source

	// 创建自定义的日志记录器
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // 慢 SQL 阈值
			LogLevel:                  logger.Info, // 日志级别
			IgnoreRecordNotFoundError: false,       // 忽略ErrRecordNotFound（记录未找到）错误
			Colorful:                  true,        // 彩色打印
		},
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger, // 设置日志记录器
	})
	if err != nil {
		panic(fmt.Sprintf("failed to open database: %v", err))
	}
	// 自动迁移数据库表结构
	//db.AutoMigrate(&User{}, &Order{})
	return db
}

func NewRedis(bc *conf.Bootstrap) *redis.Client {
	c := bc.Data.Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:         c.Addr,
		Password:     c.Password,
		DB:           0,
		ReadTimeout:  parseDuration(c.ReadTimeout),
		WriteTimeout: parseDuration(c.WriteTimeout),
	})
	// ping test
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		panic(fmt.Sprintf("failed to connect redis: %v", err))
	}
	return rdb
}

func parseDuration(s string) time.Duration {
	d, _ := time.ParseDuration(s)
	return d
}
