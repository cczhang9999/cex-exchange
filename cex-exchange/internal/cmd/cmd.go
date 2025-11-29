package cmd

import (
	"cex-exchange/internal/controller/trade"
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	_ "cex-exchange/internal/boot"
	"cex-exchange/internal/controller/admin"
	"cex-exchange/internal/controller/auth"
	"cex-exchange/internal/controller/funds"
	"cex-exchange/internal/controller/order"
	"cex-exchange/internal/middleware"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			s := g.Server()

			// 全局中间件
			s.Use(middleware.CORS)
			s.Use(middleware.ResponseHandler)

			// 公开路由
			s.Group("/api", func(group *ghttp.RouterGroup) {
				group.Bind(
					auth.NewV1(),
				)
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
