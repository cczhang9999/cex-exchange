package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

// --- methods required by the extended biz.AccountRepo interface ---

func (r *fakeAccountRepo) FindAll(ctx context.Context) ([]*biz.Account, error) {
	out := make([]*biz.Account, 0, len(r.accounts))
	for _, a := range r.accounts {
		out = append(out, a)
	}
	return out, nil
}

func (r *fakeAccountRepo) Deposit(ctx context.Context, userID uint64, asset string, amount float64) error {
	a, ok := r.findAsset(userID, asset)
	if !ok {
		a = &biz.Account{ID: uint64(len(r.accounts) + 1), UserID: userID, Asset: asset, Balance: 0, Frozen: 0}
		r.accounts[a.ID] = a
	}
	a.Balance += amount
	return nil
}

func (r *fakeAccountRepo) Withdraw(ctx context.Context, userID uint64, asset string, amount float64, address string) error {
	a, ok := r.findAsset(userID, asset)
	if !ok || a.Balance < amount {
		return fmt.Errorf("余额不足")
	}
	a.Balance -= amount
	return nil
}

func (r *fakeAccountRepo) AdjustBalance(ctx context.Context, userID uint64, asset string, delta float64, changeType, remark string) error {
	a, ok := r.findAsset(userID, asset)
	if !ok {
		a = &biz.Account{ID: uint64(len(r.accounts) + 1), UserID: userID, Asset: asset, Balance: 0, Frozen: 0}
		r.accounts[a.ID] = a
	}
	a.Balance += delta
	return nil
}

func (r *fakeAccountRepo) findAsset(userID uint64, asset string) (*biz.Account, bool) {
	for _, a := range r.accounts {
		if a.UserID == userID && a.Asset == asset {
			return a, true
		}
	}
	return nil, false
}

func TestHttpAccount(t *testing.T) {
	repo := newFakeOrderRepo()
	uc := biz.NewOrderUsecase(repo)
	accUc := biz.NewAccountUsecase(newFakeAccountRepo())
	svc := service.NewExchangeService(nil, uc, accUc, nil, nil, nil, nil, nil)
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

// fakeAccountFlowRepo 是 biz.AccountFlowRepo 的内存实现，用于测试
type fakeAccountFlowRepo struct {
	flows []*biz.AccountFlow
}

func newFakeAccountFlowRepo() *fakeAccountFlowRepo {
	remark := "充值"
	refID := uint64(1001)
	now := time.Now()
	return &fakeAccountFlowRepo{
		flows: []*biz.AccountFlow{
			{ID: 1, UserID: 1, AccountID: 1, Asset: "USDT", ChangeType: "deposit", Amount: 1000, Balance: 1000, Remark: &remark, CreatedAt: now, UpdatedAt: now},
			{ID: 2, UserID: 1, AccountID: 2, Asset: "BTC", ChangeType: "trade", Amount: -0.5, Balance: 0, RefID: &refID, CreatedAt: now, UpdatedAt: now},
		},
	}
}

func (r *fakeAccountFlowRepo) Create(ctx context.Context, flow *biz.AccountFlow) (*biz.AccountFlow, error) {
	flow.ID = uint64(len(r.flows) + 1)
	flow.CreatedAt = time.Now()
	flow.UpdatedAt = time.Now()
	r.flows = append(r.flows, flow)
	return flow, nil
}

func (r *fakeAccountFlowRepo) FindPageByUserID(ctx context.Context, q biz.AccountFlowQuery) (*biz.AccountFlowPage, error) {
	var matched []*biz.AccountFlow
	for _, f := range r.flows {
		if f.UserID != q.UserID {
			continue
		}
		if q.Asset != "" && f.Asset != q.Asset {
			continue
		}
		if q.ChangeType != "" && f.ChangeType != q.ChangeType {
			continue
		}
		if !q.StartTime.IsZero() && f.CreatedAt.Before(q.StartTime) {
			continue
		}
		if !q.EndTime.IsZero() && f.CreatedAt.After(q.EndTime) {
			continue
		}
		matched = append(matched, f)
	}

	page, pageSize := q.Page, q.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	start := (page - 1) * pageSize
	if start > len(matched) {
		start = len(matched)
	}
	end := start + pageSize
	if end > len(matched) {
		end = len(matched)
	}

	return &biz.AccountFlowPage{
		Total:    int64(len(matched)),
		Page:     page,
		PageSize: pageSize,
		Flows:    matched[start:end],
	}, nil
}

func TestHttpAccountFlows(t *testing.T) {
	uc := biz.NewOrderUsecase(newFakeOrderRepo())
	flowUc := biz.NewAccountFlowUsecase(newFakeAccountFlowRepo())
	svc := service.NewExchangeService(nil, uc, nil, flowUc, nil, nil, nil, nil)
	r := server.NewHTTPServer(nil, svc)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/account_flows?asset=USDT&change_type=deposit&page=1&page_size=10",
		nil,
	)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	t.Logf("返回 JSON: %s", w.Body.String())

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}
