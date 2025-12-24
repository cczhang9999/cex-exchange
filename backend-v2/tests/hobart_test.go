package main

import (
	"backend-v2/internal/conf"
	"backend-v2/internal/data"
	"context"
	"fmt"
	"os"
	"testing"
)

func TestBasic(t *testing.T) {

	uid := 1
	fmt.Println(uid)
	ss := fmt.Sprintf("user:orders:%d:", uid)
	fmt.Println(ss)
	type user struct {
		ID   int
		Name string
	}
	userList := []user{
		{1, "Alice"},
		{2, "Bob"},
		{3, "Charlie"},
	}
	fmt.Println("userList", userList)
	// 创建一个map来存储userList
	userMap := make(map[int][]user)
	fmt.Println("userMap", userMap)
	// 将userList中的元素放入map中
	userMap[1] = userList
	userMap[2] = userList
	fmt.Printf("&userList: %p\n", &userList)
	fmt.Println("userMap", userMap)
	fmt.Println("userMap", userMap[1])

	var filteredUsers []user
	for _, u := range userList {
		if u.ID != 1 {
			filteredUsers = append(filteredUsers, u)
		}
	}
	fmt.Println(filteredUsers)

}

func TestListUsers(t *testing.T) {
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
