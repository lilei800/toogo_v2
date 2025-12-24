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

const coverList = [
  'https://assets.naiveadmin.com/assets/article/1.jpeg',
  'https://assets.naiveadmin.com/assets/article/2.jpeg',
  'https://assets.naiveadmin.com/assets/article/3.jpeg',
  'https://assets.naiveadmin.com/assets/article/4.jpg',
  'https://assets.naiveadmin.com/assets/article/5.jpeg',
  'https://assets.naiveadmin.com/assets/article/6.jpeg',
  'https://assets.naiveadmin.com/assets/article/7.jpeg',
  'https://assets.naiveadmin.com/assets/article/8.jpeg',
  'https://assets.naiveadmin.com/assets/article/9.jpeg',
  'https://assets.naiveadmin.com/assets/article/10.jpeg',
];

const articleList = (pageSize) => {
  const result: any[] = [];
  doCustomTimes(pageSize, () => {
    result.push({
      id: faker.string.numeric(10),
      title: faker.person.fullName(),
      tags: faker.helpers.arrayElements(
        [
          '有限理性',
          '智商',
          '情绪智力',
          '心理理论',
          '多动症',
          '抑郁症',
          '梦的解析',
          '催眠',
          '投射测验',
          '习惯化范式',
        ],
        faker.number.int({ min: 2, max: 4 }),
      ),
      summary: faker.music.genre(),
      avatar: faker.helpers.arrayElement(avatarList),
      cover: faker.helpers.arrayElement(coverList),
      author: faker.person.firstName(),
      collection: faker.number.int({ min: 10, max: 999 }),
      like: faker.number.int({ min: 10, max: 999 }),
      comment: faker.number.int({ min: 10, max: 999 }),
      date: dayjs(faker.date.anytime()).format('YYYY-MM-DD'),
    });
  });
  return result;
};

export default defineMock({
  '/api/article/list': ({ query }) => {
    const { page = 1, pageSize = 1 } = query;
    const list = articleList(Number(pageSize));
    return resultSuccess({
      page: Number(page),
      pageSize: Number(pageSize),
      pageCount: 60,
      itemCount: 60 * Number(pageSize),
      list,
    });
  },
});
