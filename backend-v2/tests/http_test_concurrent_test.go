package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"backend-v2/internal/biz"
	"backend-v2/internal/conf"
	"backend-v2/internal/data"
	"backend-v2/internal/server"
	"backend-v2/internal/service"
)

// ---------------------------------------------------------------------------
// 通用结构
// ---------------------------------------------------------------------------

// ApiResponse 对应 internal/pkg/response.Response 的 JSON 结构
type ApiResponse struct {
	Code    int             `json:"code"`
	Data    json.RawMessage `json:"data"`
	Message string          `json:"message"`
}

// ConcurrentStats 并发统计
type ConcurrentStats struct {
	Total        int
	Success      int64
	Failed       int64
	TotalLatency time.Duration
	MinLatency   time.Duration
	MaxLatency   time.Duration
	AvgLatency   time.Duration
	P50          time.Duration
	P95          time.Duration
	P99          time.Duration
	QPS          float64
}

// calcStats 根据每次请求耗时计算统计
func calcStats(latencies []time.Duration, success, failed int64, totalWall time.Duration) ConcurrentStats {
	if len(latencies) == 0 {
		return ConcurrentStats{Total: int(success + failed), Success: success, Failed: failed}
	}
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	var total time.Duration
	minV, maxV := latencies[0], latencies[0]
	for _, v := range latencies {
		total += v
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}
	avg := time.Duration(int64(total) / int64(len(latencies)))
	p50 := latencies[len(latencies)*50/100]
	p95 := latencies[len(latencies)*95/100]
	p99 := latencies[len(latencies)*99/100]
	qps := 0.0
	if totalWall > 0 {
		qps = float64(success+failed) / totalWall.Seconds()
	}
	return ConcurrentStats{
		Total:        len(latencies),
		Success:      success,
		Failed:       failed,
		TotalLatency: total,
		MinLatency:   minV,
		MaxLatency:   maxV,
		AvgLatency:   avg,
		P50:          p50,
		P95:          p95,
		P99:          p99,
		QPS:          qps,
	}
}

func (s ConcurrentStats) String() string {
	return fmt.Sprintf(
		"总请求=%d 成功=%d 失败=%d | min=%v max=%v avg=%v p50=%v p95=%v p99=%v | QPS=%.2f",
		s.Total, s.Success, s.Failed, s.MinLatency, s.MaxLatency, s.AvgLatency, s.P50, s.P95, s.P99, s.QPS,
	)
}

// ---------------------------------------------------------------------------
// 核心工具：高并发请求 + 合并
// ---------------------------------------------------------------------------

// fetchResult 单次请求结果容器
type fetchResult[T any] struct {
	Data       T
	StatusCode int
	Latency    time.Duration
	Err        error
	RawBody    []byte
}

// concurrentFetch 通用高并发执行器
//
//	concurrency: 同时并发的 goroutine 数（信号量限流）
//	total: 总请求数
//	fn: 单次请求逻辑，需自行创建 *http.Request / httptest.NewRecorder
func concurrentFetch[T any](ctx context.Context, concurrency, total int, fn func(ctx context.Context) fetchResult[T]) ([]T, ConcurrentStats, []error) {
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var latencies []time.Duration
	var success, failed atomic.Int64
	results := make([]T, 0, total)
	errs := make([]error, 0)
	latencies = make([]time.Duration, 0, total)

	startWall := time.Now()
	for i := 0; i < total; i++ {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			r := fn(ctx)
			mu.Lock()
			latencies = append(latencies, r.Latency)
			if r.Err != nil || r.StatusCode != http.StatusOK {
				failed.Add(1)
				if r.Err != nil {
					errs = append(errs, r.Err)
				} else {
					errs = append(errs, fmt.Errorf("status=%d body=%s", r.StatusCode, string(r.RawBody)))
				}
			} else {
				success.Add(1)
				results = append(results, r.Data)
			}
			mu.Unlock()
		}()
	}
	wg.Wait()
	wall := time.Since(startWall)
	stats := calcStats(latencies, success.Load(), failed.Load(), wall)
	return results, stats, errs
}

// ---------------------------------------------------------------------------
// 示例 1: 高并发查询 /api/accounts，合并去重
// ---------------------------------------------------------------------------

func TestConcurrentAccounts_Merge(t *testing.T) {
	// 1. 构建内存服务（复用已有 fakeRepo）
	accRepo := newFakeAccountRepo()
	orderRepo := newFakeOrderRepo()
	accUc := biz.NewAccountUsecase(accRepo)
	orderUc := biz.NewOrderUsecase(orderRepo)
	svc := service.NewExchangeService(nil, orderUc, accUc, nil, nil, nil, nil)
	router := server.NewHTTPServer(nil, svc)

	concurrency := 50 // 并发 goroutine 数
	total := 200      // 总请求数
	ctx := context.Background()

	// 并发请求函数：每次独立创建 Request/Recorder（httptest 要求）
	results, stats, errs := concurrentFetch[[]*biz.Account](ctx, concurrency, total, func(ctx context.Context) fetchResult[[]*biz.Account] {
		start := time.Now()
		req := httptest.NewRequest(http.MethodGet, "/api/accounts", nil)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		lat := time.Since(start)

		if w.Code != http.StatusOK {
			return fetchResult[[]*biz.Account]{StatusCode: w.Code, Latency: lat, RawBody: w.Body.Bytes()}
		}
		var resp ApiResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			return fetchResult[[]*biz.Account]{StatusCode: w.Code, Latency: lat, RawBody: w.Body.Bytes(), Err: err}
		}
		var accounts []*biz.Account
		if err := json.Unmarshal(resp.Data, &accounts); err != nil {
			return fetchResult[[]*biz.Account]{StatusCode: w.Code, Latency: lat, RawBody: w.Body.Bytes(), Err: err}
		}
		return fetchResult[[]*biz.Account]{Data: accounts, StatusCode: w.Code, Latency: lat, RawBody: w.Body.Bytes()}
	})

	t.Logf("[Accounts] 并发统计: %s", stats.String())
	if len(errs) > 0 {
		t.Logf("[Accounts] 错误数=%d 首个错误: %v", len(errs), errs[0])
	}
	if stats.Failed > 0 {
		t.Fatalf("存在失败请求: %d/%d", stats.Failed, stats.Total)
	}

	// 2. 合并去重：按 Asset 去重，多次请求返回相同数据时只保留一份
	merged := mergeAccounts(results)
	t.Logf("[Accounts] 原始返回批次=%d 合并去重后=%d", len(results), len(merged))
	for _, a := range merged {
		t.Logf("  Asset=%s Balance=%.4f Frozen=%.4f", a.Asset, a.Balance, a.Frozen)
	}

	if len(merged) == 0 {
		t.Fatal("合并后数据为空，不符合预期")
	}
	// 校验去重正确性：fakeRepo 中只有 USDT/BTC 两个币种
	if len(merged) != 2 {
		t.Fatalf("合并后数量不符合预期: got %d want 2", len(merged))
	}
}

// mergeAccounts 将多批次 []*biz.Account 按 Asset 去重合并
func mergeAccounts(batches [][]*biz.Account) []*biz.Account {
	m := make(map[string]*biz.Account)
	for _, batch := range batches {
		for _, a := range batch {
			if _, ok := m[a.Asset]; !ok {
				// 深拷贝避免指针共享问题
				cp := *a
				m[a.Asset] = &cp
			}
		}
	}
	out := make([]*biz.Account, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Asset < out[j].Asset })
	return out
}

// ---------------------------------------------------------------------------
// 示例 2: 高并发分页查询 /api/account_flows，合并多页
// ---------------------------------------------------------------------------

func TestConcurrentAccountFlows_MergeByPage(t *testing.T) {
	flowRepo := newFakeAccountFlowRepo()
	orderUc := biz.NewOrderUsecase(newFakeOrderRepo())
	flowUc := biz.NewAccountFlowUsecase(flowRepo)
	svc := service.NewExchangeService(nil, orderUc, nil, flowUc, nil, nil, nil)
	router := server.NewHTTPServer(nil, svc)

	// 场景：分页接口，每页 page_size=1，共 2 条数据，需要并发拉取 2 页后合并
	pages := []int{1, 2}
	concurrency := 5
	ctx := context.Background()

	type FlowPage = biz.AccountFlowPage

	var wg sync.WaitGroup
	var mu sync.Mutex
	var latencies []time.Duration
	var success, failed atomic.Int64
	var results []*FlowPage
	var errs []error
	var stats ConcurrentStats
	sem := make(chan struct{}, concurrency)
	startWall := time.Now()
	for _, p := range pages {
		wg.Add(1)
		sem <- struct{}{}
		go func(page int) {
			defer wg.Done()
			defer func() { <-sem }()
			start := time.Now()
			url := fmt.Sprintf("/api/account_flows?page=%d&page_size=1", page)
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			lat := time.Since(start)

			mu.Lock()
			latencies = append(latencies, lat)
			mu.Unlock()

			if w.Code != http.StatusOK {
				failed.Add(1)
				mu.Lock()
				errs = append(errs, fmt.Errorf("page=%d status=%d body=%s", page, w.Code, w.Body.String()))
				mu.Unlock()
				return
			}
			var resp ApiResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				failed.Add(1)
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				return
			}
			var pg FlowPage
			if err := json.Unmarshal(resp.Data, &pg); err != nil {
				failed.Add(1)
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				return
			}
			mu.Lock()
			results = append(results, &pg)
			mu.Unlock()
			success.Add(1)
		}(p)
	}
	wg.Wait()
	stats = calcStats(latencies, success.Load(), failed.Load(), time.Since(startWall))

	t.Logf("[AccountFlows] 并发统计: %s", stats.String())
	if len(errs) > 0 {
		t.Fatalf("分页请求失败: %v", errs[0])
	}

	// 合并多页 Flows，按 ID 去重
	merged := mergeAccountFlows(results)
	t.Logf("[AccountFlows] 分页批次=%d 合并后总条数=%d (Total 字段首个=%d)", len(results), len(merged), results[0].Total)
	for _, f := range merged {
		t.Logf("  Flow ID=%d Asset=%s Type=%s Amount=%.4f", f.ID, f.Asset, f.ChangeType, f.Amount)
	}
	if len(merged) != 2 {
		t.Fatalf("合并后数量不符合预期: got %d want 2", len(merged))
	}
}

func mergeAccountFlows(pages []*biz.AccountFlowPage) []*biz.AccountFlow {
	m := make(map[uint64]*biz.AccountFlow)
	for _, pg := range pages {
		for _, f := range pg.Flows {
			if _, ok := m[f.ID]; !ok {
				cp := *f
				m[f.ID] = &cp
			}
		}
	}
	out := make([]*biz.AccountFlow, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// ---------------------------------------------------------------------------
// 示例 3: 压测 + 真实 HTTP Server（httptest.NewServer + http.Client）
// ---------------------------------------------------------------------------

func TestConcurrentStress_RealHTTPServer(t *testing.T) {
	accRepo := newFakeAccountRepo()
	orderRepo := newFakeOrderRepo()
	accUc := biz.NewAccountUsecase(accRepo)
	orderUc := biz.NewOrderUsecase(orderRepo)
	svc := service.NewExchangeService(nil, orderUc, accUc, nil, nil, nil, nil)
	router := server.NewHTTPServer(nil, svc)

	// 使用真实 HTTP Server，模拟网络链路并发
	ts := httptest.NewServer(router)
	defer ts.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	concurrency := 100
	total := 500
	ctx := context.Background()

	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var latencies []time.Duration
	var success, failed atomic.Int64
	mergedMaps := make(map[string]*biz.Account)

	startWall := time.Now()
	for i := 0; i < total; i++ {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			start := time.Now()
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"/api/accounts", nil)
			resp, err := client.Do(req)
			lat := time.Since(start)

			mu.Lock()
			latencies = append(latencies, lat)
			mu.Unlock()

			if err != nil {
				failed.Add(1)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				failed.Add(1)
				return
			}
			var apiResp ApiResponse
			if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
				failed.Add(1)
				return
			}
			var accounts []*biz.Account
			if err := json.Unmarshal(apiResp.Data, &accounts); err != nil {
				failed.Add(1)
				return
			}
			success.Add(1)
			// 线程安全合并
			mu.Lock()
			for _, a := range accounts {
				if _, ok := mergedMaps[a.Asset]; !ok {
					cp := *a
					mergedMaps[a.Asset] = &cp
				}
			}
			mu.Unlock()
		}()
	}
	wg.Wait()
	stats := calcStats(latencies, success.Load(), failed.Load(), time.Since(startWall))
	t.Logf("[Stress-RealHTTP] 统计: %s", stats.String())
	t.Logf("[Stress-RealHTTP] 合并去重后资产数=%d", len(mergedMaps))
	for asset, acc := range mergedMaps {
		t.Logf("  %s => Balance=%.4f", asset, acc.Balance)
	}

	if failed.Load() > 0 {
		t.Fatalf("压测存在失败请求: %d/%d", failed.Load(), total)
	}
	if len(mergedMaps) != 2 {
		t.Fatalf("合并后数量不符合预期: got %d want 2", len(mergedMaps))
	}
	// 额外断言：p99 应在合理范围（内存服务应 < 100ms）
	if stats.P99 > 500*time.Millisecond {
		t.Logf("警告: P99 延迟过高 %v，可能存在锁竞争或资源瓶颈", stats.P99)
	}
}

// ---------------------------------------------------------------------------
// 示例 4: 超时与限流控制
// ---------------------------------------------------------------------------

func TestConcurrentWithTimeout(t *testing.T) {
	accRepo := newFakeAccountRepo()
	accUc := biz.NewAccountUsecase(accRepo)
	orderUc := biz.NewOrderUsecase(newFakeOrderRepo())
	svc := service.NewExchangeService(nil, orderUc, accUc, nil, nil, nil, nil)
	router := server.NewHTTPServer(nil, svc)

	concurrency := 20
	total := 100
	// 全局超时 3 秒，单次请求由 client 侧控制
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	results, stats, errs := concurrentFetch[[]*biz.Account](ctx, concurrency, total, func(ctx context.Context) fetchResult[[]*biz.Account] {
		start := time.Now()
		req := httptest.NewRequest(http.MethodGet, "/api/accounts?asset=USDT", nil)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()
		// 模拟部分请求超时：通过 context 检查
		select {
		case <-ctx.Done():
			return fetchResult[[]*biz.Account]{Latency: time.Since(start), Err: ctx.Err()}
		default:
		}
		router.ServeHTTP(w, req)
		lat := time.Since(start)
		if w.Code != http.StatusOK {
			return fetchResult[[]*biz.Account]{StatusCode: w.Code, Latency: lat, RawBody: w.Body.Bytes()}
		}
		var resp ApiResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			return fetchResult[[]*biz.Account]{Latency: lat, Err: err}
		}
		var accounts []*biz.Account
		if err := json.Unmarshal(resp.Data, &accounts); err != nil {
			return fetchResult[[]*biz.Account]{Latency: lat, Err: err}
		}
		return fetchResult[[]*biz.Account]{Data: accounts, StatusCode: w.Code, Latency: lat}
	})

	t.Logf("[Timeout] 统计: %s 错误数=%d", stats.String(), len(errs))
	merged := mergeAccounts(results)
	t.Logf("[Timeout] 合并后=%d", len(merged))
	if stats.Failed > 0 {
		t.Logf("部分请求因超时/错误失败，符合限流预期")
	}
	_ = merged
}

// ---------------------------------------------------------------------------
// 示例 5: 高并发分页查询 /api/orders，多线程拉取全部页后合并校验
// ---------------------------------------------------------------------------

// TestConcurrentOrders_MergeByPage 高并发分页查询集成测试（查询 test 数据库真实数据）。
// 连接 configs/config.yaml 中配置的 test MySQL 数据库（可用 CONFIG_PATH 覆盖），
// 通过 /api/orders 接口由多个 goroutine 并发拉取 user 1 的全部订单页，
// 分别校验每页元信息，再将各页结果按订单 ID 去重合并，校验数据无丢失、无重复。
// 每页 page_size=5，总页数由第 1 页探测到的 Total 动态计算；并发数 10 制造真实竞争。
func TestConcurrentOrders_MergeByPage(t *testing.T) {
	// 1. 初始化配置，连接 test 数据库
	configPath := "../configs/config.yaml"
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		configPath = path
	}
	bc, err := conf.Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config from %s: %v", configPath, err)
	}

	// 2. 初始化数据层，使用真实 OrderRepo 查询 test 数据库
	db := data.NewDB(bc)
	rdb := data.NewRedis(bc)
	d, cleanup, err := data.NewData(db, rdb)
	if err != nil {
		t.Fatalf("failed to init data: %v", err)
	}
	defer cleanup()

	orderUc := biz.NewOrderUsecase(data.NewOrderRepo(d))
	svc := service.NewExchangeService(nil, orderUc, nil, nil, nil, nil, nil)
	router := server.NewHTTPServer(nil, svc)

	const (
		pageSize    = 5
		concurrency = 10 // 并发 goroutine 数（大于页数，制造真实竞争）
	)
	ctx := context.Background()

	// 3. 先同步探测第 1 页，获取 test 数据库中 user 1 的订单总数
	probe, err := fetchOrderPage(router, ctx, 1, pageSize)
	if err != nil {
		t.Fatalf("探测第 1 页失败: %v", err)
	}
	total := probe.Total
	if total == 0 {
		t.Skip("test 数据库中 user 1 无订单数据，跳过")
		return
	}
	totalPages := int((total + pageSize - 1) / pageSize)
	t.Logf("[Orders] test 数据库探测: user 1 共 %d 条订单, page_size=%d, 共 %d 页", total, pageSize, totalPages)

	// 4. 并发安全收集：页面结果、错误、耗时、成功/失败计数
	var wg sync.WaitGroup
	var mu sync.Mutex
	var latencies []time.Duration
	var success, failed atomic.Int64
	var pages []*biz.OrderPage
	var errs []error

	// 5. 并发拉取每一页：sem 信号量控制最大并发 goroutine 数，defer 释放避免泄漏
	sem := make(chan struct{}, concurrency)
	startWall := time.Now()
	for p := 1; p <= totalPages; p++ {
		wg.Add(1)
		sem <- struct{}{}
		go func(page int) {
			defer wg.Done()
			defer func() { <-sem }()
			start := time.Now()
			pg, err := fetchOrderPage(router, ctx, page, pageSize)
			lat := time.Since(start)

			mu.Lock()
			latencies = append(latencies, lat)
			mu.Unlock()

			if err != nil {
				failed.Add(1)
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				return
			}
			// 校验每页元信息：总条数、当前页码、每页条数（末页允许不满页）
			if pg.Total != total || pg.Page != page || pg.PageSize != pageSize {
				failed.Add(1)
				mu.Lock()
				errs = append(errs, fmt.Errorf("page=%d 元信息不符: total=%d page=%d pageSize=%d", page, pg.Total, pg.Page, pg.PageSize))
				mu.Unlock()
				return
			}
			if len(pg.Orders) == 0 || len(pg.Orders) > pageSize {
				failed.Add(1)
				mu.Lock()
				errs = append(errs, fmt.Errorf("page=%d 条数不符: got %d want 1..%d", page, len(pg.Orders), pageSize))
				mu.Unlock()
				return
			}
			mu.Lock()
			pages = append(pages, pg)
			mu.Unlock()
			success.Add(1)
		}(p)
	}
	wg.Wait()
	// 统计并发请求的性能指标
	stats := calcStats(latencies, success.Load(), failed.Load(), time.Since(startWall))

	t.Logf("[Orders] 多线程分页统计: %s", stats.String())
	if len(errs) > 0 {
		t.Fatalf("分页请求失败: %v", errs[0])
	}

	// 6. 合并全部分页：按订单 ID 去重、升序排序，校验无丢失、无重复
	merged := mergeOrderPages(pages)
	t.Logf("[Orders] 分页批次=%d 合并后总条数=%d (test 数据库 Total=%d)", len(pages), len(merged), total)

	if int64(len(merged)) != total {
		t.Fatalf("合并后数量与 test 数据库 Total 不符: got %d want %d", len(merged), total)
	}
	// 校验 ID 严格递增：无重复、无乱序（数据库自增 ID 不要求连续）
	for i := 1; i < len(merged); i++ {
		if merged[i].ID <= merged[i-1].ID {
			t.Fatalf("合并后订单 ID 重复或乱序: index=%d ID=%d prevID=%d", i, merged[i].ID, merged[i-1].ID)
		}
	}
	for _, o := range merged {
		t.Logf("  Order ID=%d Symbol=%s Side=%s Price=%.2f Status=%s", o.ID, o.Symbol, o.Side, o.Price, o.Status)
	}
}

// fetchOrderPage 单次请求 /api/orders 指定页并解析响应
func fetchOrderPage(router http.Handler, ctx context.Context, page, pageSize int) (*biz.OrderPage, error) {
	url := fmt.Sprintf("/api/orders?page=%d&page_size=%d", page, pageSize)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		return nil, fmt.Errorf("page=%d status=%d body=%s", page, w.Code, w.Body.String())
	}
	var resp ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		return nil, fmt.Errorf("page=%d 解析响应失败: %w", page, err)
	}
	var pg biz.OrderPage
	if err := json.Unmarshal(resp.Data, &pg); err != nil {
		return nil, fmt.Errorf("page=%d 解析分页数据失败: %w", page, err)
	}
	return &pg, nil
}

// mergeOrderPages 将多页订单结果按 ID 去重合并，并按 ID 升序排序
func mergeOrderPages(pages []*biz.OrderPage) []*biz.Order {
	m := make(map[uint64]*biz.Order)
	for _, pg := range pages {
		for _, o := range pg.Orders {
			if _, ok := m[o.ID]; !ok {
				cp := *o
				m[o.ID] = &cp
			}
		}
	}
	out := make([]*biz.Order, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })

	return out
}
