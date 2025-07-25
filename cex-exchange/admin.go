package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PageQuery struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
}

type AddFundsRequest struct {
	UserIDStr string `json:"user_id" binding:"required"`
	Asset     string `json:"asset" binding:"required"`
	Amount    string `json:"amount" binding:"required"`
	Remark    string `json:"remark" binding:"required"`
}

// 用户列表
func AdminListUsers(c *gin.Context) {
	var q PageQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		q.Page, q.Limit = 1, 20
	}
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 || q.Limit > 100 {
		q.Limit = 20
	}
	var users []User
	DB.Offset((q.Page - 1) * q.Limit).Limit(q.Limit).Order("id desc").Find(&users)
	c.JSON(http.StatusOK, users)
}

// 订单列表
func AdminListOrders(c *gin.Context) {
	var q PageQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		q.Page, q.Limit = 1, 20
	}
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 || q.Limit > 100 {
		q.Limit = 20
	}
	var orders []Order
	DB.Offset((q.Page - 1) * q.Limit).Limit(q.Limit).Order("id desc").Find(&orders)
	c.JSON(http.StatusOK, orders)
}

// 资金账户列表
func AdminListAccounts(c *gin.Context) {
	var q PageQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		q.Page, q.Limit = 1, 20
	}
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 || q.Limit > 100 {
		q.Limit = 20
	}
	var accounts []Account
	DB.Offset((q.Page - 1) * q.Limit).Limit(q.Limit).Order("id desc").Find(&accounts)
	c.JSON(http.StatusOK, accounts)
}

// 管理员添加资金
func AdminAddFunds(c *gin.Context) {
	var req AddFundsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 验证金额格式
	amount, err := strconv.ParseFloat(req.Amount, 64)
	// 转换用户ID
	userID, err := strconv.ParseUint(req.UserIDStr, 10, 64)
	if err != nil || amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "金额格式错误"})
		return
	}

	// 验证用户是否存在
	var user User
	if err := DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户不存在"})
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
	if err := tx.Where("user_id = ? AND asset = ?", userID, req.Asset).First(&account).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建新账户
			account = Account{
				UserID:  userID,
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
		UserID:     userID,
		AccountID:  account.ID,
		Asset:      req.Asset,
		ChangeType: "admin_add",
		Amount:     req.Amount,
		Balance:    account.Balance,
		Remark:     fmt.Sprintf("管理员添加资金: %s", req.Remark),
	}
	if err := tx.Create(&flow).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建流水失败"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{
		"message": "资金添加成功",
		"account": account,
	})
}
