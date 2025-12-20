package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	_ "cex-exchange/internal/boot"
	"cex-exchange/internal/controller/admin"
	"cex-exchange/internal/controller/auth"
	"cex-exchange/internal/controller/funds"
	"cex-exchange/internal/controller/order"
	"cex-exchange/internal/controller/trade"
	wsController "cex-exchange/internal/controller/websocket"
	"cex-exchange/internal/middleware"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			s := g.Server()

			// 初始化WebSocket管理器
			//ws.InitManager()
			//g.Log().Info(ctx, "WebSocket管理器已启动")

			// 启动行情数据推送
			//go ws.StartMarketDataPusher()
			//g.Log().Info(ctx, "行情数据推送已启动")

			// 全局中间件
			s.Use(middleware.CORS)
			s.Use(middleware.ResponseHandler)

			// 公开路由
			s.Group("/api", func(group *ghttp.RouterGroup) {
				group.Bind(
					auth.NewV1(),
				)
			})

			// WebSocket路由（公开访问）
			wsCtrl := &wsController.Controller{}
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.GET("/ws", wsCtrl.HandleWebSocket)
				group.GET("/ws/status", wsCtrl.GetStatus)
			})

			// 需要认证的路由
			s.Group("/api", func(group *ghttp.RouterGroup) {
				group.Middleware(middleware.Auth)
				group.Bind(
					order.NewV1(),
					funds.NewV1(),
					admin.NewV1(),
					trade.NewV1(),
				)
			})

			s.Run()
			return nil
		},
	}
)
