package biz

import (
	"backend-v2/internal/conf"
	"backend-v2/internal/pkg/jwt"
	"context"
	"fmt"
	"time"
)

// UserStatus 用户状态
type UserStatus int

const (
	UserStatusNormal  UserStatus = 1 // 正常
	UserStatusBlocked UserStatus = 0 // 禁用/封禁
)

type User struct {
	ID        uint64     `json:"id"`
	Username  string     `json:"username"`
	Password  string     `json:"-"`
	Email     string     `json:"email"`
	Phone     string     `json:"phone"`
	Status    UserStatus `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// IsBlocked 用户是否已被封禁
func (u *User) IsBlocked() bool {
	return u.Status == UserStatusBlocked
}

type UserRepo interface {
	Save(ctx context.Context, user *User) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	ValidatePassword(user *User, password string) bool
	FindUserList(ctx context.Context) ([]*User, error)
	FindByID(ctx context.Context, id uint64) (*User, error)
	UpdateStatus(ctx context.Context, id uint64, status UserStatus) error
}

type UserUsecase struct {
	repo   UserRepo
	secret string
	expiry time.Duration
}

func NewUserUsecase(repo UserRepo, conf *conf.Bootstrap) *UserUsecase {
	expiry, _ := time.ParseDuration(conf.Auth.JwtExpiry)
	return &UserUsecase{
		repo:   repo,
		secret: conf.Auth.JwtSecret,
		expiry: expiry,
	}
}

func (uc *UserUsecase) Register(ctx context.Context, username, password, email string, phone string) (*User, error) {
	// TODO: Check if user exists
	// TODO: Hash password
	u := &User{
		Username: username,
		Password: password, // In real world this should be hashed
		Email:    email,
		Phone:    phone,
		Status:   UserStatusNormal,
	}
	return uc.repo.Save(ctx, u)
}

// Login 登录：查询用户、校验密码、拒绝已封禁用户
func (uc *UserUsecase) Login(ctx context.Context, username, password string) (string, uint64, error) {
	u, err := uc.repo.FindByUsername(ctx, username)
	if err != nil {
		return "", 0, err
	}
	if !uc.repo.ValidatePassword(u, password) {
		return "", 0, nil // Invalid password
	}
	if u.Status == UserStatusBlocked {
		return "", 0, fmt.Errorf("账户已被封禁")
	}
	token, err := jwt.GenerateToken(u.ID, uc.secret, uc.expiry)
	if err != nil {
		return "", 0, err
	}
	return token, u.ID, nil
}

func (uc *UserUsecase) ListUsers(ctx context.Context) ([]*User, error) {
	return uc.repo.FindUserList(ctx)
}

// GetUser 获取单个用户
func (uc *UserUsecase) GetUser(ctx context.Context, id uint64) (*User, error) {
	return uc.repo.FindByID(ctx, id)
}

// BlockUser 封禁/解封用户
func (uc *UserUsecase) BlockUser(ctx context.Context, id uint64, block bool) error {
	status := UserStatusNormal
	if block {
		status = UserStatusBlocked
	}
	return uc.repo.UpdateStatus(ctx, id, status)
}
