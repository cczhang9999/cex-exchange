package data

import (
	"backend-v2/internal/conf"
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to open database: %v", err))
	}
	db.AutoMigrate(&User{})
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
