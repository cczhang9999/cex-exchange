package main

import (
	"errors"
	"fmt"
)

// 示例 2: 接口和依赖注入
// 这是微服务中常用的接口设计模式

// UserDao 接口定义
type UserDao interface {
	GetByID(id uint64) (*User, error)
	GetByUsername(username string) (*User, error)
	Create(data map[string]interface{}) (uint64, error)
}

// User 结构体
type User struct {
	ID       uint64
	Username string
	Email    string
}

// MockUserDao 模拟的实现
type MockUserDao struct {
	users map[uint64]*User
}

func NewMockUserDao() *MockUserDao {
	return &MockUserDao{
		users: make(map[uint64]*User),
	}
}

func (m *MockUserDao) GetByID(id uint64) (*User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *MockUserDao) GetByUsername(username string) (*User, error) {
	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (m *MockUserDao) Create(data map[string]interface{}) (uint64, error) {
	// 模拟创建用户
	id := uint64(len(m.users) + 1)
	user := &User{
		ID:       id,
		Username: data["username"].(string),
		Email:    data["email"].(string),
	}
	m.users[id] = user
	return id, nil
}

// UserService 依赖接口而非具体实现
type UserService struct {
	dao UserDao // 通过接口注入依赖
}

func NewUserService(dao UserDao) *UserService {
	return &UserService{dao: dao}
}

func (s *UserService) GetUser(id uint64) (*User, error) {
	return s.dao.GetByID(id)
}

func (s *UserService) Register(username, email string) (uint64, error) {
	// 业务逻辑验证
	if len(username) < 3 {
		return 0, errors.New("username too short")
	}
	return s.dao.Create(map[string]interface{}{
		"username": username,
		"email":    email,
	})
}

func main() {
	fmt.Println("=== Go 学习示例 2: 接口和依赖注入 ===")

	// 创建 Mock 实现
	mockDao := NewMockUserDao()

	// 通过接口注入
	service := NewUserService(mockDao)

	// 注册用户
	id, err := service.Register("alice", "alice@example.com")
	if err != nil {
		fmt.Printf("注册失败: %v\n", err)
		return
	}
	fmt.Printf("用户注册成功，ID: %d\n", id)

	// 获取用户
	user, err := service.GetUser(id)
	if err != nil {
		fmt.Printf("获取用户失败: %v\n", err)
		return
	}
	fmt.Printf("获取到用户: %+v\n", user)
}
