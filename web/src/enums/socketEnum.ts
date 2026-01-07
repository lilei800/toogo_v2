export enum SocketEnum {
  EventPing = 'ping',
  EventJoin = 'join',
  EventQuit = 'quit',
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
  // Toogo - 持仓增量变更（服务端主动推送，无需订阅；来源=交易所私有WS position/account 事件触发的即时刷新）
  EventToogoRobotPositionsDelta = 'toogo/robot/positions/delta',
  // Toogo - 机器人挂单实时数据（open orders snapshot 推送）
  EventToogoRobotOrdersSubscribe = 'toogo/robot/orders/subscribe',
  EventToogoRobotOrdersUnsubscribe = 'toogo/robot/orders/unsubscribe',
  EventToogoRobotOrdersPush = 'toogo/robot/orders/push',
  // Toogo - 机器人挂单增量变更（服务端主动推送，无需订阅；来源=交易所私有WS/对账）
  EventToogoRobotOrdersDelta = 'toogo/robot/orders/delta',
  // Toogo - 交易关键事件（平仓成功/订单状态变更等，服务端主动推送，无需订阅）
  EventToogoRobotTradeEvent = 'toogo/robot/trade/event',

  // Support - 客服聊天
  EventSupportSessionUpdated = 'support/session/updated',
  EventSupportMessage = 'support/message',
  HeartBeatInterval = 1000,
  CodeSuc = 0,
  CodeErr = -1,
}
