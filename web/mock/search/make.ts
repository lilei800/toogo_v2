import { defineMock } from '@alova/mock';
import { faker } from '@faker-js/faker';
import dayjs from 'dayjs';
import { resultSuccess, doCustomTimes } from '../util';

const avatarList = [
  'https://assets.naiveadmin.com/assets/avatar/avatar-1.jpg',
  'https://assets.naiveadmin.com/assets/avatar/avatar-2.jpg',
  'https://assets.naiveadmin.com/assets/avatar/avatar-3.jpg',
  'https://assets.naiveadmin.com/assets/avatar/avatar-4.jpg',
  'https://assets.naiveadmin.com/assets/avatar/avatar-5.jpg',
  'https://assets.naiveadmin.com/assets/avatar/avatar-6.jpg',
];

const makeList = (pageSize) => {
  const result: any[] = [];
  doCustomTimes(pageSize, () => {
    result.push({
      id: faker.string.numeric(4),
      doctor: faker.person.firstName(),
      avatar: faker.helpers.arrayElement(avatarList),
      subject: faker.helpers.arrayElement([
        '中医内科',
        '中医外科',
        '中医儿科',
        '中医妇科',
        '中医针灸科',
        '中医五官科',
        '中医骨伤科',
      ]),
      date: dayjs(faker.date.anytime()).format('YYYY-MM-DD'),
    });
  });
  return result;
};

export default defineMock({
  '/api/make/list': ({ query }) => {
    const { page = 1, pageSize = 1 } = query;
    const list = makeList(Number(pageSize));
    return resultSuccess({
      page: Number(page),
      pageSize: Number(pageSize),
      pageCount: 60,
      itemCount: 60 * Number(pageSize),
      list,
    });
  },
});
