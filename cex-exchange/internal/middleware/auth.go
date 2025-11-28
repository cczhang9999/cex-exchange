package middleware

import (
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/golang-jwt/jwt/v4"
)

// Auth JWT 认证中间件
func Auth(r *ghttp.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		r.Response.WriteJson(g.Map{
			"code": 401,
			"message": "未提供认证令牌",
		})
		r.ExitAll()
		return
	}
	
	// 提取 token
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		r.Response.WriteJson(g.Map{
			"code": 401,
			"message": "认证令牌格式错误",
		})
		r.ExitAll()
		return
	}
	
	tokenString := parts[1]
	secret := g.Cfg().MustGet(r.Context(), "jwt.secret").String()
	
	// 解析 token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	
	if err != nil || !token.Valid {
		g.Log().Errorf(r.Context(), "JWT Parse Error: %v, Token Valid: %v", err, token.Valid)
		r.Response.WriteJson(g.Map{
			"code": 401,
			"message": "认证令牌无效或已过期",
		})
		r.ExitAll()
		return
	}
	
	// 提取用户 ID
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if uid, ok := claims["uid"].(float64); ok {
			r.SetCtxVar("uid", uint64(uid))
		}
	}
	
	r.Middleware.Next()
}
