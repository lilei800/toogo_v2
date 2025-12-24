// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TradingProxyConfigDao is the data access object for the table hg_trading_proxy_config.
type TradingProxyConfigDao struct {
	table    string                       // table is the underlying table name of the DAO.
	group    string                       // group is the database configuration group name of the current DAO.
	columns  TradingProxyConfigColumns    // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler           // handlers for customized model modification.
}

// TradingProxyConfigColumns defines and stores column names for the table hg_trading_proxy_config.
type TradingProxyConfigColumns struct {
	Id           string // 主键ID
	TenantId     string // 租户ID
	UserId       string // 用户ID
	Enabled      string // 是否启用：0=禁用,1=启用
	ProxyType    string // 代理类型：socks5/http
	ProxyAddress string // 代理地址
	AuthEnabled  string // 是否需要认证
	Username     string // 用户名
	Password     string // 密码（加密）
	LastTestTime string // 最后测试时间
	TestStatus   string // 测试状态：0=未测试,1=成功,2=失败
	TestMessage  string // 测试消息
	CreatedAt    string // 创建时间
	UpdatedAt    string // 更新时间
}

// tradingProxyConfigColumns holds the columns for the table hg_trading_proxy_config.
var tradingProxyConfigColumns = TradingProxyConfigColumns{
	Id:           "id",
	TenantId:     "tenant_id",
	UserId:       "user_id",
	Enabled:      "enabled",
	ProxyType:    "proxy_type",
	ProxyAddress: "proxy_address",
	AuthEnabled:  "auth_enabled",
	Username:     "username",
	Password:     "password",
	LastTestTime: "last_test_time",
	TestStatus:   "test_status",
	TestMessage:  "test_message",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
}

// NewTradingProxyConfigDao creates and returns a new DAO object for table data access.
func NewTradingProxyConfigDao(handlers ...gdb.ModelHandler) *TradingProxyConfigDao {
	return &TradingProxyConfigDao{
		group:    "default",
		table:    "hg_trading_proxy_config",
		columns:  tradingProxyConfigColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *TradingProxyConfigDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *TradingProxyConfigDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *TradingProxyConfigDao) Columns() TradingProxyConfigColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *TradingProxyConfigDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, it automatically sets the context for current operation.
func (dao *TradingProxyConfigDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *TradingProxyConfigDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

