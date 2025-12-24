// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ToogoVolatilityConfigDao is the data access object for table hg_toogo_volatility_config.
type ToogoVolatilityConfigDao struct {
	table   string
	group   string
	columns ToogoVolatilityConfigColumns
}

// ToogoVolatilityConfigColumns defines and stores column names for table hg_toogo_volatility_config.
// 适配新算法：市场状态阈值 + delta值 + DThreshold + 5个时间周期权重（1m/5m/15m/30m/1h）
type ToogoVolatilityConfigColumns struct {
	Id                      string
	Symbol                   string
	HighVolatilityThreshold  string
	LowVolatilityThreshold   string
	TrendStrengthThreshold   string
	DThreshold              string
	RangeVolatilityThreshold string
	Delta1m                 string
	Delta5m                string
	Delta15m               string
	Delta30m               string
	Delta1h                string
	Weight1m                 string
	Weight5m                string
	Weight15m               string
	Weight30m               string
	Weight1h                string
	IsActive                string
	CreatedAt               string
	UpdatedAt               string
}

// toogoVolatilityConfigColumns holds the columns for table hg_toogo_volatility_config.
var toogoVolatilityConfigColumns = ToogoVolatilityConfigColumns{
	Id:                      "id",
	Symbol:                   "symbol",
	HighVolatilityThreshold:  "high_volatility_threshold",
	LowVolatilityThreshold:   "low_volatility_threshold",
	TrendStrengthThreshold:   "trend_strength_threshold",
	DThreshold:              "d_threshold",
	RangeVolatilityThreshold: "range_volatility_threshold",
	Delta1m:                 "delta_1m",
	Delta5m:                "delta_5m",
	Delta15m:               "delta_15m",
	Delta30m:               "delta_30m",
	Delta1h:                "delta_1h",
	Weight1m:                 "weight_1m",
	Weight5m:                "weight_5m",
	Weight15m:               "weight_15m",
	Weight30m:               "weight_30m",
	Weight1h:                "weight_1h",
	IsActive:                "is_active",
	CreatedAt:               "created_at",
	UpdatedAt:               "updated_at",
}

// NewToogoVolatilityConfigDao creates and returns a new DAO object for table data access.
func NewToogoVolatilityConfigDao() *ToogoVolatilityConfigDao {
	return &ToogoVolatilityConfigDao{
		group:   "default",
		table:   "hg_toogo_volatility_config",
		columns: toogoVolatilityConfigColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *ToogoVolatilityConfigDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *ToogoVolatilityConfigDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *ToogoVolatilityConfigDao) Columns() ToogoVolatilityConfigColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *ToogoVolatilityConfigDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *ToogoVolatilityConfigDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *ToogoVolatilityConfigDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

