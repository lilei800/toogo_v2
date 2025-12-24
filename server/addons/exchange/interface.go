// Package exchange
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE
// 交易所统一接口定义

package exchange

import (
	"context"
	"time"
)

// IExchange 交易所接口（统一抽象）
type IExchange interface {
	// GetName 获取交易所名称
	GetName() string

	// TestConnection 测试连接
	TestConnection(ctx context.Context) (balance string, err error)

	// GetTicker 获取行情
	GetTicker(ctx context.Context, symbol string) (ticker *Ticker, err error)

	// GetKline 获取K线
	GetKline(ctx context.Context, symbol string, interval string, limit int) (klines []*Kline, err error)

	// PlaceOrder 下单
	PlaceOrder(ctx context.Context, order *OrderRequest) (orderId string, err error)

	// CloseOrder 平仓
	CloseOrder(ctx context.Context, orderId string) error

	// GetOrder 获取订单
	GetOrder(ctx context.Context, orderId string) (order *OrderInfo, err error)

	// GetPositions 获取持仓
	GetPositions(ctx context.Context, symbol string) (positions []*Position, err error)
}

// Ticker 行情数据
type Ticker struct {
	Symbol    string    `json:"symbol"`
	LastPrice float64   `json:"lastPrice"`
	High24h   float64   `json:"high24h"`
	Low24h    float64   `json:"low24h"`
	Volume24h float64   `json:"volume24h"`
	Change24h float64   `json:"change24h"`
	Timestamp time.Time `json:"timestamp"`
}

// Kline K线数据
type Kline struct {
	OpenTime  time.Time `json:"openTime"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    float64   `json:"volume"`
	CloseTime time.Time `json:"closeTime"`
}

// OrderRequest 下单请求
type OrderRequest struct {
	Symbol     string  `json:"symbol"`
	Side       string  `json:"side"`
	OrderType  string  `json:"orderType"`
	Quantity   float64 `json:"quantity"`
	Price      float64 `json:"price"`
	Leverage   int     `json:"leverage"`
	MarginMode string  `json:"marginMode"`
}

// OrderInfo 订单信息
type OrderInfo struct {
	OrderId    string    `json:"orderId"`
	Symbol     string    `json:"symbol"`
	Side       string    `json:"side"`
	OrderType  string    `json:"orderType"`
	Quantity   float64   `json:"quantity"`
	Price      float64   `json:"price"`
	AvgPrice   float64   `json:"avgPrice"`
	Status     string    `json:"status"`
	FilledQty  float64   `json:"filledQty"`
	CreateTime time.Time `json:"createTime"`
	UpdateTime time.Time `json:"updateTime"`
}

// Position 持仓信息
type Position struct {
	Symbol           string  `json:"symbol"`
	Side             string  `json:"side"`
	Size             float64 `json:"size"`
	AvgPrice         float64 `json:"avgPrice"`
	Leverage         int     `json:"leverage"`
	Margin           float64 `json:"margin"` // 保证金 = 持仓价值 / 杠杆
	UnrealizedProfit float64 `json:"unrealizedProfit"`
	LiquidationPrice float64 `json:"liquidationPrice"`
}
