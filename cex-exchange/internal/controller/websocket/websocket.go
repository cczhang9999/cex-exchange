package websocket

import (
	"net/http"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gorilla/websocket"

	ws "cex-exchange/internal/websocket"
)

// Controller WebSocket控制器
type Controller struct{}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源，生产环境应该限制
	},
}

// HandleWebSocket 处理WebSocket连接
func (c *Controller) HandleWebSocket(req *ghttp.Request) {
	ctx := req.Context()
	// 升级HTTP连接为WebSocket
	conn, err := upgrader.Upgrade(req.Response.Writer, req.Request, nil)
	if err != nil {
		g.Log().Error(ctx, "WebSocket升级失败:", err)
		req.Response.WriteJson(g.Map{
			"code":    500,
			"message": "WebSocket升级失败",
		})
		return
	}

	// 处理WebSocket连接
	ws.WSManager.HandleConnection(ctx, conn)
}

// GetStatus 获取WebSocket状态
func (c *Controller) GetStatus(req *ghttp.Request) {
	req.Response.WriteJson(g.Map{
		"code": 200,
		"data": g.Map{
			"connected_clients": ws.WSManager.GetClientCount(),
			"status":            "running",
		},
	})
}
