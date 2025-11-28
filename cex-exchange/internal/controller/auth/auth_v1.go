package auth

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type ControllerV1 struct{}

func NewV1() *ControllerV1 {
	return &ControllerV1{}
}

type RegisterReq struct {
	g.Meta   `path:"/register" method:"post" tags:"Auth" summary:"用户注册"`
	Username string `json:"username" v:"required|length:3,20#请输入用户名|用户名长度为3-20个字符"`
	Password string `json:"password" v:"required|length:6,20#请输入密码|密码长度为6-20个字符"`
	Email    string `json:"email" v:"required|email#请输入邮箱|邮箱格式不正确"`
}

type RegisterRes struct {
	Token string `json:"token"`
}

func (c *ControllerV1) Register(ctx context.Context, req *RegisterReq) (res *RegisterRes, err error) {
	// 检查用户名是否存在
	count, err := g.Model("users").Ctx(ctx).Where("username", req.Username).Count()
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, gerror.New("用户名已存在")
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 插入用户
	result, err := g.Model("users").Ctx(ctx).Data(g.Map{
		"username": req.Username,
		"password": string(hashedPassword),
		"email":    req.Email,
	}).Insert()

	if err != nil {
		return nil, err
	}

	uid, _ := result.LastInsertId()

	// 生成 token
	token, err := generateToken(uint64(uid))
	if err != nil {
		return nil, err
	}

	return &RegisterRes{Token: token}, nil
}

type LoginReq struct {
	g.Meta   `path:"/login" method:"post" tags:"Auth" summary:"用户登录"`
	Username string `json:"username" v:"required#请输入用户名"`
	Password string `json:"password" v:"required#请输入密码"`
}

type LoginRes struct {
	Token string `json:"token"`
}

func (c *ControllerV1) Login(ctx context.Context, req *LoginReq) (res *LoginRes, err error) {
	// 查询用户
	var user struct {
		ID            uint64 `json:"id"`
		Password_hash string `json:"password"`
	}

	err = g.Model("users").Ctx(ctx).
		Where("username", req.Username).
		Scan(&user)

	if err != nil {
		g.Log().Errorf(ctx, "Login failed: query error: %v", err)
		return nil, gerror.New("系统错误")
	}
	if user.ID == 0 {
		g.Log().Errorf(ctx, "Login failed: user not found: %s", req.Username)
		return nil, gerror.New("用户不存在")
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password_hash), []byte(req.Password))
	if err != nil {
		g.Log().Errorf(ctx, "Login failed: password mismatch for user: %s", req.Username)
		return nil, gerror.New("密码错误")
	}

	// 生成 token
	token, err := generateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &LoginRes{Token: token}, nil
}

// generateToken 生成 JWT token
func generateToken(uid uint64) (string, error) {
	secret := g.Cfg().MustGet(context.Background(), "jwt.secret").String()
	expires := g.Cfg().MustGet(context.Background(), "jwt.expires").Int()

	claims := jwt.MapClaims{
		"uid": uid,
		"exp": time.Now().Add(time.Duration(expires) * time.Second).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
