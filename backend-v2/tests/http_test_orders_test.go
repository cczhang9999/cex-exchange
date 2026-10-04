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

// fakeOrderRepo 是 biz.OrderRepo 的内存实现，用于测试
type fakeOrderRepo struct {
	orders map[uint64]*biz.Order
}

func newFakeOrderRepo() *fakeOrderRepo {
	return &fakeOrderRepo{orders: map[uint64]*biz.Order{}}
}

func (r *fakeOrderRepo) Save(ctx context.Context, order *biz.Order) (*biz.Order, error) {
	if order.ID == 0 {
		order.ID = uint64(len(r.orders) + 1)
	}
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()
	r.orders[order.ID] = order
	return order, nil
}

func (r *fakeOrderRepo) FindByID(ctx context.Context, id uint64) (*biz.Order, error) {
	return r.orders[id], nil
}

func (r *fakeOrderRepo) FindByUserID(ctx context.Context, userID uint64) ([]*biz.Order, error) {
	var out []*biz.Order
	for _, o := range r.orders {
		if o.UserID == userID {
			out = append(out, o)
		}
	}
	return out, nil
}

func (r *fakeOrderRepo) UpdateStatus(ctx context.Context, id uint64, status biz.OrderStatus) error {
	if o, ok := r.orders[id]; ok {
		o.Status = status
		o.UpdatedAt = time.Now()
	}
	return nil
}

func (r *fakeOrderRepo) UpdateFilled(ctx context.Context, id uint64, filled float64) error {
	if o, ok := r.orders[id]; ok {
		o.Filled = filled
		o.UpdatedAt = time.Now()
	}
	return nil
}

func (r *fakeOrderRepo) FindOpenOrders(ctx context.Context, symbol string) ([]*biz.Order, error) {
	var out []*biz.Order
	for _, o := range r.orders {
		if o.Symbol == symbol && o.Status == biz.OrderStatusOpen {
			out = append(out, o)
		}
	}
	return out, nil
}

func (r *fakeOrderRepo) FindByUserIDWithUser(ctx context.Context, userID uint64) ([]*biz.Order, error) {
	return r.FindByUserID(ctx, userID)
}

func (r *fakeOrderRepo) FindOrdersWithUserInfo(ctx context.Context, userID uint64) (map[uint64]*biz.Order, error) {
	out := map[uint64]*biz.Order{}
	orders, _ := r.FindByUserID(ctx, userID)
	for _, o := range orders {
		out[o.ID] = o
	}
	return out, nil
}

func (r *fakeOrderRepo) FindOrdersWithUserUsingJoins(ctx context.Context, symbol string) ([]*biz.Order, error) {
	var out []*biz.Order
	for _, o := range r.orders {
		if o.Symbol == symbol {
			out = append(out, o)
		}
	}
	return out, nil
}

func TestHttpMyOrders(t *testing.T) {
	repo := newFakeOrderRepo()
	uc := biz.NewOrderUsecase(repo)
	svc := service.NewExchangeService(nil, uc, nil, nil)
	r := server.NewHTTPServer(nil, svc)

	// 预置用户 1 的订单数据（service.GetUserOrders 内部硬编码查询 userID=1）
	//_, err := uc.CreateOrder(
	//	context.Background(),
	//	1,
	//	"BTC/USDT",
	//	biz.OrderSideBuy,
	//	biz.OrderTypeLimit,
	//	65000.0,
	//	0.5,
	//)
	//if err != nil {
	//	t.Fatalf("预置订单失败: %v", err)
	//}

	// 创建 HTTP 请求
	req := httptest.NewRequest(
		http.MethodGet,
		"/api/my_orders?user_id=1",
		nil,
	)

	// 创建响应记录器
	w := httptest.NewRecorder()

	// 执行请求
	r.ServeHTTP(w, req)

	// 打印返回的 JSON
	t.Logf("返回 JSON: %s", w.Body.String())

	// 判断 HTTP 状态码
	if w.Code != http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d",
			w.Code,
		)
	}
}
