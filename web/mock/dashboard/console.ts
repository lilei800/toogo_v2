import { defineMock } from '@alova/mock';
import { faker } from '@faker-js/faker';
import { resultSuccess } from '../util';

interface Info {
  enrollment: number;
  ordersTotal: number;
  paymentAmount: number;
  totalSales: number;
}

const info: Info = {
  enrollment: Number(faker.commerce.price({ min: 10000, max: 99999, dec: 0 })), // 注册人数
  ordersTotal: Number(faker.commerce.price({ min: 10000, max: 99999, dec: 0 })), // 订单总数
  paymentAmount: Number(faker.commerce.price({ min: 1000, max: 99999, dec: 0 })), // 支付金额
  totalSales: Number(faker.commerce.price({ min: 10000, max: 99999, dec: 0 })), // 总销售额
};

export default defineMock({
  // 主控台数据
  '/api/dashboard/console': () => {
    return resultSuccess(info);
  },
});
