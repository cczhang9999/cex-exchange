package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type PageQuery struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
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
