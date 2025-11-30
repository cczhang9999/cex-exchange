package boot

import (
	"cex-exchange/internal/service"

	_ "cex-exchange/internal/service"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/gctx"
)

func init() {
	var ctx = gctx.New()

	// 显式设置配置文件路径
	configPath := "manifest/config"
	adapter := g.Cfg().GetAdapter().(*gcfg.AdapterFile)
	adapter.SetPath(configPath)
	adapter.SetFileName("config.yaml") // 强制指定文件名
	adapter.Clear() // 清除缓存
	g.Log().Infof(ctx, "Config path set to: %s", configPath)

	// 测试数据库连接
	db := g.DB()
	if db == nil {
		g.Log().Fatal(ctx, "Database connection failed: db is nil")
		return
	}

	// Ping 数据库
	if err := db.PingMaster(); err != nil {
		g.Log().Fatalf(ctx, "❌ Database ping failed: %v", err)
	} else {
		g.Log().Info(ctx, "✅ Database connected successfully!")
	}

	// 初始化 Redis
	service.Redis.Init()
}
