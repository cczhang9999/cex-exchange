package data

import (
	"backend-v2/internal/biz"
	"context"
	"time"

	"github.com/google/wire"
	"gorm.io/gorm"
)

var KlineProviderSet = wire.NewSet(NewKlineRepo)

// Kline GORM Model
// 对应表结构：
// id bigint(20) unsigned AUTO_INCREMENT  K线ID
// symbol varchar(32)  交易对
// interval varchar(16)  K线周期，如1m,5m,1h
// open decimal(32,16)  开盘价
// high decimal(32,16)  最高价
// low decimal(32,16)  最低价
// close decimal(32,16)  收盘价
// volume decimal(32,16)  成交量
// open_time timestamp  K线开始时间
// close_time timestamp  K线结束时间
// created_at / updated_at / deleted_at
type Kline struct {
	gorm.Model
	Symbol    string
	Interval  string
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	OpenTime  time.Time
	CloseTime time.Time
}

type klineRepo struct {
	data *Data
}

func NewKlineRepo(data *Data) biz.KlineRepo {
	return &klineRepo{
		data: data,
	}
}

// Save 保存 K 线数据
func (r *klineRepo) Save(ctx context.Context, k *biz.Kline) (*biz.Kline, error) {
	m := &Kline{
		Symbol:    k.Symbol,
		Interval:  k.Interval,
		Open:      k.Open,
		High:      k.High,
		Low:       k.Low,
		Close:     k.Close,
		Volume:    k.Volume,
		OpenTime:  k.OpenTime,
		CloseTime: k.CloseTime,
	}
	if err := r.data.db.WithContext(ctx).Create(m).Error; err != nil {
		return nil, err
	}
	k.ID = uint64(m.ID)
	k.CreatedAt = m.CreatedAt
	k.UpdatedAt = m.UpdatedAt
	return k, nil
}

// FindByQuery 根据条件查询 K 线列表
func (r *klineRepo) FindByQuery(ctx context.Context, q biz.KlineQuery) ([]*biz.Kline, error) {
	tx := r.data.db.WithContext(ctx).Model(&Kline{}).
		Where("symbol = ? AND interval = ?", q.Symbol, q.Interval)
	if !q.StartTime.IsZero() {
		tx = tx.Where("open_time >= ?", q.StartTime)
	}
	if !q.EndTime.IsZero() {
		tx = tx.Where("open_time <= ?", q.EndTime)
	}
	var models []Kline
	if err := tx.Order("open_time ASC").Limit(q.Limit).Find(&models).Error; err != nil {
		return nil, err
	}

	res := make([]*biz.Kline, 0, len(models))
	for _, m := range models {
		res = append(res, &biz.Kline{
			ID:        uint64(m.ID),
			Symbol:    m.Symbol,
			Interval:  m.Interval,
			Open:      m.Open,
			High:      m.High,
			Low:       m.Low,
			Close:     m.Close,
			Volume:    m.Volume,
			OpenTime:  m.OpenTime,
			CloseTime: m.CloseTime,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		})
	}
	return res, nil
}
