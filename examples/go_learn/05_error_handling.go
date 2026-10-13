package main

import (
	"errors"
	"fmt"
)

// 示例 5: 错误处理模式
// Go 的错误处理是显式的，需要养成检查错误的习惯

// 自定义错误类型
var (
	ErrInsufficientBalance = errors.New("余额不足")
	ErrInvalidAmount       = errors.New("无效的金额")
	ErrOrderNotFound       = errors.New("订单不存在")
	ErrUnauthorized        = errors.New("无权操作")
)

// Order 订单结构
type Order struct {
	ID     int64
	UserID uint64
	Amount float64
}

// PlaceOrder 下单函数，演示错误处理
func PlaceOrder(userID uint64, amount float64, balance float64) (*Order, error) {
	// 参数验证
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	// 业务逻辑检查
	if balance < amount {
		return nil, ErrInsufficientBalance
	}

	// 创建订单
	order := &Order{
		ID:     12345,
		UserID: userID,
		Amount: amount,
	}

	return order, nil
}

// CancelOrder 取消订单，演示多层错误检查
func CancelOrder(order *Order, requestUserID uint64) error {
	if order == nil {
		return ErrOrderNotFound
	}

	// 验证权限
	if order.UserID != requestUserID {
		return ErrUnauthorized
	}

	// 取消订单逻辑...
	return nil
}

// ValidationError 自定义验证错误类型，实现 error 接口
type ValidationError struct {
	Field   string
	Message string
}

// Error 实现 error 接口
func (e *ValidationError) Error() string {
	return fmt.Sprintf("验证错误: %s - %s", e.Field, e.Message)
}

func main() {
	fmt.Println("=== Go 学习示例 5: 错误处理 ===")

	// 1. 基本错误处理
	order, err := PlaceOrder(1001, 0, 1000)
	if err != nil {
		fmt.Printf("下单失败: %v\n", err)
	} else {
		fmt.Printf("下单成功: %+v\n", order)
	}

	// 2. 使用 errors.Is 进行错误比较
	_, err = PlaceOrder(1001, -100, 1000)
	if errors.Is(err, ErrInvalidAmount) {
		fmt.Println("错误类型: 无效的金额")
	}

	// 3. 使用 errors.As 转换错误类型
	var validationErr *ValidationError
	customErr := &ValidationError{Field: "amount", Message: "金额必须大于0"}

	if errors.As(customErr, &validationErr) {
		fmt.Printf("验证错误: %s - %s\n", validationErr.Field, validationErr.Message)
	}

	// 4. 多层错误检查
	testOrder := &Order{ID: 1, UserID: 1001, Amount: 500}

	err = CancelOrder(testOrder, 9999) // 错误的用户尝试取消
	if err != nil {
		switch {
		case errors.Is(err, ErrOrderNotFound):
			fmt.Println("订单不存在")
		case errors.Is(err, ErrUnauthorized):
			fmt.Println("您无权取消此订单")
		default:
			fmt.Printf("取消失败: %v\n", err)
		}
	}
}
