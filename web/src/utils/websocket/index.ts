import { SocketEnum } from '@/enums/socketEnum';
import { useUserStoreWidthOut } from '@/store/modules/user';
import { isJsonString } from '@/utils/is';
import { registerGlobalMessage } from '@/utils/websocket/registerMessage';

// WebSocket娑堟伅鏍煎紡
export interface WebSocketMessage {
  event: string;
  data: any;
  code: number;
  timestamp: number;
}

let socket: WebSocket;
let isActive: boolean;
const messageHandler: Map<string, Function> = new Map();

export default () => {
  const heartCheck = {
    timeout: 5000,
    timeoutObj: setTimeout(() => {}),
    serverTimeoutObj: setInterval(() => {}),
    reset: function () {
      clearTimeout(this.timeoutObj);
      clearTimeout(this.serverTimeoutObj);
      return this;
    },
    start: function () {
      // eslint-disable-next-line @typescript-eslint/no-this-alias
      const self = this;
      clearTimeout(this.timeoutObj);
      clearTimeout(this.serverTimeoutObj);
      this.timeoutObj = setTimeout(function () {
        socket.send(
          JSON.stringify({
            event: SocketEnum.EventPing,
          })
        );
        self.serverTimeoutObj = setTimeout(function () {
          console.log('[WebSocket] Log');
          socket.close();
        }, self.timeout);
      }, this.timeout);
    },
  };

  const useUserStore = useUserStoreWidthOut();
  let lockReconnect = false;
  let timer: ReturnType<typeof setTimeout>;

  const getWsAddr = (): string => {
    const cfgAddr = useUserStore.config?.wsAddr;
    if (cfgAddr && cfgAddr !== '') {
      // 鍏煎锛欻otGo ws 璺敱閫氬父鏄?/socket/锛堟湯灏惧甫/锛夛紝閮ㄥ垎鏈嶅姟鍣ㄥ /socket 浼?301 鍒?/socket/
      // WebSocket 涓嶄竴瀹氳兘鑷姩璺熼殢 301锛屽洜姝よ繖閲岀粺涓€琛ラ綈鏈熬鏂滄潬
      return cfgAddr.endsWith('/socket') ? `${cfgAddr}/` : cfgAddr;
    }

    // 鍏滃簳锛氫娇鐢ㄥ綋鍓?hostname + HotGo 榛樿 ws 鍓嶇紑 /socket
    // - 寮€鍙戠幆澧冨父瑙侊細鍓嶇 8009锛屽悗绔?8000
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const hostname = window.location.hostname;
    let host = window.location.host;

    // Dev fallback: when frontend runs on 8009 (or any non-8000 port), connect WS directly to backend :8000.
    // This avoids Vite ws proxy handshake issues in some environments.
    const isDev = import.meta.env.DEV;
    if (isDev && window.location.port && window.location.port !== '8000') {
      host = `${hostname}:8000`;
    }

    return `${wsProtocol}//${host}/socket/`;
  };

  const createSocket = () => {
    console.log('[WebSocket] createSocket...');
    if (useUserStore.token === '') {
      console.error('[WebSocket] Error');
      resetReconnect();
      return;
    }
    try {
      const wsAddr = getWsAddr();
      if (!wsAddr) {
        console.error('[WebSocket] Error');
        resetReconnect();
        return;
      }
      console.log('[WebSocket] Closed');
      socket = new WebSocket(`${wsAddr}?authorization=${useUserStore.token}`);
      init();
      if (lockReconnect) {
        lockReconnect = false;
      }
    } catch (e) {
      console.error(`[WebSocket] createSocket err: ${e}`);
      resetReconnect();
      return;
    }
  };

  const resetReconnect = () => {
    if (lockReconnect) {
      lockReconnect = false;
    }
    reconnect();
  };

  const reconnect = () => {
      console.log('[WebSocket] Closed');
    if (lockReconnect) return;
    lockReconnect = true;
    clearTimeout(timer);
    timer = setTimeout(() => {
      createSocket();
    }, 1000 * 10);
  };

  const init = () => {
    socket.onopen = function (_) {
      console.log('[WebSocket] Connected');
      heartCheck.reset().start();
      isActive = true;
    };

    socket.onmessage = function (event) {
      isActive = true;
      // console.log('WebSocket:鏀跺埌涓€鏉℃秷鎭?, event.data);

      if (!isJsonString(event.data)) {
      console.log('[WebSocket] Closed');
        return;
      }

      heartCheck.reset().start();

      const message = JSON.parse(event.data) as WebSocketMessage;
      onMessage(message);
    };

    socket.onerror = function (_) {
      console.log('[WebSocket] Log');
      reconnect();
      isActive = false;
    };

    socket.onclose = function (_) {
      console.log('[WebSocket] Closed');
      heartCheck.reset();
      reconnect();
      isActive = false;
    };

    window.onbeforeunload = function () {
      socket.close();
      isActive = false;
    };
  };

  createSocket();
  registerGlobalMessage();
};

function onMessage(message: WebSocketMessage) {
  let handled = false;
  messageHandler.forEach((value: Function, key: string) => {
    if (message.event === key || key === '*') {
      handled = true;
      value.call(null, message);
    }
  });

  if (!handled) {
      console.log('[WebSocket] Closed');
  }
}

// 鍙戦€佹秷鎭?
export function sendMsg(event: string, data: any = null, isRetry = true) {
  if (socket === undefined || !isActive) {
    if (!isRetry) {
      console.log('[WebSocket] Log');
      return;
    }
    console.log('[WebSocket] Log');
    setTimeout(() => {
      sendMsg(event, data);
    }, 200);
    return;
  }

  try {
    socket.send(JSON.stringify({ event, data }));
  } catch (err: any) {
      console.log('[WebSocket] Closed');
    if (!isRetry) {
      return;
    }

    console.log('[WebSocket] Log');
    setTimeout(() => {
      sendMsg(event, data);
    }, 100);
  }
}

// 娣诲姞娑堟伅澶勭悊
export function addOnMessage(key: string, value: Function): void {
  messageHandler.set(key, value);
}

// 绉婚櫎娑堟伅澶勭悊
export function removeOnMessage(key: string): boolean {
  return messageHandler.delete(key);
}

// 鏌ョ湅鎵€鏈夋秷鎭鐞?
export function getAllOnMessage(): Map<string, Function> {
  return messageHandler;
}


