export enum SocketEnum {
  EventPing = 'ping',
  EventKick = 'kick',
  EventNotice = 'notice',
  EventConnected = 'connected',
  EventAdminMonitorTrends = 'admin/monitor/trends',
  EventAdminMonitorRunInfo = 'admin/monitor/runInfo',
  EventAdminOrderNotify = 'admin/order/notify',
  // Toogo - 机器人实时数据（批量实时分析推送）
  EventToogoRobotRealtimeSubscribe = 'toogo/robot/realtime/subscribe',
  EventToogoRobotRealtimeUnsubscribe = 'toogo/robot/realtime/unsubscribe',
  EventToogoRobotRealtimePush = 'toogo/robot/realtime/push',
  // Toogo - 机器人持仓实时数据（positions snapshot 推送）
  EventToogoRobotPositionsSubscribe = 'toogo/robot/positions/subscribe',
  EventToogoRobotPositionsUnsubscribe = 'toogo/robot/positions/unsubscribe',
  EventToogoRobotPositionsPush = 'toogo/robot/positions/push',
  // Toogo - 交易关键事件（平仓成功/订单状态变更等，服务端主动推送，无需订阅）
  EventToogoRobotTradeEvent = 'toogo/robot/trade/event',
  HeartBeatInterval = 1000,
  CodeSuc = 0,
  CodeErr = -1,
}
