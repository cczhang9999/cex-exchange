package model

import "github.com/gogf/gf/v2/frame/g"

type PlaceOrderBookRes struct {
	OrderBids []Order `json:"bids"`
	OrderAsks []Order `json:"asks"`
}

type PlaceOrderReq struct {
	g.Meta `path:"/order" method:"post" tags:"Order" summary:"下单"`
	Symbol string `json:"symbol" v:"required#请输入交易对"`
	Side   string `json:"side" v:"required|in:buy,sell#请选择方向|方向只能是buy或sell"`
	Type   string `json:"type" v:"required|in:limit,market#请选择类型|类型只能是limit或market"`
	Price  string `json:"price"`
	Amount string `json:"amount" v:"required#请输入数量"`
}
