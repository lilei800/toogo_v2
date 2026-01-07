// Toogo WebSocket 客户端
// 重新导出 HotGo 原版 WebSocket 功能
export {
  default as default,
  sendMsg,
  addOnMessage,
  removeOnMessage,
  getAllOnMessage,
  type WebSocketMessage,
} from '@/utils/websocket/index';

import { useUserStore } from '@/store/modules/user';

// ============ Toogo 专用 WebSocket ============

// 消息类型
export enum MessageType {
  TICKER = 'ticker',
  POSITION = 'position',
  ORDER = 'order',
  TRADE = 'trade',
  ROBOT = 'robot',
  SIGNAL = 'signal',
  PNL = 'pnl',
  SYSTEM = 'system',
  ERROR = 'error',
}

// WebSocket消息
export interface WsMessage {
  type: MessageType;
  channel?: string;
  data: any;
  timestamp: number;
}

// 行情数据
export interface TickerData {
  symbol: string;
  price: number;
  high_24h: number;
  low_24h: number;
  volume_24h: number;
  change_24h: number;
}

// 持仓数据
export interface PositionData {
  robot_id: number;
  symbol: string;
  side: 'long' | 'short';
  size: number;
  entry_price: number;
  mark_price: number;
  unrealized_pnl: number;
  leverage: number;
  margin: number;
  liquid_price: number;
}

// 订单数据
export interface OrderData {
  robot_id: number;
  order_id: string;
  symbol: string;
  side: string;
  type: string;
  status: string;
  price: number;
  size: number;
  filled: number;
  pnl: number;
  created_at: string;
}

// 机器人状态
export interface RobotStatusData {
  robot_id: number;
  name: string;
  status: number;
  total_profit: number;
  consumed_power: number;
  open_orders: number;
  market_state: string;
  risk_level: string;
  signal: string;
}

// 交易信号
export interface SignalData {
  robot_id: number;
  symbol: string;
  signal: 'buy' | 'sell' | 'hold';
  confidence: number;
  market_state: string;
  reason: string;
}

// 消息处理回调
type ToogoMessageHandler = (message: WsMessage) => void;

class ToogoWebSocket {
  private ws: WebSocket | null = null;
  private url: string;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectInterval = 3000;
  private heartbeatInterval: number | null = null;
  private handlers: Map<MessageType, Set<ToogoMessageHandler>> = new Map();
  private isConnecting = false;
  private subscriptions: Set<string> = new Set();

  constructor() {
    // 从环境变量或配置获取WebSocket地址
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = import.meta.env.VITE_WS_HOST || window.location.host;
    this.url = `${wsProtocol}//${host}/ws`;
  }

  // 连接WebSocket
  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        resolve();
        return;
      }

      if (this.isConnecting) {
        reject(new Error('Already connecting'));
        return;
      }

      this.isConnecting = true;

      // 获取用户token
      const userStore = useUserStore();
      const token = userStore.getToken;
      const connectUrl = token ? `${this.url}?token=${token}` : this.url;

      this.ws = new WebSocket(connectUrl);

      this.ws.onopen = () => {
        console.log('[ToogoWebSocket] Connected');
        this.isConnecting = false;
        this.reconnectAttempts = 0;
        this.startHeartbeat();

        // 重新订阅之前的频道
        this.subscriptions.forEach((channel) => {
          this.subscribe(channel);
        });

        resolve();
      };

      this.ws.onmessage = (event) => {
        this.handleMessage(event.data);
      };

      this.ws.onclose = (event) => {
        console.log('[ToogoWebSocket] Closed:', event.code, event.reason);
        this.isConnecting = false;
        this.stopHeartbeat();
        this.tryReconnect();
      };

      this.ws.onerror = (error) => {
        console.error('[ToogoWebSocket] Error:', error);
        this.isConnecting = false;
        reject(error);
      };
    });
  }

  // 断开连接
  disconnect() {
    this.stopHeartbeat();
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.subscriptions.clear();
    this.reconnectAttempts = this.maxReconnectAttempts; // 阻止重连
  }

  // 发送消息
  send(data: any) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    } else {
      console.warn('[ToogoWebSocket] Not connected');
    }
  }

  // 订阅频道
  subscribe(channel: string) {
    this.subscriptions.add(channel);
    this.send({ action: 'subscribe', channel });
  }

  // 取消订阅
  unsubscribe(channel: string) {
    this.subscriptions.delete(channel);
    this.send({ action: 'unsubscribe', channel });
  }

  // 订阅行情
  subscribeTicker(symbol: string) {
    this.subscribe(`ticker:${symbol}`);
  }

  // 取消订阅行情
  unsubscribeTicker(symbol: string) {
    this.unsubscribe(`ticker:${symbol}`);
  }

  // 注册消息处理器
  on(type: MessageType, handler: ToogoMessageHandler) {
    if (!this.handlers.has(type)) {
      this.handlers.set(type, new Set());
    }
    this.handlers.get(type)!.add(handler);
  }

  // 移除消息处理器
  off(type: MessageType, handler: ToogoMessageHandler) {
    this.handlers.get(type)?.delete(handler);
  }

  // 移除所有处理器
  offAll(type?: MessageType) {
    if (type) {
      this.handlers.delete(type);
    } else {
      this.handlers.clear();
    }
  }

  // 处理收到的消息
  private handleMessage(data: string) {
    try {
      // 可能是多条消息以换行分隔
      const messages = data.split('\n').filter(Boolean);

      for (const msgStr of messages) {
        const message: WsMessage = JSON.parse(msgStr);

        // 触发对应类型的处理器
        const handlers = this.handlers.get(message.type as MessageType);
        if (handlers) {
          handlers.forEach((handler) => handler(message));
        }
      }
    } catch (error) {
      console.error('[ToogoWebSocket] Parse message error:', error);
    }
  }

  // 开始心跳
  private startHeartbeat() {
    this.heartbeatInterval = window.setInterval(() => {
      this.send({ action: 'ping' });
    }, 30000);
  }

  // 停止心跳
  private stopHeartbeat() {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
  }

  // 尝试重连
  private tryReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.log('[ToogoWebSocket] Max reconnect attempts reached');
      return;
    }

    this.reconnectAttempts++;
    console.log(
      `[ToogoWebSocket] Reconnecting... (${this.reconnectAttempts}/${this.maxReconnectAttempts})`,
    );

    setTimeout(() => {
      this.connect().catch(console.error);
    }, this.reconnectInterval);
  }

  // 获取连接状态
  get isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }
}

// 导出单例
export const toogoWs = new ToogoWebSocket();

// Vue组合式API封装
export function useToogoWebSocket() {
  const connect = () => toogoWs.connect();
  const disconnect = () => toogoWs.disconnect();
  const subscribe = (channel: string) => toogoWs.subscribe(channel);
  const unsubscribe = (channel: string) => toogoWs.unsubscribe(channel);
  const subscribeTicker = (symbol: string) => toogoWs.subscribeTicker(symbol);
  const unsubscribeTicker = (symbol: string) => toogoWs.unsubscribeTicker(symbol);
  const on = (type: MessageType, handler: ToogoMessageHandler) => toogoWs.on(type, handler);
  const off = (type: MessageType, handler: ToogoMessageHandler) => toogoWs.off(type, handler);
  const offAll = (type?: MessageType) => toogoWs.offAll(type);

  return {
    connect,
    disconnect,
    subscribe,
    unsubscribe,
    subscribeTicker,
    unsubscribeTicker,
    on,
    off,
    offAll,
    isConnected: () => toogoWs.isConnected,
  };
}
