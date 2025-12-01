package service

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"

	"cex-exchange/internal/dao"
)

type IAuth interface {
	Register(ctx context.Context, username, password, email string) (uint64, error)
	Login(ctx context.Context, username, password string) (string, uint64, error)
	GenerateToken(ctx context.Context, uid uint64) (string, error)
	ValidateToken(ctx context.Context, token string) (uint64, error)
}

type authImpl struct{}

var Auth = &authImpl{}

// Register 用户注册
func (s *authImpl) Register(ctx context.Context, username, password, email string) (uint64, error) {
	// 检查用户名是否存在
	exists, err := dao.User.Exists(ctx, username)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, gerror.New("用户名已存在")
	}
	
	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	
	// 创建用户
	uid, err := dao.User.Create(ctx, g.Map{
		"username": username,
		"password": string(hashedPassword),
		"email":    email,
	})
	
	if err != nil {
		return 0, err
	}
	
	return uid, nil
}

// Login 用户登录
func (s *authImpl) Login(ctx context.Context, username, password string) (string, uint64, error) {
	// 查询用户
	user, err := dao.User.GetByUsername(ctx, username)
	if err != nil {
		return "", 0, gerror.New("用户名或密码错误")
	}
	
	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", 0, gerror.New("用户名或密码错误")
	}
	
	// 生成 token
	token, err := s.GenerateToken(ctx, user.ID)
	if err != nil {
		return "", 0, err
	}
	
	return token, user.ID, nil
}

// GenerateToken 生成 JWT token
func (s *authImpl) GenerateToken(ctx context.Context, uid uint64) (string, error) {
	secret := g.Cfg().MustGet(ctx, "jwt.secret").String()
	expires := g.Cfg().MustGet(ctx, "jwt.expires").Int()
	
	claims := jwt.MapClaims{
		"uid": uid,
		"exp": time.Now().Add(time.Duration(expires) * time.Second).Unix(),
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateToken 验证 JWT token 并返回用户 ID
func (s *authImpl) ValidateToken(ctx context.Context, tokenString string) (uint64, error) {
	secret := g.Cfg().MustGet(ctx, "jwt.secret").String()
	
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	
	if err != nil {
		return 0, gerror.New("token 无效")
	}
	
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		uid := uint64(claims["uid"].(float64))
		return uid, nil
	}
	
	return 0, gerror.New("token 无效")
}
