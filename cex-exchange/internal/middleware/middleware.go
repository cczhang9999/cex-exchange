package middleware

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// CORS 跨域中间件
func CORS(r *ghttp.Request) {
	r.Response.CORSDefault()
	r.Middleware.Next()
}

// ResponseHandler 统一响应处理
func ResponseHandler(r *ghttp.Request) {
	r.Middleware.Next()
	
	// 如果已经有错误，不再处理
	if r.Response.BufferLength() > 0 {
		return
	}
	
	var (
		err = r.GetError()
		res = r.GetHandlerResponse()
	)
	
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    1,
			"message": err.Error(),
			"data":    nil,
		})
		r.Response.WriteStatus(400)
	} else if res != nil {
		r.Response.WriteJson(g.Map{
			"code": 0,
			"message": "success",
			"data": res,
		})
	}
}
