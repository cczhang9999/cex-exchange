package test

import (
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

func TestHash(t *testing.T) {
	ctx := context.Background()
	fmt.Println("dd", ctx)
}
