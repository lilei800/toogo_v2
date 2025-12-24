// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TradingApiConfigDao is the data access object for the table hg_trading_api_config.
type TradingApiConfigDao struct {
	table    string                     // table is the underlying table name of the DAO.
	group    string                     // group is the database configuration group name of the current DAO.
	columns  TradingApiConfigColumns    // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler         // handlers for customized model modification.
}

// TradingApiConfigColumns defines and stores column names for the table hg_trading_api_config.
type TradingApiConfigColumns struct {
	Id             string // 主键ID
	TenantId       string // 租户ID
	UserId         string // 用户ID
	ApiName        string // API接口名称
	Platform       string // 平台名称：bitget/binance/okx
	BaseUrl        string // API地址
	ApiKey         string // API Key（加密）
	SecretKey      string // Secret Key（加密）
	Passphrase     string // Passphrase（加密，可选）
	IsDefault      string // 是否默认：0=否,1=是
	Status         string // 状态：1=正常,2=禁用
	LastVerifyTime string // 最后验证时间
	VerifyStatus   string // 验证状态：0=未验证,1=成功,2=失败
	VerifyMessage  string // 验证消息
	Remark         string // 备注
	CreatedAt      string // 创建时间
	UpdatedAt      string // 更新时间
	DeletedAt      string // 删除时间
}

// tradingApiConfigColumns holds the columns for the table hg_trading_api_config.
var tradingApiConfigColumns = TradingApiConfigColumns{
	Id:             "id",
	TenantId:       "tenant_id",
	UserId:         "user_id",
	ApiName:        "api_name",
	Platform:       "platform",
	BaseUrl:        "base_url",
	ApiKey:         "api_key",
	SecretKey:      "secret_key",
	Passphrase:     "passphrase",
	IsDefault:      "is_default",
	Status:         "status",
	LastVerifyTime: "last_verify_time",
	VerifyStatus:   "verify_status",
	VerifyMessage:  "verify_message",
	Remark:         "remark",
	CreatedAt:      "created_at",
	UpdatedAt:      "updated_at",
	DeletedAt:      "deleted_at",
}

// NewTradingApiConfigDao creates and returns a new DAO object for table data access.
func NewTradingApiConfigDao(handlers ...gdb.ModelHandler) *TradingApiConfigDao {
	return &TradingApiConfigDao{
		group:    "default",
		table:    "hg_trading_api_config",
		columns:  tradingApiConfigColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *TradingApiConfigDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *TradingApiConfigDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *TradingApiConfigDao) Columns() TradingApiConfigColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *TradingApiConfigDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, it automatically sets the context for current operation.
func (dao *TradingApiConfigDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rollbacks the transaction and returns the error from function f if it returns non-nil error.
// It commits the transaction and returns nil if function f returns nil.
func (dao *TradingApiConfigDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

