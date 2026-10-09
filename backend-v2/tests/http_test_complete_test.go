package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	pb "backend-v2/api/proto"
	"backend-v2/internal/biz"
	"backend-v2/internal/conf"
	"backend-v2/internal/server"
	"backend-v2/internal/service"
)

// fakeUserRepo 是 biz.UserRepo 的最小内存实现，用于测试
type fakeUserRepo struct {
	users map[uint64]*biz.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: map[uint64]*biz.User{
		1: {ID: 1, Username: "alice", Password: "123456", Email: "a@b.c", Status: biz.UserStatusNormal, CreatedAt: time.Now()},
	}}
}

func (r *fakeUserRepo) Save(ctx context.Context, u *biz.User) (*biz.User, error) {
	u.ID = uint64(len(r.users) + 1)
	r.users[u.ID] = u
	return u, nil
}
func (r *fakeUserRepo) FindByUsername(ctx context.Context, username string) (*biz.User, error) {
	for _, u := range r.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, nil
}
func (r *fakeUserRepo) ValidatePassword(u *biz.User, password string) bool {
	return u != nil && u.Password == password
}
func (r *fakeUserRepo) FindUserList(ctx context.Context) ([]*biz.User, error) {
	out := make([]*biz.User, 0, len(r.users))
	for _, u := range r.users {
		out = append(out, u)
	}
	return out, nil
}
func (r *fakeUserRepo) FindByID(ctx context.Context, id uint64) (*biz.User, error) {
	return r.users[id], nil
}
func (r *fakeUserRepo) UpdateStatus(ctx context.Context, id uint64, status biz.UserStatus) error {
	if u, ok := r.users[id]; ok {
		u.Status = status
	}
	return nil
}

// fakeTradeRepo 是 biz.TradeRepo 的最小内存实现
type fakeTradeRepo struct {
	trades []*biz.Trade
}

func newFakeTradeRepo() *fakeTradeRepo {
	return &fakeTradeRepo{trades: []*biz.Trade{
		{ID: 1, Symbol: "BTC/USDT", BuyOrderID: 1, SellOrderID: 2, Price: 50000, Amount: 0.5, BuyUserID: 1, SellUserID: 2, CreatedAt: time.Now()},
	}}
}

func (r *fakeTradeRepo) ListByUserID(ctx context.Context, userID uint64) ([]*biz.Trade, error) {
	out := make([]*biz.Trade, 0)
	for _, t := range r.trades {
		if t.BuyUserID == userID || t.SellUserID == userID {
			cp := *t
			out = append(out, &cp)
		}
	}
	return out, nil
}

// newCompleteTestService 构建一个包含全部 usecase 的服务（用于 HTTP 集成测试）
func newCompleteTestService(t *testing.T, cfg *conf.Bootstrap) *service.ExchangeService {
	t.Helper()
	orderUc := biz.NewOrderUsecase(newFakeOrderRepo())
	accUc := biz.NewAccountUsecase(newFakeAccountRepo())
	flowUc := biz.NewAccountFlowUsecase(newFakeAccountFlowRepo())
	tradeUc := biz.NewTradeUsecase(newFakeTradeRepo())
	userUc := biz.NewUserUsecase(newFakeUserRepo(), cfg)
	return service.NewExchangeService(userUc, orderUc, accUc, flowUc, nil, nil, cfg, tradeUc)
}

func doReq(t *testing.T, handler http.HandlerFunc, method, target string, body interface{}, token string) *httptest.ResponseRecorder {
	t.Helper()
	var bodyReader *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		bodyReader = bytes.NewReader(b)
	}
	var req *http.Request
	if bodyReader != nil {
		req = httptest.NewRequest(method, target, bodyReader)
	} else {
		req = httptest.NewRequest(method, target, nil)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	handler(w, req)
	return w
}

func TestCompleteAPI(t *testing.T) {
	const secret = "complete-secret"
	cfg := grpcTestConfig(secret)
	svc := newCompleteTestService(t, cfg)
	r := server.NewHTTPServer(cfg, svc)
	token := genTestToken(t, secret, 1)

	// ---------- 行情 Ticker ----------
	w := doReq(t, r.ServeHTTP, http.MethodGet, "/api/ticker?symbol=BTC/USDT", nil, "")
	if w.Code != http.StatusOK {
		t.Fatalf("ticker status=%d body=%s", w.Code, w.Body.String())
	}
	var ar struct {
		Code int             `json:"code"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &ar); err != nil {
		t.Fatalf("ticker unmarshal: %v", err)
	}
	var ticker pb.Ticker
	if err := json.Unmarshal(ar.Data, &ticker); err != nil {
		t.Fatalf("ticker data: %v", err)
	}
	if ticker.Symbol != "BTC/USDT" {
		t.Fatalf("ticker symbol=%s want BTC/USDT", ticker.Symbol)
	}
	t.Logf("✅ GET /api/ticker -> symbol=%s last=%s", ticker.Symbol, ticker.LastPrice)

	// ---------- 下单 ----------
	w = doReq(t, r.ServeHTTP, http.MethodPost, "/api/order", map[string]string{
		"symbol": "BTC/USDT", "side": "buy", "type": "limit", "price": "50000", "amount": "0.1",
	}, token)
	if w.Code != http.StatusOK {
		t.Fatalf("order status=%d body=%s", w.Code, w.Body.String())
	}
	var orderResp struct {
		Code int `json:"code"`
		Data struct {
			OrderID int64  `json:"order_id"`
			Message string `json:"message"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &orderResp); err != nil {
		t.Fatalf("order unmarshal: %v", err)
	}
	if orderResp.Code != 0 || orderResp.Data.OrderID <= 0 {
		t.Fatalf("place order failed: %+v body=%s", orderResp, w.Body.String())
	}
	orderID := orderResp.Data.OrderID
	t.Logf("✅ POST /api/order -> order_id=%d", orderID)

	// ---------- 查询我的订单 ----------
	w = doReq(t, r.ServeHTTP, http.MethodGet, "/api/my_orders?symbol=BTC/USDT", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("my_orders status=%d body=%s", w.Code, w.Body.String())
	}
	t.Logf("✅ GET /api/my_orders -> %s", w.Body.String())

	// ---------- 我的成交 ----------
	w = doReq(t, r.ServeHTTP, http.MethodGet, "/api/my_trades", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("my_trades status=%d body=%s", w.Code, w.Body.String())
	}
	var tradeResp struct {
		Code int          `json:"code"`
		Data []*biz.Trade `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &tradeResp); err != nil {
		t.Fatalf("my_trades unmarshal: %v", err)
	}
	if len(tradeResp.Data) == 0 {
		t.Fatalf("my_trades should return at least 1 trade, got 0")
	}
	t.Logf("✅ GET /api/my_trades -> %d trades", len(tradeResp.Data))

	// ---------- 账户 ----------
	w = doReq(t, r.ServeHTTP, http.MethodGet, "/api/accounts", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("accounts status=%d body=%s", w.Code, w.Body.String())
	}
	t.Logf("✅ GET /api/accounts -> %s", w.Body.String())

	// ---------- 充值 ----------
	w = doReq(t, r.ServeHTTP, http.MethodPost, "/api/deposit", map[string]string{"asset": "USDT", "amount": "1000"}, token)
	if w.Code != http.StatusOK {
		t.Fatalf("deposit status=%d body=%s", w.Code, w.Body.String())
	}
	t.Logf("✅ POST /api/deposit -> %s", w.Body.String())

	// ---------- 提现（余额不足应失败） ----------
	w = doReq(t, r.ServeHTTP, http.MethodPost, "/api/withdraw", map[string]string{"asset": "USDT", "amount": "9999999"}, token)
	var errResp struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil || errResp.Code != 400 {
		t.Fatalf("withdraw-insufficient: expected body code 400, got http=%d body=%s", w.Code, w.Body.String())
	}
	t.Logf("✅ POST /api/withdraw (insufficient) rejected -> %s", w.Body.String())

	// ---------- 提现（充值后提现应成功） ----------
	w = doReq(t, r.ServeHTTP, http.MethodPost, "/api/withdraw", map[string]string{"asset": "USDT", "amount": "500"}, token)
	if w.Code != http.StatusOK {
		t.Fatalf("withdraw status=%d body=%s", w.Code, w.Body.String())
	}
	t.Logf("✅ POST /api/withdraw -> %s", w.Body.String())

	// ---------- 撤单 ----------
	w = doReq(t, r.ServeHTTP, http.MethodPost, "/api/order/cancel/"+strconv.FormatInt(orderID, 10), nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("cancel status=%d body=%s", w.Code, w.Body.String())
	}
	t.Logf("✅ POST /api/order/cancel/:id -> %s", w.Body.String())

	// ---------- 管理员接口 ----------
	w = doReq(t, r.ServeHTTP, http.MethodGet, "/api/admin/users", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("admin/users status=%d body=%s", w.Code, w.Body.String())
	}
	t.Logf("✅ GET /api/admin/users -> %s", w.Body.String())

	w = doReq(t, r.ServeHTTP, http.MethodGet, "/api/admin/accounts", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("admin/accounts status=%d body=%s", w.Code, w.Body.String())
	}
	t.Logf("✅ GET /api/admin/accounts -> %s", w.Body.String())

	w = doReq(t, r.ServeHTTP, http.MethodGet, "/api/admin/orders", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("admin/orders status=%d body=%s", w.Code, w.Body.String())
	}
	t.Logf("✅ GET /api/admin/orders -> %s", w.Body.String())

	w = doReq(t, r.ServeHTTP, http.MethodPost, "/api/admin/accounts/add-funds", map[string]interface{}{"user_id": uint64(1), "asset": "USDT", "amount": "100", "remark": "test"}, token)
	if w.Code != http.StatusOK {
		t.Fatalf("admin add-funds status=%d body=%s", w.Code, w.Body.String())
	}
	t.Logf("✅ POST /api/admin/accounts/add-funds -> %s", w.Body.String())

	w = doReq(t, r.ServeHTTP, http.MethodPost, "/api/admin/accounts/adjust", map[string]interface{}{"user_id": uint64(1), "asset": "USDT", "amount": "50", "remark": "adjust"}, token)
	if w.Code != http.StatusOK {
		t.Fatalf("admin adjust status=%d body=%s", w.Code, w.Body.String())
	}
	t.Logf("✅ POST /api/admin/accounts/adjust -> %s", w.Body.String())

	w = doReq(t, r.ServeHTTP, http.MethodPost, "/api/admin/users/1/block", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("admin block status=%d body=%s", w.Code, w.Body.String())
	}
	t.Logf("✅ POST /api/admin/users/:id/block -> %s", w.Body.String())
}
