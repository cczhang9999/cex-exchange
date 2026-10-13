package main

import (
	"fmt"
	"sync"
)

// 示例 4: 并发安全 - sync.Mutex 和 sync.WaitGroup
// 交易所订单匹配需要处理高并发，需要使用锁保护共享资源

type Account struct {
	mu      sync.Mutex
	Balance float64
	Name    string
}

// Deposit 存钱 - 使用互斥锁保证并发安全
func (a *Account) Deposit(amount float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Balance += amount
	fmt.Printf("%s 存钱: %.2f, 新余额: %.2f\n", a.Name, amount, a.Balance)
}

// Withdraw 取钱 - 使用互斥锁保证并发安全
func (a *Account) Withdraw(amount float64) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.Balance >= amount {
		a.Balance -= amount
		fmt.Printf("%s 取钱: %.2f, 新余额: %.2f\n", a.Name, amount, a.Balance)
		return true
	}
	fmt.Printf("%s 取钱失败: 余额不足 (余额: %.2f, 需要: %.2f)\n", a.Name, a.Balance, amount)
	return false
}

// Trade 交易 - 模拟买卖双方的原子操作
type TradeEngine struct {
	mu      sync.Mutex
	Buyers  map[uint64]*Account
	Sellers map[uint64]*Account
}

func NewTradeEngine() *TradeEngine {
	return &TradeEngine{
		Buyers:  make(map[uint64]*Account),
		Sellers: make(map[uint64]*Account),
	}
}

func (t *TradeEngine) Match(buyUserID, sellUserID uint64, price, amount float64) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	buyer := t.Buyers[buyUserID]
	seller := t.Sellers[sellUserID]

	if buyer == nil || seller == nil {
		fmt.Println("交易失败: 用户不存在")
		return false
	}

	totalCost := price * amount
	if buyer.Balance < totalCost {
		fmt.Printf("交易失败: 买方余额不足 (余额: %.2f, 需要: %.2f)\n", buyer.Balance, totalCost)
		return false
	}

	// 执行交易
	buyer.Balance -= totalCost
	seller.Balance += totalCost
	fmt.Printf("交易成功: 买方 %.2f -> 卖方 %.2f, 价格: %.2f, 数量: %.2f\n",
		buyUserID, sellUserID, price, amount)
	return true
}

func main() {
	fmt.Println("=== Go 学习示例 4: 并发安全 ===")

	// 演示账户并发操作
	var wg sync.WaitGroup
	account := &Account{Balance: 1000, Name: "钱包"}

	// 启动 5 个 goroutine 同时存钱
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			account.Deposit(100)
		}(i)
	}

	// 启动 3 个 goroutine 同时取钱
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			account.Withdraw(200)
		}(i)
	}

	wg.Wait()
	fmt.Printf("最终余额: %.2f\n", account.Balance)

	// 演示交易引擎
	fmt.Println("\n--- 交易引擎演示 ---")
	engine := NewTradeEngine()
	engine.Buyers[1] = &Account{Name: "Alice", Balance: 50000}
	engine.Sellers[2] = &Account{Name: "Bob", Balance: 1.0}

	// 执行交易: Alice 买入 0.5 BTC @ 50000
	engine.Match(1, 2, 50000, 0.5)
}
