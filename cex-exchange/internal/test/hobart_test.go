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
}

func TestHash(t *testing.T) {
	ctx := context.Background()
	fmt.Println(ctx)
}
