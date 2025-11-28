package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DepositRequest struct {
	Asset  string `json:"asset" binding:"required"`
	Amount string `json:"amount" binding:"required"`
}

type WithdrawRequest struct {
	Asset  string `json:"asset" binding:"required"`
	Amount string `json:"amount" binding:"required"`
}

// 充值接口
func Deposit(c *gin.Context) {
	uid := c.GetUint64("uid")
	fmt.Printf("Deposit request from user %d\n", uid)

	var req DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("Bind JSON error: %v\n", err)

		fmt.Printf("Deposit request: %+v\n", req)
		return
	}

	fmt.Printf("Deposit request: %+v\n", req)
	// 验证金额格式
	amount, err := strconv.ParseFloat(req.Amount, 64)
	if err != nil || amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "金额格式错误"})
		return
	}

	// 使用事务处理
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var account Account
	if err := tx.Where("user_id = ? AND asset = ?", uid, req.Asset).First(&account).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建新账户
			account = Account{
				UserID:  uid,
				Asset:   req.Asset,
				Balance: req.Amount,
				Frozen:  "0",
			}
			if err := tx.Create(&account).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "创建账户失败"})
				return
			}
		} else {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询账户失败"})
			return
		}
	} else {
		// 更新现有账户余额
		currentBalance, _ := strconv.ParseFloat(account.Balance, 64)
		newBalance := currentBalance + amount
		if err := tx.Model(&account).Update("balance", strconv.FormatFloat(newBalance, 'f', -1, 64)).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新余额失败"})
			return
		}
		account.Balance = strconv.FormatFloat(newBalance, 'f', -1, 64)
	}

	// 创建流水记录
	flow := AccountFlow{
		UserID:     uid,
		AccountID:  account.ID,
		Asset:      req.Asset,
		ChangeType: "deposit",
		Amount:     req.Amount,
		Balance:    account.Balance,
		Remark:     "充值",
	}
	if err := tx.Create(&flow).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建流水失败"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "充值成功"})
}

// 提现接口
func Withdraw(c *gin.Context) {
	uid := c.GetUint64("uid")
	var req WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 验证金额格式
	amount, err := strconv.ParseFloat(req.Amount, 64)
	if err != nil || amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "金额格式错误"})
		return
	}

	// 使用事务处理
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var account Account
	if err := tx.Where("user_id = ? AND asset = ?", uid, req.Asset).First(&account).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "账户不存在"})
		return
	}

	// 检查余额是否充足
	balance, err := strconv.ParseFloat(account.Balance, 64)
	if err != nil || balance < amount {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "余额不足"})
		return
	}

	// 更新余额
	newBalance := balance - amount
	if err := tx.Model(&account).Update("balance", strconv.FormatFloat(newBalance, 'f', -1, 64)).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新余额失败"})
		return
	}

	// 创建流水记录
	flow := AccountFlow{
		UserID:     uid,
		AccountID:  account.ID,
		Asset:      req.Asset,
		ChangeType: "withdraw",
		Amount:     req.Amount,
		Balance:    strconv.FormatFloat(newBalance, 'f', -1, 64),
		Remark:     "提现",
	}
	if err := tx.Create(&flow).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建流水失败"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "提现成功"})
}

// 资金流水查询
func ListAccountFlows(c *gin.Context) {
	uid := c.GetUint64("uid")
	var flows []AccountFlow
	DB.Where("user_id = ?", uid).Order("id desc").Limit(50).Find(&flows)
	c.JSON(http.StatusOK, flows)
}

// 资产账户列表
func ListAccounts(c *gin.Context) {
	uid := c.GetUint64("uid")
	var accounts []Account
	DB.Where("user_id = ?", uid).Find(&accounts)
	c.JSON(200, accounts)
}
