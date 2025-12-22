package main

import (
	"backend-v2/internal/conf"
	"backend-v2/internal/data"
	"context"
	"fmt"
	"os"
	"testing"
)

func TestCreateOrder(t *testing.T) {
	// 1. 初始化配置
	configPath := "../configs/config.yaml"
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		configPath = path
	}
	bc, err := conf.Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config from %s: %v", configPath, err)
	}
	// 2. 初始化数据层
	db := data.NewDB(bc)
	rdb := data.NewRedis(bc)
	d, cleanup, err := data.NewData(db, rdb)
	if err != nil {
		t.Fatalf("failed to init data: %v", err)
	}
	defer cleanup()
	repo := data.NewOrderRepo(d)
	// 3. 调用并验证
	ctx := context.Background()
	orderMyList, err := repo.FindByUserID(ctx, 1)
	if err != nil {
		t.Fatalf("FindOpenOrders failed: %v", err)
	}
	for _, o := range orderMyList {
		fmt.Printf("Order: %+v\n", o)
	}
}
