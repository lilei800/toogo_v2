// Package payment
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Author  Toogo Team
// @Description NOWPayments API 对接
// @Docs https://documenter.getpostman.com/view/7907941/S1a32n38
package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"strings"
	"time"
)

// NOWPayments NOWPayments支付网关
type NOWPayments struct {
	ApiKey      string // API Key
	IpnSecret   string // IPN Secret (用于验证回调)
	ApiEndpoint string // API端点
	IsSandbox   bool   // 是否沙盒环境
}

// NOWPaymentsConfig 配置
type NOWPaymentsConfig struct {
	ApiKey    string `json:"apiKey"`
	IpnSecret string `json:"ipnSecret"`
	IsSandbox bool   `json:"isSandbox"`
}

// NewNOWPayments 创建NOWPayments实例
func NewNOWPayments(config *NOWPaymentsConfig) *NOWPayments {
	endpoint := "https://api.nowpayments.io/v1"
	if config.IsSandbox {
		endpoint = "https://api-sandbox.nowpayments.io/v1"
	}
	return &NOWPayments{
		ApiKey:      config.ApiKey,
		IpnSecret:   config.IpnSecret,
		ApiEndpoint: endpoint,
		IsSandbox:   config.IsSandbox,
	}
}

// ========== 请求/响应结构 ==========

// CreatePaymentReq 创建支付请求
type CreatePaymentReq struct {
	PriceAmount      float64 `json:"price_amount"`                 // 支付金额
	PriceCurrency    string  `json:"price_currency"`               // 价格币种 (如: usd, usdt)
	PayCurrency      string  `json:"pay_currency"`                 // 支付币种 (如: usdttrc20)
	IpnCallbackUrl   string  `json:"ipn_callback_url,omitempty"`   // IPN回调地址
	OrderId          string  `json:"order_id,omitempty"`           // 订单ID
	OrderDescription string  `json:"order_description,omitempty"`  // 订单描述
	SuccessUrl       string  `json:"success_url,omitempty"`        // 成功跳转地址
	CancelUrl        string  `json:"cancel_url,omitempty"`         // 取消跳转地址
	IsFeePaidByUser  bool    `json:"is_fee_paid_by_user,omitempty"` // 手续费是否由用户支付
}

// CreatePaymentRes 创建支付响应
type CreatePaymentRes struct {
	PaymentId        string  `json:"payment_id"`         // NOWPayments支付ID
	PaymentStatus    string  `json:"payment_status"`     // 支付状态
	PayAddress       string  `json:"pay_address"`        // 支付地址
	PriceAmount      float64 `json:"price_amount"`       // 价格金额
	PriceCurrency    string  `json:"price_currency"`     // 价格币种
	PayAmount        float64 `json:"pay_amount"`         // 需要支付金额
	PayCurrency      string  `json:"pay_currency"`       // 支付币种
	OrderId          string  `json:"order_id"`           // 订单ID
	OrderDescription string  `json:"order_description"`  // 订单描述
	IpnCallbackUrl   string  `json:"ipn_callback_url"`   // 回调地址
	CreatedAt        string  `json:"created_at"`         // 创建时间
	UpdatedAt        string  `json:"updated_at"`         // 更新时间
	PurchaseId       string  `json:"purchase_id"`        // 购买ID
	Network          string  `json:"network"`            // 网络
	ExpirationTime   string  `json:"expiration_estimate_date"` // 过期时间
}

// CreateInvoiceReq 创建发票请求
type CreateInvoiceReq struct {
	PriceAmount    float64 `json:"price_amount"`              // 金额
	PriceCurrency  string  `json:"price_currency"`            // 币种
	PayCurrency    string  `json:"pay_currency,omitempty"`    // 支付币种(可选)
	IpnCallbackUrl string  `json:"ipn_callback_url"`          // IPN回调地址
	OrderId        string  `json:"order_id,omitempty"`        // 订单ID
	OrderDescription string `json:"order_description,omitempty"` // 描述
	SuccessUrl     string  `json:"success_url,omitempty"`     // 成功跳转
	CancelUrl      string  `json:"cancel_url,omitempty"`      // 取消跳转
	IsFixedRate    bool    `json:"is_fixed_rate,omitempty"`   // 是否固定汇率
	IsFeePaidByUser bool   `json:"is_fee_paid_by_user,omitempty"` // 手续费由用户支付
}

// CreateInvoiceRes 创建发票响应
type CreateInvoiceRes struct {
	Id             string  `json:"id"`              // 发票ID
	TokenId        string  `json:"token_id"`        // Token ID
	OrderId        string  `json:"order_id"`        // 订单ID
	OrderDescription string `json:"order_description"` // 描述
	PriceAmount    float64 `json:"price_amount"`    // 金额
	PriceCurrency  string  `json:"price_currency"`  // 币种
	PayCurrency    string  `json:"pay_currency"`    // 支付币种
	IpnCallbackUrl string  `json:"ipn_callback_url"` // 回调地址
	InvoiceUrl     string  `json:"invoice_url"`     // 发票页面URL
	SuccessUrl     string  `json:"success_url"`     // 成功跳转
	CancelUrl      string  `json:"cancel_url"`      // 取消跳转
	CreatedAt      string  `json:"created_at"`      // 创建时间
	UpdatedAt      string  `json:"updated_at"`      // 更新时间
	IsFixedRate    bool    `json:"is_fixed_rate"`   // 固定汇率
	IsFeePaidByUser bool   `json:"is_fee_paid_by_user"` // 手续费由用户支付
}

// PaymentStatusRes 支付状态响应
type PaymentStatusRes struct {
	PaymentId        string  `json:"payment_id"`
	PaymentStatus    string  `json:"payment_status"`
	PayAddress       string  `json:"pay_address"`
	PriceAmount      float64 `json:"price_amount"`
	PriceCurrency    string  `json:"price_currency"`
	PayAmount        float64 `json:"pay_amount"`
	ActuallyPaid     float64 `json:"actually_paid"`
	ActuallyPaidFiat float64 `json:"actually_paid_at_fiat"`
	PayCurrency      string  `json:"pay_currency"`
	OrderId          string  `json:"order_id"`
	OrderDescription string  `json:"order_description"`
	PurchaseId       string  `json:"purchase_id"`
	OutcomeAmount    float64 `json:"outcome_amount"`
	OutcomeCurrency  string  `json:"outcome_currency"`
}

// IPNCallback IPN回调数据
type IPNCallback struct {
	PaymentId         int64   `json:"payment_id"`
	PaymentStatus     string  `json:"payment_status"`
	PayAddress        string  `json:"pay_address"`
	PriceAmount       float64 `json:"price_amount"`
	PriceCurrency     string  `json:"price_currency"`
	PayAmount         float64 `json:"pay_amount"`
	ActuallyPaid      float64 `json:"actually_paid"`
	PayCurrency       string  `json:"pay_currency"`
	OrderId           string  `json:"order_id"`
	OrderDescription  string  `json:"order_description"`
	PurchaseId        string  `json:"purchase_id"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
	OutcomeAmount     float64 `json:"outcome_amount"`
	OutcomeCurrency   string  `json:"outcome_currency"`
}

// CreatePayoutReq 创建提现请求
type CreatePayoutReq struct {
	Address       string  `json:"address"`                   // 提现地址
	Currency      string  `json:"currency"`                  // 币种 (如: usdttrc20)
	Amount        float64 `json:"amount"`                    // 提现金额
	IpnCallbackUrl string `json:"ipn_callback_url,omitempty"` // IPN回调
	UniqueExternalPaymentId string `json:"unique_external_payment_id,omitempty"` // 外部订单ID
}

// CreatePayoutRes 创建提现响应
type CreatePayoutRes struct {
	Id            string  `json:"id"`
	Status        string  `json:"status"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	Address       string  `json:"address"`
	Hash          string  `json:"hash"`
	ExtraId       string  `json:"extra_id"`
	IpnCallbackUrl string `json:"ipn_callback_url"`
	CreatedAt     string  `json:"created_at"`
	RequestedAt   string  `json:"requested_at"`
	UniqueExternalPaymentId string `json:"unique_external_payment_id"`
	BatchWithdrawalId string `json:"batch_withdrawal_id"`
}

// ========== API方法 ==========

// GetStatus 获取API状态
func (n *NOWPayments) GetStatus(ctx context.Context) (bool, error) {
	res, err := n.request(ctx, "GET", "/status", nil)
	if err != nil {
		return false, err
	}
	message := gjson.New(res).Get("message").String()
	return message == "OK", nil
}

// GetCurrencies 获取支持的币种列表
func (n *NOWPayments) GetCurrencies(ctx context.Context) ([]string, error) {
	res, err := n.request(ctx, "GET", "/currencies", nil)
	if err != nil {
		return nil, err
	}
	currencies := gjson.New(res).Get("currencies").Strings()
	return currencies, nil
}

// GetMinimumPaymentAmount 获取最小支付金额
func (n *NOWPayments) GetMinimumPaymentAmount(ctx context.Context, currencyFrom, currencyTo string) (float64, error) {
	url := fmt.Sprintf("/min-amount?currency_from=%s&currency_to=%s", currencyFrom, currencyTo)
	res, err := n.request(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}
	return gjson.New(res).Get("min_amount").Float64(), nil
}

// GetEstimatedPrice 获取预估价格
func (n *NOWPayments) GetEstimatedPrice(ctx context.Context, amount float64, currencyFrom, currencyTo string) (float64, error) {
	url := fmt.Sprintf("/estimate?amount=%f&currency_from=%s&currency_to=%s", amount, currencyFrom, currencyTo)
	res, err := n.request(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}
	return gjson.New(res).Get("estimated_amount").Float64(), nil
}

// CreatePayment 创建支付 (直接获取支付地址)
func (n *NOWPayments) CreatePayment(ctx context.Context, req *CreatePaymentReq) (*CreatePaymentRes, error) {
	res, err := n.request(ctx, "POST", "/payment", req)
	if err != nil {
		return nil, err
	}

	var result CreatePaymentRes
	if err := json.Unmarshal(res, &result); err != nil {
		return nil, gerror.Wrap(err, "解析响应失败")
	}
	return &result, nil
}

// CreateInvoice 创建发票 (生成支付页面链接)
func (n *NOWPayments) CreateInvoice(ctx context.Context, req *CreateInvoiceReq) (*CreateInvoiceRes, error) {
	res, err := n.request(ctx, "POST", "/invoice", req)
	if err != nil {
		return nil, err
	}

	var result CreateInvoiceRes
	if err := json.Unmarshal(res, &result); err != nil {
		return nil, gerror.Wrap(err, "解析响应失败")
	}
	return &result, nil
}

// GetPaymentStatus 获取支付状态
func (n *NOWPayments) GetPaymentStatus(ctx context.Context, paymentId string) (*PaymentStatusRes, error) {
	res, err := n.request(ctx, "GET", "/payment/"+paymentId, nil)
	if err != nil {
		return nil, err
	}

	var result PaymentStatusRes
	if err := json.Unmarshal(res, &result); err != nil {
		return nil, gerror.Wrap(err, "解析响应失败")
	}
	return &result, nil
}

// CreatePayout 创建提现
func (n *NOWPayments) CreatePayout(ctx context.Context, req *CreatePayoutReq) (*CreatePayoutRes, error) {
	res, err := n.request(ctx, "POST", "/payout", req)
	if err != nil {
		return nil, err
	}

	var result CreatePayoutRes
	if err := json.Unmarshal(res, &result); err != nil {
		return nil, gerror.Wrap(err, "解析响应失败")
	}
	return &result, nil
}

// VerifyIPN 验证IPN回调签名
func (n *NOWPayments) VerifyIPN(ipnSecretFromHeader string, body []byte) bool {
	if n.IpnSecret == "" {
		return true // 未设置IPN密钥则跳过验证
	}

	// 对请求体进行排序后计算HMAC-SHA512
	mac := hmac.New(sha512.New, []byte(n.IpnSecret))
	
	// NOWPayments要求对JSON进行排序
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return false
	}
	
	sortedJson, err := json.Marshal(data)
	if err != nil {
		return false
	}
	
	mac.Write(sortedJson)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	
	return strings.EqualFold(expectedSignature, ipnSecretFromHeader)
}

// ParseIPNCallback 解析IPN回调数据
func (n *NOWPayments) ParseIPNCallback(body []byte) (*IPNCallback, error) {
	var callback IPNCallback
	if err := json.Unmarshal(body, &callback); err != nil {
		return nil, gerror.Wrap(err, "解析IPN回调数据失败")
	}
	return &callback, nil
}

// ========== 辅助方法 ==========

// request 发送请求
func (n *NOWPayments) request(ctx context.Context, method, path string, data interface{}) ([]byte, error) {
	client := gclient.New()
	client.SetTimeout(30 * time.Second)
	client.SetHeader("x-api-key", n.ApiKey)
	client.SetHeader("Content-Type", "application/json")

	url := n.ApiEndpoint + path

	var resp *gclient.Response
	var err error

	switch method {
	case "GET":
		resp, err = client.Get(ctx, url)
	case "POST":
		resp, err = client.Post(ctx, url, data)
	default:
		return nil, gerror.Newf("不支持的请求方法: %s", method)
	}

	if err != nil {
		return nil, gerror.Wrap(err, "请求失败")
	}
	defer resp.Close()

	body := resp.ReadAll()

	// 记录日志
	g.Log().Debugf(ctx, "NOWPayments %s %s, Response: %s", method, path, string(body))

	// 检查错误
	if resp.StatusCode >= 400 {
		errMsg := gjson.New(body).Get("message").String()
		if errMsg == "" {
			errMsg = string(body)
		}
		return nil, gerror.Newf("NOWPayments API错误(%d): %s", resp.StatusCode, errMsg)
	}

	return body, nil
}

// GetPaymentStatusText 获取支付状态文本
func GetPaymentStatusText(status string) string {
	statusMap := map[string]string{
		"waiting":       "等待支付",
		"confirming":    "确认中",
		"confirmed":     "已确认",
		"sending":       "发送中",
		"partially_paid": "部分支付",
		"finished":      "已完成",
		"failed":        "失败",
		"refunded":      "已退款",
		"expired":       "已过期",
	}
	if text, ok := statusMap[status]; ok {
		return text
	}
	return status
}

// IsPaymentSuccess 判断支付是否成功
func IsPaymentSuccess(status string) bool {
	return status == "finished" || status == "confirmed"
}

// GetCurrencyCode 根据网络获取币种代码
// network: TRC20, ERC20, BEP20
func GetCurrencyCode(currency, network string) string {
	network = strings.ToLower(network)
	currency = strings.ToLower(currency)
	
	if currency == "usdt" {
		switch network {
		case "trc20":
			return "usdttrc20"
		case "erc20":
			return "usdterc20"
		case "bep20":
			return "usdtbsc"
		default:
			return "usdttrc20"
		}
	}
	return currency
}

// ========== 全局实例 ==========

var nowPaymentsInstance *NOWPayments

// InitNOWPayments 初始化NOWPayments
func InitNOWPayments(ctx context.Context) error {
	// 从配置读取
	apiKey := g.Cfg().MustGet(ctx, "nowpayments.apiKey").String()
	ipnSecret := g.Cfg().MustGet(ctx, "nowpayments.ipnSecret").String()
	isSandbox := g.Cfg().MustGet(ctx, "nowpayments.isSandbox").Bool()

	if apiKey == "" {
		return gerror.New("NOWPayments API Key未配置")
	}

	nowPaymentsInstance = NewNOWPayments(&NOWPaymentsConfig{
		ApiKey:    apiKey,
		IpnSecret: ipnSecret,
		IsSandbox: isSandbox,
	})

	// 测试连接
	ok, err := nowPaymentsInstance.GetStatus(ctx)
	if err != nil {
		return gerror.Wrap(err, "NOWPayments连接测试失败")
	}
	if !ok {
		return gerror.New("NOWPayments服务不可用")
	}

	g.Log().Infof(ctx, "NOWPayments初始化成功, Sandbox: %v", isSandbox)
	return nil
}

// GetNOWPayments 获取NOWPayments实例
func GetNOWPayments() *NOWPayments {
	return nowPaymentsInstance
}

