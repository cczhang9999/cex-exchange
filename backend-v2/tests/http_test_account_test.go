package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend-v2/internal/biz"
	"backend-v2/internal/server"
	"backend-v2/internal/service"
)

// fakeAccountRepo 是 biz.AccountRepo 的内存实现，用于测试
type fakeAccountRepo struct {
	accounts map[uint64]*biz.Account
}

func newFakeAccountRepo() *fakeAccountRepo {
	return &fakeAccountRepo{
		accounts: map[uint64]*biz.Account{
			1: {ID: 1, UserID: 1, Asset: "USDT", Balance: 1000, Frozen: 0},
			2: {ID: 2, UserID: 1, Asset: "BTC", Balance: 0.5, Frozen: 0},
		},
	}
}

func (r *fakeAccountRepo) FindByUserID(ctx context.Context, userID uint64) ([]*biz.Account, error) {
	var out []*biz.Account
	for _, a := range r.accounts {
		if a.UserID == userID {
			out = append(out, a)
		}
	}
	return out, nil
}

func (r *fakeAccountRepo) FindByUserIDAndAsset(ctx context.Context, userID uint64, asset string) (*biz.Account, error) {
	for _, a := range r.accounts {
		if a.UserID == userID && a.Asset == asset {
			return a, nil
		}
	}
	return nil, nil
}

func TestHttpAccount(t *testing.T) {
	repo := newFakeOrderRepo()
	uc := biz.NewOrderUsecase(repo)
	accUc := biz.NewAccountUsecase(newFakeAccountRepo())
	svc := service.NewExchangeService(nil, uc, accUc, nil)
	r := server.NewHTTPServer(nil, svc)

	// 创建 HTTP 请求
	req := httptest.NewRequest(
		http.MethodGet,
		"/api/accounts",
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
