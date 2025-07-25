package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Name     string `yaml:"name"`
	} `yaml:"database"`
	JWT struct {
		Secret  string `yaml:"secret"`
		Expires int    `yaml:"expires"`
	} `yaml:"jwt"`
}

var (
	DB     *gorm.DB
	config Config
)

func loadConfig() {
	f, err := os.Open("config.yaml")
	if err != nil {
		log.Fatalf("无法打开配置文件: %v", err)
	}
	defer f.Close()
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&config); err != nil {
		log.Fatalf("配置文件解析失败: %v", err)
	}
}

func initDB() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Database.User,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.Name,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   logger.Default.LogMode(logger.Info), // 启用SQL日志
	})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	DB = db
}

func main() {
	loadConfig()
	initDB()
	DB.AutoMigrate(&User{}, &Account{}, &AccountFlow{}, &Order{}, &Trade{}, &Kline{})
	r := gin.Default()

	r.POST("/api/register", Register)
	r.POST("/api/login", Login)
	r.GET("/api/klines", GetKlines)

	auth := r.Group("/api", AuthMiddleware())
	auth.POST("/deposit", Deposit)
	auth.POST("/withdraw", Withdraw)
	auth.GET("/account_flows", ListAccountFlows)
	auth.POST("/order", PlaceOrder)
	auth.POST("/order/cancel/:id", CancelOrder)
	auth.GET("/orderbook", OrderBook)
	auth.GET("/trades", TradeHistory)
	auth.GET("/admin/users", AdminListUsers)
	auth.GET("/admin/orders", AdminListOrders)
	auth.GET("/admin/accounts", AdminListAccounts)
	auth.POST("/admin/accounts/add-funds", AdminAddFunds)
	auth.GET("/accounts", ListAccounts)
	auth.GET("/my_orders", ListMyOrders)
	auth.GET("/my_trades", ListMyTrades)

	// TODO: 受保护接口示例
	// r.GET("/api/profile", AuthMiddleware(), Profile)

	r.Run(fmt.Sprintf(":%d", config.Server.Port))
}
