package main

import (
	"backend-v2/internal/conf"
	"backend-v2/internal/data"
	"context"
	"fmt"
	"testing"
)

func TestBasic(t *testing.T) {

	uid := 1
	fmt.Println(uid)
	ss := fmt.Sprintf("user:orders:%d:", uid)
	fmt.Println(ss)
	userList := []struct {
		ID   int
		Name string
	}{
		{1, "Alice"},
		{2, "Bob"},
		{3, "Charlie"},
	}


	fmt.Println("userList", userList)	

	for _, u := range userList {
		key := fmt.Sprintf("user:orders:%d:", u.ID)
		fmt.Printf("User: %s, Key: %s\n", u.Name, key)
	}



}

func TestListUsers(t *testing.T) {
	// 1. 初始化配置
	bc := &conf.Bootstrap{
		Data: &conf.Data{
			Database: &conf.Database{
				Driver: "mysql",
				Source: "hobart:123456@tcp(212.227.166.131:9257)/cex_exchange?charset=utf8mb4&parseTime=True&loc=Local",
			},
			Redis: &conf.Redis{
				Addr:         "194.164.194.118:9502",
				Password:     "pass123editmelol",
				ReadTimeout:  "5s",
				WriteTimeout: "5s",
			},
		},
	}

	// 2. 初始化数据层
	db := data.NewDB(bc)
	rdb := data.NewRedis(bc)
	d, cleanup, err := data.NewData(db, rdb)
	if err != nil {
		t.Fatalf("failed to init data: %v", err)
	}
	defer cleanup()

	repo := data.NewUserRepo(d)

	// 3. 调用并验证
	ctx := context.Background()
	users, err := repo.FindUserList(ctx)
	if err != nil {
		t.Fatalf("FindUserList failed: %v", err)
	}

	fmt.Printf("Successfully fetched %d users from database\n", len(users))
	for _, u := range users {
		fmt.Printf("ID: %d, Username: %s, Email: %s,CreatedAt: %v,UpdatedAt: %v\n", u.ID, u.Username, u.Email, u.CreatedAt, u.UpdatedAt)
	}
}

func TestHash(t *testing.T) {
	ctx := context.Background()
	fmt.Println("dd", ctx)
}
