package dao

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	"cex-exchange/internal/model"
)

// UserDao 用户数据访问对象
type UserDao struct {
	table string
	db    gdb.DB
}

var User = &UserDao{
	table: "users",
}

// Model 获取模型
func (d *UserDao) Model(ctx context.Context) *gdb.Model {
	return g.Model(d.table).Ctx(ctx)
}

// GetByID 根据ID获取用户
func (d *UserDao) GetByID(ctx context.Context, id uint64) (*model.User, error) {
	var user model.User
	err := d.Model(ctx).Where("id", id).Scan(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (d *UserDao) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := d.Model(ctx).Where("username", username).Scan(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create 创建用户
func (d *UserDao) Create(ctx context.Context, data g.Map) (uint64, error) {
	result, err := d.Model(ctx).Data(data).Insert()
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return uint64(id), nil
}

// List 获取用户列表
func (d *UserDao) List(ctx context.Context, page, limit int) ([]model.User, int, error) {
	var users []model.User
	
	// 获取总数
	total, err := d.Model(ctx).Count()
	if err != nil {
		return nil, 0, err
	}
	
	// 分页查询
	err = d.Model(ctx).
		Page(page, limit).
		Order("id desc").
		Scan(&users)
	
	if err != nil {
		return nil, 0, err
	}
	
	return users, total, nil
}

// Exists 检查用户名是否存在
func (d *UserDao) Exists(ctx context.Context, username string) (bool, error) {
	count, err := d.Model(ctx).Where("username", username).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
