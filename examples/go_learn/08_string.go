package main

import (
	"fmt"
	"strings"
)

// 示例 8: 字符串处理和格式化
// 交易所中经常需要处理交易对、金额格式化等

func main() {
	fmt.Println("=== Go 学习示例 8: 字符串处理 ===")

	// 1. 解析交易对 "BTC/USDT" -> [BTC, USDT]
	symbol := "BTC/USDT"
	parts := strings.Split(symbol, "/")
	if len(parts) == 2 {
		base := parts[0]  // BTC
		quote := parts[1] // USDT
		fmt.Printf("交易对解析: 基础货币=%s, 计价货币=%s\n", base, quote)
	}

	// 2. 字符串连接
	username := "alice"
	userID := uint64(12345)
	// 方式1: fmt.Sprintf
	key := fmt.Sprintf("user:%d:%s", userID, username)
	fmt.Printf("格式化字符串: %s\n", key)

	// 方式2: strings.Builder
	var sb strings.Builder
	sb.WriteString("cache:")
	sb.WriteString("user:")
	sb.WriteString(fmt.Sprintf("%d", userID))
	fmt.Printf("Builder 拼接: %s\n", sb.String())

	// 3. 字符串替换和修改
	orderType := "limit"
	orderType = strings.ToUpper(orderType) // 转大写
	fmt.Printf("转大写: %s\n", orderType)

	// 4. 字符串包含判断
	side := "buy"
	if strings.Contains(side, "buy") || strings.Contains(side, "sell") {
		fmt.Printf("订单方向有效: %s\n", side)
	}

	// 5. 修剪空格和换行
	rawInput := "  12345 \n"
	trimmed := strings.TrimSpace(rawInput)
	fmt.Printf("修剪后: '%s'\n", trimmed)

	// 6. 金额格式化展示
	price := 50000.123456
	// 保留 2 位小数
	fmt.Printf("价格格式化: %.2f\n", price)

	amount := 1.5
	fmt.Printf("数量格式化: %.4f\n", amount)

	// 7. 数字转字符串
	orderID := int64(123456789)
	orderIDStr := fmt.Sprintf("%d", orderID)
	fmt.Printf("订单ID字符串: %s\n", orderIDStr)

	// 8. 字符串转数字
	numStr := "100"
	num, _ := fmt.Sscanf(numStr, "%d", new(int))
	fmt.Printf("字符串转数字: %d\n", num)
}
