package biz

import (
	"context"
	"time"
)

// Kline 定义 K 线业务模型
type Kline struct {
	ID        uint64     `json:"id"`
	Symbol    string     `json:"symbol"`
	Interval  string     `json:"interval"`
	Open      float64    `json:"open"`
	High      float64    `json:"high"`
	Low       float64    `json:"low"`
	Close     float64    `json:"close"`
	Volume    float64    `json:"volume"`
	OpenTime  time.Time  `json:"open_time"`
	CloseTime time.Time  `json:"close_time"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// KlineQuery 定义 K 线查询条件
type KlineQuery struct {
	Symbol    string
	Interval  string
	StartTime time.Time
	EndTime   time.Time
	Limit     int
}

// KlineRepo 定义 K 线存储接口
type KlineRepo interface {
	// Save 保存 K 线数据
	Save(ctx context.Context, k *Kline) (*Kline, error)
	// FindByQuery 根据条件查询 K 线列表
	FindByQuery(ctx context.Context, q KlineQuery) ([]*Kline, error)
}

// KlineUsecase 定义 K 线业务逻辑
type KlineUsecase struct {
	repo KlineRepo
}

// NewKlineUsecase 创建 K 线业务逻辑实例
func NewKlineUsecase(repo KlineRepo) *KlineUsecase {
	return &KlineUsecase{
		repo: repo,
	}
}

// GetKlines 查询 K 线列表
func (uc *KlineUsecase) GetKlines(ctx context.Context, symbol, interval string, startTime, endTime int64, limit int32) ([]*Kline, error) {
	q := KlineQuery{
		Symbol:   symbol,
		Interval: interval,
	}
	if startTime > 0 {
		q.StartTime = time.Unix(startTime, 0)
	}
	if endTime > 0 {
		q.EndTime = time.Unix(endTime, 0)
	}
	if limit <= 0 {
		q.Limit = 500
	} else {
		q.Limit = int(limit)
	}
	return uc.repo.FindByQuery(ctx, q)
}

// SaveKline 保存 K 线数据
func (uc *KlineUsecase) SaveKline(ctx context.Context, k *Kline) (*Kline, error) {
	return uc.repo.Save(ctx, k)
}
