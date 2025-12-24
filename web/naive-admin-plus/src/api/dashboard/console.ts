import { Alova } from '@/utils/http/alova/index';

export interface ConsoleItem {
  enrollment: number;
  ordersTotal: number;
  paymentAmount: number;
  totalSales: number;
}

//获取主控台信息
export function getConsoleInfo() {
  return Alova.Get<ConsoleItem>('/dashboard/console', {
    meta: {
      authRole: null,
    },
  });
}
