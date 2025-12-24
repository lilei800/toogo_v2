// =================================================================================
// 机器人运行区间记录DAO
// =================================================================================

package dao

import (
	"hotgo/internal/dao/internal"
)

// TradingRobotRunSession 机器人运行区间记录DAO
var TradingRobotRunSession = &tradingRobotRunSessionDao{
	internal.NewTradingRobotRunSessionDao(),
}

type tradingRobotRunSessionDao struct {
	*internal.TradingRobotRunSessionDao
}
