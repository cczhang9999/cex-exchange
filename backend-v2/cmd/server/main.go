package main

import (
	"backend-v2/internal/conf"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	bc := &conf.Bootstrap{
		Server: &conf.Server{
			Http: &conf.ServerHTTP{Addr: ":8080", Timeout: "1s"},
			Grpc: &conf.ServerGRPC{Addr: ":9000", Timeout: "1s"},
		},
		Data: &conf.Data{
			Database: &conf.Database{Driver: "mysql", Source: "hobart:123456@tcp(212.227.166.131:9257)/cex_exchange?charset=utf8mb4&parseTime=True&loc=Local"},
			Redis:    &conf.Redis{Addr: "194.164.194.118:9502", Password: "pass123editmelol", ReadTimeout: "5s", WriteTimeout: "5s"},
		},
		Auth: &conf.Auth{
			JwtSecret: "super-secret-key-change-me",
			JwtExpiry: "24h",
		},
	}

	app, cleanup, err := initApp(bc)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	if err := app.Run(); err != nil {
		panic(err)
	}

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutting down server...")
}
