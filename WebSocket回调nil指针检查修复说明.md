# WebSocket 回调 nil 指针检查修复说明

## 问题描述

在 WebSocket 服务中，当触发回调函数时，如果回调函数指针为 nil，会导致 panic，引起服务崩溃。

## 修复内容

### 1. Bitget WebSocket (`bitget_ws.go`)

在以下位置添加了 nil 检查：

#### 1.1 Ticker 回调 (约第330行)
```go
for _, cb := range callbacks {
    if cb != nil {  // 添加 nil 检查
        go cb(ticker)
    }
}
```

#### 1.2 K线回调 (约第394行)
```go
for _, cb := range callbacks {
    if cb != nil && klinesCopy != nil {  // 添加 nil 检查
        go cb(klinesCopy)
    }
}
```

### 2. Binance WebSocket (`binance_ws.go`)

在以下位置添加了 nil 检查：

#### 2.1 Ticker 回调 (约第295行)
```go
for _, cb := range callbacks {
    if cb != nil {  // 添加 nil 检查
        go cb(ticker)
    }
}
```

#### 2.2 K线回调 (约第351行)
```go
for _, cb := range callbacks {
    if cb != nil && klinesCopy != nil {  // 添加 nil 检查
        go cb(klinesCopy)
    }
}
```

## 修复原因

1. **防止 nil 函数指针调用**：回调列表中可能存在 nil 函数指针，直接调用会导致 panic
2. **防止 nil 数据传递**：K线数据可能为 nil，传递给回调函数可能导致问题
3. **提高系统稳定性**：避免因单个回调问题导致整个 WebSocket 服务崩溃

## 修复时间

2025-12-08

## 测试验证

修复后服务器启动正常，WebSocket 连接正常工作：
- Bitget WebSocket 连接成功
- Binance WebSocket 连接成功
- Ticker 和 K线数据正常接收
- 回调函数正常触发

## 相关文件

- `server/internal/library/exchange/bitget_ws.go`
- `server/internal/library/exchange/binance_ws.go`
