package biz

import (
	"backend-v2/internal/conf"
	"backend-v2/internal/pkg/jwt"
	"context"
	"time"
)

type User struct {
	ID        uint64
	Username  string
	Password  string
	Email     string
	Phone     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserRepo interface {
	Save(ctx context.Context, user *User) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	ValidatePassword(user *User, password string) bool
	FindUserList(ctx context.Context) ([]*User, error)
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
	}
	return uc.repo.Save(ctx, u)
}

func (uc *UserUsecase) Login(ctx context.Context, username, password string) (string, uint64, error) {
	u, err := uc.repo.FindByUsername(ctx, username)
	if err != nil {
		return "", 0, err
	}
	if !uc.repo.ValidatePassword(u, password) {
		return "", 0, nil // Invalid password
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
