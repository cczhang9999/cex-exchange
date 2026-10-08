package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend-v2/internal/biz"
	"backend-v2/internal/server"
	"backend-v2/internal/service"
)

// fakeKlineRepo 是 biz.KlineRepo 的内存实现，用于测试
type fakeKlineRepo struct {
	klines []*biz.Kline
}

func newFakeKlineRepo() *fakeKlineRepo {
	now := time.Now()
	return &fakeKlineRepo{
		klines: []*biz.Kline{
			{ID: 1, Symbol: "BTC/USDT", Interval: "1m", Open: 65000.0, High: 65100.0, Low: 64900.0, Close: 65050.0, Volume: 12.5, OpenTime: now, CloseTime: now.Add(1 * time.Minute), CreatedAt: now, UpdatedAt: now},
			{ID: 2, Symbol: "BTC/USDT", Interval: "1m", Open: 65050.0, High: 65200.0, Low: 65030.0, Close: 65180.0, Volume: 8.3, OpenTime: now.Add(1 * time.Minute), CloseTime: now.Add(2 * time.Minute), CreatedAt: now, UpdatedAt: now},
			{ID: 3, Symbol: "BTC/USDT", Interval: "1m", Open: 65180.0, High: 65300.0, Low: 65150.0, Close: 65250.0, Volume: 15.0, OpenTime: now.Add(2 * time.Minute), CloseTime: now.Add(3 * time.Minute), CreatedAt: now, UpdatedAt: now},
			{ID: 4, Symbol: "ETH/USDT", Interval: "5m", Open: 3000.0, High: 3050.0, Low: 2990.0, Close: 3040.0, Volume: 100.0, OpenTime: now, CloseTime: now.Add(5 * time.Minute), CreatedAt: now, UpdatedAt: now},
		},
	}
}

func (r *fakeKlineRepo) Save(ctx context.Context, k *biz.Kline) (*biz.Kline, error) {
	k.ID = uint64(len(r.klines) + 1)
	k.CreatedAt = time.Now()
	k.UpdatedAt = time.Now()
	r.klines = append(r.klines, k)
	return k, nil
}

func (r *fakeKlineRepo) FindByQuery(ctx context.Context, q biz.KlineQuery) ([]*biz.Kline, error) {
	var out []*biz.Kline
	for _, k := range r.klines {
		if k.Symbol != q.Symbol || k.Interval != q.Interval {
			continue
		}
		if !q.StartTime.IsZero() && k.OpenTime.Before(q.StartTime) {
			continue
		}
		if !q.EndTime.IsZero() && k.OpenTime.After(q.EndTime) {
			continue
		}
		out = append(out, k)
	}
	if q.Limit > 0 && len(out) > q.Limit {
		out = out[:q.Limit]
	}
	return out, nil
}

func TestHttpKlines(t *testing.T) {
	klineUc := biz.NewKlineUsecase(newFakeKlineRepo())
	uc := biz.NewOrderUsecase(newFakeOrderRepo())
	svc := service.NewExchangeService(nil, uc, nil, nil, klineUc, nil, nil)
	r := server.NewHTTPServer(nil, svc)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/klines?symbol=BTC/USDT&interval=1m&limit=100",
		nil,
	)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	t.Logf("返回 JSON: %s", w.Body.String())

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestGetKlinesBySymbol(t *testing.T) {
	repo := newFakeKlineRepo()
	uc := biz.NewKlineUsecase(repo)
	ctx := context.Background()

	klines, err := uc.GetKlines(ctx, "BTC/USDT", "1m", 0, 0, 10)
	if err != nil {
		t.Fatalf("GetKlines 失败: %v", err)
	}
	if len(klines) != 3 {
		t.Fatalf("期望获取 3 条 BTC/USDT 的 K线，实际获取 %d 条", len(klines))
	}
	for _, k := range klines {
		t.Logf("K线: ID=%d, Symbol=%s, Interval=%s, Open=%f, High=%f, Low=%f, Close=%f, Volume=%f",
			k.ID, k.Symbol, k.Interval, k.Open, k.High, k.Low, k.Close, k.Volume)
	}
}

func TestGetKlinesWithLimit(t *testing.T) {
	repo := newFakeKlineRepo()
	uc := biz.NewKlineUsecase(repo)
	ctx := context.Background()

	klines, err := uc.GetKlines(ctx, "BTC/USDT", "1m", 0, 0, 2)
	if err != nil {
		t.Fatalf("GetKlines 失败: %v", err)
	}
	if len(klines) != 2 {
		t.Fatalf("期望获取 2 条 K线（limit=2），实际获取 %d 条", len(klines))
	}
}

func TestGetKlinesEmpty(t *testing.T) {
	repo := newFakeKlineRepo()
	uc := biz.NewKlineUsecase(repo)
	ctx := context.Background()

	klines, err := uc.GetKlines(ctx, "ETH/USDT", "1m", 0, 0, 10)
	if err != nil {
		t.Fatalf("GetKlines 失败: %v", err)
	}
	if len(klines) != 0 {
		t.Fatalf("期望获取 0 条 K线，实际获取 %d 条", len(klines))
	}
}

func TestGetKlinesWithTimeRange(t *testing.T) {
	repo := newFakeKlineRepo()
	uc := biz.NewKlineUsecase(repo)
	ctx := context.Background()

	now := time.Now()
	startTime := now.Add(1 * time.Minute).Unix()
	endTime := now.Add(2 * time.Minute).Unix()

	klines, err := uc.GetKlines(ctx, "BTC/USDT", "1m", startTime, endTime, 100)
	if err != nil {
		t.Fatalf("GetKlines 失败: %v", err)
	}
	if len(klines) != 1 {
		t.Fatalf("期望获取 1 条 K线（时间范围过滤），实际获取 %d 条", len(klines))
	}
	t.Logf("时间范围查询 K线: ID=%d, OpenTime=%v", klines[0].ID, klines[0].OpenTime)
}
