package data

import (
	"backend-v2/internal/biz"
	"context"

	"github.com/google/wire"
	"gorm.io/gorm"
)

var UserProviderSet = wire.NewSet(NewUserRepo)

type userRepo struct {
	data *Data
}

func NewUserRepo(data *Data) biz.UserRepo {
	return &userRepo{
		data: data,
	}
}

// User GORM Model
type User struct {
	gorm.Model
	Username string
	Password string
	Email    string
	Phone    string
}

func (r *userRepo) Save(ctx context.Context, u *biz.User) (*biz.User, error) {
	user := &User{
		Username: u.Username,
		Password: u.Password,
		Email:    u.Email,
		Phone:    u.Phone,
	}
	if err := r.data.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, err
	}
	u.ID = uint64(user.ID)
	return u, nil
}

func (r *userRepo) FindByUsername(ctx context.Context, username string) (*biz.User, error) {
	var user User
	if err := r.data.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &biz.User{
		ID:       uint64(user.ID),
		Username: user.Username,
		Password: user.Password,
		Email:    user.Email,
		Phone:    user.Phone,
	}, nil
}

func (r *userRepo) ValidatePassword(u *biz.User, password string) bool {
	return u.Password == password // Simple check for now
}
