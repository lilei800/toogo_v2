import { http } from '@/utils/http/axios';

// 说明：
// - 本文件用于“客户端(前台)”调用后端 /api 前缀接口
// - 通过 per-request 覆盖 urlPrefix，避免影响后台(admin)已有接口

export interface SupportSession {
  id: number;
  userId: number;
  agentId: number;
  status: number; // 1排队 2进行中 3已关闭
  subject?: string;
  lastMsg?: string;
  lastMsgAt?: string;
  unreadUser?: number;
  unreadAgent?: number;
  createdAt?: string;
  updatedAt?: string;
  closedAt?: string;
}

export interface SupportMessage {
  id: number;
  sessionId: number;
  senderRole: number; // 1用户 2客服 3系统
  senderId: number;
  msgType: number;
  content: string;
  createdAt?: string;
}

const apiOpt = { urlPrefix: '/api' };

export function Start() {
  return http.request<{ session: SupportSession }>(
    {
      url: '/supportChat/start',
      method: 'POST',
    },
    apiOpt,
  );
}

export function Send(params: { sessionId: number; content: string }) {
  return http.request<{ message: SupportMessage }>(
    {
      url: '/supportChat/send',
      method: 'POST',
      params,
    },
    apiOpt,
  );
}

export function MessageList(params: { sessionId: number; page?: number; pageSize?: number }) {
  return http.request<{
    list: SupportMessage[];
    totalCount: number;
    page: number;
    pageSize: number;
  }>(
    {
      url: '/supportChat/messageList',
      method: 'GET',
      params,
    },
    apiOpt,
  );
}
