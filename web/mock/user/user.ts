import { UserInfo } from '@/api/system/user';
import { defineMock } from '@alova/mock';
import { faker } from '@faker-js/faker';
import dayjs from 'dayjs';
import { resultSuccess } from '../util';

const userInfo: UserInfo = {
  id: faker.string.uuid(),
  avatar: faker.image.avatar(),
  birthday: faker.date.birthdate(),
  email: faker.internet.email(),
  username: faker.internet.userName(),
  account: faker.person.firstName('female'),
  sex: faker.person.sexType(),
  role: 'admin',
  permissions: [
    {
      label: '主控台',
      value: 'dashboard_console',
    },
    {
      label: '监控页',
      value: 'dashboard_monitor',
    },
    {
      label: '工作台',
      value: 'dashboard_workplace',
    },
    {
      label: '基础列表',
      value: 'basic_list',
    },
    {
      label: '基础列表删除',
      value: 'basic_list_delete',
    },
  ],
  status: faker.helpers.arrayElement([1, 2]),
  createDate: faker.date.anytime(),
};

const userList = Array.from({ length: 10 }).map(() => {
  const key = faker.number.int({ min: 2, max: 6 });
  return {
    id: faker.string.uuid(),
    avatar: `https://assets.naiveadmin.com/assets/avatar/avatar-${key}.jpg?v=${faker.string.numeric(
      4,
    )}`,
    birthday: faker.date.birthdate(),
    email: faker.internet.email(),
    username: faker.internet.userName(),
    account: faker.person.firstName(),
    sex: faker.person.sexType(),
    mobile: faker.helpers.fromRegExp(`1898888${faker.string.numeric(4)}`),
    role: faker.helpers.arrayElement(['admin', 'root']),
    permissions: [],
    status: faker.helpers.arrayElement(['close', 'refuse', 'pass']),
    createDate: dayjs(faker.date.anytime()).format('YYYY-MM-DD HH:mm'),
  };
});

export default defineMock({
  // 登录
  '[POST]/api/login': () =>
    resultSuccess({
      id: faker.number.int({ max: 99999999, min: 10000 }),
      token: faker.string.uuid(),
    }),
  // 用户信息
  '/api/admin_info': () => resultSuccess(userInfo),
  // 用户列表
  '/api/user_list': ({ query }) => {
    const { page = 1, pageSize = 10, name } = query;
    const list = userList;
    // 并非真实，只是为了模拟搜索结果
    const count = name ? 30 : 60;
    return resultSuccess({
      page: Number(page),
      pageSize: Number(pageSize),
      pageCount: count,
      total: count * Number(pageSize),
      list,
    });
  },
  // 退出
  '[POST]/api/logout': () => resultSuccess({}),
});
