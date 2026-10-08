package main

import (
	"backend-v2/internal/biz"
	"backend-v2/internal/conf"
	"backend-v2/internal/data"
	"context"
	"fmt"
	"os"
	"sort"
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
	// userList 转 map
	userMap := make(map[int]user, len(userList))
	for _, u := range userList {
		userMap[u.ID] = u
	}
	fmt.Println("userMap", userMap)

	var filteredUsers []user
	for _, u := range userList {
		if u.ID != 1 {
			filteredUsers = append(filteredUsers, u)
		}
	}
	a := []int{1, 2}
	b := []int{5, 4, 3}

	a = append(a, b...)
	fmt.Println(" a===", a)
	sort.Slice(a, func(i, j int) bool { return a[i] > a[j] })
	fmt.Println(" after sort===", a)
	fmt.Println(filteredUsers)

	fmt.Printf("status=%v\n", biz.OrderStatusOpen)
	fmt.Println(biz.OrderStatusOpen == "open")

	order := biz.Order{
		ID:     1001,
		UserID: 2001,
		Amount: 99.9,
	}

	fmt.Println(order)

}

func TestListUsers(t *testing.T) {
	// 1. 初始化配置
	configPath := "../configs/config-test.yaml"
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
	//for _, u := range users {
	//	fmt.Printf("ID: %d, Username: %s, Email: %s,CreatedAt: %v,UpdatedAt: %v\n", u.ID, u.Username, u.Email, u.CreatedAt, u.UpdatedAt)
	//}
}
