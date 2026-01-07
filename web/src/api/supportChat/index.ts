import { http } from '@/utils/http/axios';

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

export interface SupportCannedReply {
  id: number;
  agentId: number;
  title: string;
  content: string;
  sort: number;
  status: number;
  createdAt?: string;
  updatedAt?: string;
}

export function AgentOnline(params: { online: boolean }) {
  return http.request({
    url: '/supportChat/agentOnline',
    method: 'POST',
    params,
  });
}

export function SessionList(params: { status?: number; page?: number; pageSize?: number }) {
  return http.request<{
    list: SupportSession[];
    totalCount: number;
    page: number;
    pageSize: number;
  }>({
    url: '/supportChat/sessionList',
    method: 'GET',
    params,
  });
}

export function Accept(params: { sessionId: number }) {
  return http.request<{ session: SupportSession }>({
    url: '/supportChat/accept',
    method: 'POST',
    params,
  });
}

export function AcceptNext() {
  return http.request<{ session: SupportSession }>({
    url: '/supportChat/acceptNext',
    method: 'POST',
  });
}

export function Close(params: { sessionId: number }) {
  return http.request({
    url: '/supportChat/close',
    method: 'POST',
    params,
  });
}

export function Transfer(params: { sessionId: number; toAgentId: number }) {
  return http.request<{ session: SupportSession }>({
    url: '/supportChat/transfer',
    method: 'POST',
    params,
  });
}

export function Send(params: { sessionId: number; content: string }) {
  return http.request<{ message: SupportMessage }>({
    url: '/supportChat/send',
    method: 'POST',
    params,
  });
}

export function MessageList(params: { sessionId: number; page?: number; pageSize?: number }) {
  return http.request<{
    list: SupportMessage[];
    totalCount: number;
    page: number;
    pageSize: number;
  }>({
    url: '/supportChat/messageList',
    method: 'GET',
    params,
  });
}

export function CannedList(params: { page?: number; pageSize?: number }) {
  return http.request<{
    list: SupportCannedReply[];
    totalCount: number;
    page: number;
    pageSize: number;
  }>({
    url: '/supportChat/canned/list',
    method: 'GET',
    params,
  });
}

export function CannedEdit(params: Partial<SupportCannedReply>) {
  return http.request({
    url: '/supportChat/canned/edit',
    method: 'POST',
    params,
  });
}

export function CannedDelete(params: { id: number }) {
  return http.request({
    url: '/supportChat/canned/delete',
    method: 'POST',
    params,
  });
}
