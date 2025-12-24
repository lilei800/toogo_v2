// =================================================================================
// 机器人运行区间记录内部DAO
// =================================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TradingRobotRunSessionDao 机器人运行区间记录DAO
type TradingRobotRunSessionDao struct {
	table   string
	group   string
	columns TradingRobotRunSessionColumns
}

// TradingRobotRunSessionColumns 定义字段名
type TradingRobotRunSessionColumns struct {
	Id             string
	RobotId        string
	UserId         string
	Exchange       string
	Symbol         string
	StartTime      string
	EndTime        string
	EndReason      string
	RuntimeSeconds string
	TotalPnl       string
	TotalFee       string
	TradeCount     string
	SyncedAt       string
	CreatedAt      string
	UpdatedAt      string
}

var tradingRobotRunSessionColumns = TradingRobotRunSessionColumns{
	Id:             "id",
	RobotId:        "robot_id",
	UserId:         "user_id",
	Exchange:       "exchange",
	Symbol:         "symbol",
	StartTime:      "start_time",
	EndTime:        "end_time",
	EndReason:      "end_reason",
	RuntimeSeconds: "runtime_seconds",
	TotalPnl:       "total_pnl",
	TotalFee:       "total_fee",
	TradeCount:     "trade_count",
	SyncedAt:       "synced_at",
	CreatedAt:      "created_at",
	UpdatedAt:      "updated_at",
}

// NewTradingRobotRunSessionDao 创建DAO实例
func NewTradingRobotRunSessionDao() *TradingRobotRunSessionDao {
	return &TradingRobotRunSessionDao{
		table:   "hg_trading_robot_run_session",
		group:   "default",
		columns: tradingRobotRunSessionColumns,
	}
}

// DB 获取数据库连接
func (dao *TradingRobotRunSessionDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table 获取表名
func (dao *TradingRobotRunSessionDao) Table() string {
	return dao.table
}

// Columns 获取字段名
func (dao *TradingRobotRunSessionDao) Columns() TradingRobotRunSessionColumns {
	return dao.columns
}

// Group 获取分组名
func (dao *TradingRobotRunSessionDao) Group() string {
	return dao.group
}

// Ctx 获取带上下文的Model
func (dao *TradingRobotRunSessionDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction 事务
func (dao *TradingRobotRunSessionDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
