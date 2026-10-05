package data

import (
	"backend-v2/internal/biz"
	"context"

	"github.com/google/wire"
	"gorm.io/gorm"
)

var AccountFlowProviderSet = wire.NewSet(NewAccountFlowRepo)

// AccountFlow GORM Model
// 对应表结构：
// id bigint unsigned AUTO_INCREMENT 流水ID
// user_id bigint unsigned 用户ID
// account_id bigint unsigned 账户ID
// asset varchar(32) 币种
// change_type varchar(32) 变动类型
// amount decimal(32,16) 变动金额
// balance decimal(32,16) 变动后余额
// ref_id bigint unsigned 关联业务ID
// remark varchar(255) 备注
// created_at / updated_at / deleted_at
type AccountFlow struct {
	gorm.Model
	UserID     uint64
	AccountID  uint64
	Asset      string
	ChangeType string
	Amount     float64
	Balance    float64
	RefID      *uint64
	Remark     *string
}

type accountFlowRepo struct {
	data *Data
}

func NewAccountFlowRepo(data *Data) biz.AccountFlowRepo {
	return &accountFlowRepo{
		data: data,
	}
}

// FindPageByUserID 按条件分页查询用户的资金流水
func (r *accountFlowRepo) FindPageByUserID(ctx context.Context, q biz.AccountFlowQuery) (*biz.AccountFlowPage, error) {
	tx := r.data.db.WithContext(ctx).Model(&AccountFlow{}).Where("user_id = ?", q.UserID)
	if q.Asset != "" {
		tx = tx.Where("asset = ?", q.Asset)
	}
	if q.ChangeType != "" {
		tx = tx.Where("change_type = ?", q.ChangeType)
	}
	if !q.StartTime.IsZero() {
		tx = tx.Where("created_at >= ?", q.StartTime)
	}
	if !q.EndTime.IsZero() {
		tx = tx.Where("created_at <= ?", q.EndTime)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, err
	}

	var flows []AccountFlow
	if err := tx.
		Order("id DESC").
		Offset((q.Page - 1) * q.PageSize).
		Limit(q.PageSize).
		Find(&flows).Error; err != nil {
		return nil, err
	}

	page := &biz.AccountFlowPage{
		Total:    total,
		Page:     q.Page,
		PageSize: q.PageSize,
		Flows:    make([]*biz.AccountFlow, 0, len(flows)),
	}
	for _, f := range flows {
		page.Flows = append(page.Flows, &biz.AccountFlow{
			ID:         uint64(f.ID),
			UserID:     f.UserID,
			AccountID:  f.AccountID,
			Asset:      f.Asset,
			ChangeType: f.ChangeType,
			Amount:     f.Amount,
			Balance:    f.Balance,
			RefID:      f.RefID,
			Remark:     f.Remark,
			CreatedAt:  f.CreatedAt,
			UpdatedAt:  f.UpdatedAt,
		})
	}
	return page, nil
}
