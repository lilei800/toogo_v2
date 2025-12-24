import pathlib

def replace_go_func_by_prefix(path, prefix, new_func_text):
    p = pathlib.Path(path)
    s = p.read_text(encoding='utf-8')
    nl = '\r\n' if '\r\n' in s else '\n'

    start = s.find(prefix)
    if start == -1:
        raise SystemExit(f'NOT FOUND prefix in {path}: {prefix}')

    brace_open = s.find('{', start)
    if brace_open == -1:
        raise SystemExit('No opening brace found')

    i = brace_open
    depth = 0
    in_str = in_raw = in_line = in_block = False

    while i < len(s):
        ch = s[i]
        nxt = s[i+1] if i+1 < len(s) else ''

        if in_line:
            if ch == '\n':
                in_line = False
            i += 1
            continue
        if in_block:
            if ch == '*' and nxt == '/':
                in_block = False
                i += 2
                continue
            i += 1
            continue
        if in_str:
            if ch == '\\':
                i += 2
                continue
            if ch == '"':
                in_str = False
            i += 1
            continue
        if in_raw:
            if ch == '`':
                in_raw = False
            i += 1
            continue

        if ch == '/' and nxt == '/':
            in_line = True
            i += 2
            continue
        if ch == '/' and nxt == '*':
            in_block = True
            i += 2
            continue
        if ch == '"':
            in_str = True
            i += 1
            continue
        if ch == '`':
            in_raw = True
            i += 1
            continue

        if ch == '{':
            depth += 1
        elif ch == '}':
            depth -= 1
            if depth == 0:
                end = i + 1
                new = new_func_text.replace('\n', nl)
                p.write_text(s[:start] + new + s[end:], encoding='utf-8')
                return
        i += 1

    raise SystemExit('No matching closing brace found')

robot_engine = r'D:\\go\\src\\hotgo_v2\\server\\internal\\logic\\toogo\\robot_engine.go'
robot_go = r'D:\\go\\src\\hotgo_v2\\server\\internal\\logic\\toogo\\robot.go'

new_engine_func = """func (e *RobotEngine) updateOrderStatusAfterClose(ctx context.Context, pos *exchange.Position, closeOrder *exchange.Order, closeType string) {
\trobot := e.Robot
\tif robot == nil || pos == nil {
\t\treturn
\t}

\t// 规范化平仓原因（用于 Wallet/成交明细展示与统计口径）
\tcloseReason := strings.TrimSpace(strings.ToLower(closeType))
\tswitch closeReason {
\tcase \"stop_loss\", \"take_profit\", \"manual\", \"timeout\", \"unknown\":
\t\t// ok
\tdefault:
\t\t// 兼容历史/中文描述
\t\tif strings.Contains(closeType, \"止损\") {
\t\t\tcloseReason = \"stop_loss\"
\t\t} else if strings.Contains(closeType, \"止盈\") || strings.Contains(closeType, \"回撤\") {
\t\t\tcloseReason = \"take_profit\"
\t\t} else if strings.Contains(closeType, \"手动\") {
\t\t\tcloseReason = \"manual\"
\t\t} else {
\t\t\tcloseReason = \"unknown\"
\t\t}
\t}

\t// 确定方向
\tdirection := \"long\"
\tif pos.PositionSide == \"SHORT\" {
\t\tdirection = \"short\"
\t}

\t// 查找本地 OPEN 订单（平仓必须落到 hg_trading_order，否则前端 /toogo/wallet/order/history 查不到）
\tvar order *entity.TradingOrder
\terr := dao.TradingOrder.Ctx(ctx).
\t\tWhere(\"robot_id\", robot.Id).
\t\tWhere(\"direction\", direction).
\t\tWhere(\"status\", OrderStatusOpen).
\t\tOrderDesc(\"id\").
\t\tLimit(1).
\t\tScan(&order)
\tif err != nil || order == nil {
\t\tg.Log().Warningf(ctx, \"[RobotEngine] robotId=%d 平仓后未找到本地OPEN订单，无法写入成交明细: direction=%s, err=%v\", robot.Id, direction, err)
\t\treturn
\t}

\t// 平仓价格：优先交易所均价，其次标记价，最后兜底开仓价
\tclosePrice := 0.0
\tif closeOrder != nil && closeOrder.AvgPrice > 0 {
\t\tclosePrice = closeOrder.AvgPrice
\t} else if pos.MarkPrice > 0 {
\t\tclosePrice = pos.MarkPrice
\t} else if order.OpenPrice > 0 {
\t\tclosePrice = order.OpenPrice
\t}

\t// 已实现盈亏：优先用交易所回传(若有)，否则先用 pos.UnrealizedPnl 做近似兜底，后续同步可再补全
\trealizedProfit := 0.0
\tif closeOrder != nil && closeOrder.RealizedPnl != 0 {
\t\trealizedProfit = closeOrder.RealizedPnl
\t} else if pos.UnrealizedPnl != 0 {
\t\trealizedProfit = pos.UnrealizedPnl
\t}

\t// 统一走 CloseOrder：写 status/close_price/close_time/realized_profit/close_reason，并处理扣算力等
\tGetOrderStatusSyncService().CloseOrder(ctx, order, closePrice, realizedProfit, closeReason, closeOrder, pos)

\t// 推送平仓成功事件给前端：用于详情弹窗订单列表秒级刷新（不依赖轮询/同步）
\tif robot.UserId > 0 {
\t\tcloseOrderId := \"\"
\t\tif closeOrder != nil {
\t\t\tcloseOrderId = closeOrder.OrderId
\t\t}
\t\twebsocket.SendToUser(robot.UserId, &websocket.WResponse{
\t\t\tEvent: \"toogo/robot/trade/event\",
\t\t\tData: g.Map{
\t\t\t\t\"type\":          \"close_success\",
\t\t\t\t\"robotId\":        robot.Id,
\t\t\t\t\"symbol\":         robot.Symbol,
\t\t\t\t\"positionSide\":   pos.PositionSide,
\t\t\t\t\"direction\":      direction,
\t\t\t\t\"closeOrderId\":   closeOrderId,
\t\t\t\t\"closePrice\":     closePrice,
\t\t\t\t\"realizedProfit\": realizedProfit,
\t\t\t\t\"ts\":            gtime.Now().TimestampMilli(),
\t\t\t},
\t\t})
\t}
}
"""

new_robot_func = """func (s *sToogoRobot) updateOrderAfterManualClose(ctx context.Context, robotId int64, positionSide string, closePrice, realizedProfit float64, exchangeCloseId string) {
\t// 确定方向
\tdirection := \"long\"
\tif positionSide == \"SHORT\" {
\t\tdirection = \"short\"
\t}

\t// 查找本地 OPEN 订单
\tvar order *entity.TradingOrder
\terr := dao.TradingOrder.Ctx(ctx).
\t\tWhere(\"robot_id\", robotId).
\t\tWhere(\"direction\", direction).
\t\tWhere(\"status\", OrderStatusOpen).
\t\tOrderDesc(\"id\").
\t\tLimit(1).
\t\tScan(&order)
\tif err != nil || order == nil {
\t\tg.Log().Warningf(ctx, \"[ClosePosition] robotId=%d 未找到本地OPEN订单，无法写入成交明细: direction=%s, err=%v\", robotId, direction, err)
\t\treturn
\t}

\t// 手动平仓这里不强依赖交易所回传对象（CloseOrder 会负责补全 close_time/close_reason 等）
\tGetOrderStatusSyncService().CloseOrder(ctx, order, closePrice, realizedProfit, \"manual\", nil, nil)
\tg.Log().Infof(ctx, \"[ClosePosition] robotId=%d 订单状态已更新为已平仓(统一落库): direction=%s, closePrice=%.4f, realizedProfit=%.4f\", robotId, direction, closePrice, realizedProfit)
}
"""

replace_go_func_by_prefix(robot_engine, 'func (e *RobotEngine) updateOrderStatusAfterClose(', new_engine_func)
replace_go_func_by_prefix(robot_go, 'func (s *sToogoRobot) updateOrderAfterManualClose(', new_robot_func)
print('OK')
