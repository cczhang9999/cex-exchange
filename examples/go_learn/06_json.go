package main

import (
	"encoding/json"
	"fmt"
)

// 示例 6: JSON 序列化/反序列化
// API 交互中经常需要处理 JSON 数据

// OrderRequest 下单请求结构
type OrderRequest struct {
	UserID uint64  `json:"user_id"` // JSON 字段名小写
	Symbol string  `json:"symbol"`
	Side   string  `json:"side"` // buy/sell
	Type   string  `json:"type"` // limit/market
	Price  float64 `json:"price"`
	Amount float64 `json:"amount"`
}

// OrderResponse 下单响应结构
type OrderResponse struct {
	OrderID int64  `json:"order_id"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"` // omitempty: 空值时不输出
}

// User 用户信息
type User struct {
	ID       uint64  `json:"id"`
	Username string  `json:"username"`
	Email    string  `json:"email,omitempty"`
	Balance  float64 `json:"balance"`
}

func main() {
	fmt.Println("=== Go 学习示例 6: JSON 处理 ===")

	// 1. 结构体转 JSON
	req := OrderRequest{
		UserID: 12345,
		Symbol: "BTC/USDT",
		Side:   "buy",
		Type:   "limit",
		Price:  50000.0,
		Amount: 0.5,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("序列化失败: %v\n", err)
		return
	}
	fmt.Printf("JSON 输出: %s\n", string(jsonData))

	// 2. JSON 字符串解析到结构体
	jsonStr := `{"order_id": 10086, "status": "success", "message": "Order placed"}`
	var resp OrderResponse
	err = json.Unmarshal([]byte(jsonStr), &resp)
	if err != nil {
		fmt.Printf("解析失败: %v\n", err)
		return
	}
	fmt.Printf("解析结果: %+v\n", resp)

	// 3. 使用 json.MarshalIndent 格式化输出
	user := User{
		ID:       1,
		Username: "alice",
		Email:    "alice@example.com",
		Balance:  10000.50,
	}
	prettyJSON, _ := json.MarshalIndent(user, "", "  ")
	fmt.Printf("格式化 JSON:\n%s\n", string(prettyJSON))

	// 4. 处理 JSON 中的未知字段
	type Response struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"` // 保留原始 JSON 数据
	}

	rawResp := `{"code": 200, "message": "success", "data": {"orders": [{"id": 1}, {"id": 2}]}}`
	var genericResp Response
	json.Unmarshal([]byte(rawResp), &genericResp)
	fmt.Printf("通用响应: code=%d, message=%s, raw_data=%s\n",
		genericResp.Code, genericResp.Message, string(genericResp.Data))
}
