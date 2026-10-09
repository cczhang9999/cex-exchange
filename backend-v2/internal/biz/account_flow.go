package biz

import (
	"context"
	"time"
)

// AccountFlow 定义用户资金流水业务模型
type AccountFlow struct {
	ID         uint64    `json:"id"`
	UserID     uint64    `json:"user_id"`
	AccountID  uint64    `json:"account_id"`
	Asset      string    `json:"asset"`
	ChangeType string    `json:"change_type"`
	Amount     float64   `json:"amount"`
	Balance    float64   `json:"balance"`
	RefID      *uint64   `json:"ref_id,omitempty"`
	Remark     *string   `json:"remark,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// AccountFlowQuery 定义流水查询条件
type AccountFlowQuery struct {
	UserID     uint64
	Asset      string
	ChangeType string
	StartTime  time.Time
	EndTime    time.Time
	Page       int
	PageSize   int
}

// AccountFlowPage 定义流水分页查询结果
type AccountFlowPage struct {
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	Flows    []*AccountFlow `json:"flows"`
}

// AccountFlowRepo 定义流水存储接口
type AccountFlowRepo interface {
	// FindPageByUserID 按条件分页查询用户的资金流水
	FindPageByUserID(ctx context.Context, q AccountFlowQuery) (*AccountFlowPage, error)
	// Create 记录资金流水
	Create(ctx context.Context, flow *AccountFlow) (*AccountFlow, error)
}

// AccountFlowUsecase 定义资金流水业务逻辑
type AccountFlowUsecase struct {
	repo AccountFlowRepo
}

// NewAccountFlowUsecase 创建资金流水业务逻辑实例
func NewAccountFlowUsecase(repo AccountFlowRepo) *AccountFlowUsecase {
	return &AccountFlowUsecase{
		repo: repo,
	}
}

// GetAccountFlows 分页查询用户资金流水列表
func (uc *AccountFlowUsecase) GetAccountFlows(ctx context.Context, q AccountFlowQuery) (*AccountFlowPage, error) {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 20
	}
	return uc.repo.FindPageByUserID(ctx, q)
}
