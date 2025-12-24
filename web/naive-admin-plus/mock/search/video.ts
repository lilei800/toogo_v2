import { defineMock } from '@alova/mock';
import { faker } from '@faker-js/faker';
import dayjs from 'dayjs';
import { resultSuccess, doCustomTimes } from '../util';

const avatargroupList = [
  {
    name: '张三',
    src: 'https://assets.naiveadmin.com/assets/avatar/avatar-1.jpg',
  },
  {
    name: '李四',
    src: 'https://assets.naiveadmin.com/assets/avatar/avatar-2.jpg',
  },
  {
    name: '王五',
    src: 'https://assets.naiveadmin.com/assets/avatar/avatar-3.jpg',
  },
  {
    name: '赵六',
    src: 'https://assets.naiveadmin.com/assets/avatar/avatar-4.jpg',
  },
  {
    name: '七仔',
    src: 'https://assets.naiveadmin.com/assets/avatar/avatar-5.jpg',
  },
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

const videoList = (pageSize) => {
  const result: any[] = [];
  doCustomTimes(pageSize, () => {
    result.push({
      id: faker.string.numeric(4),
      title: faker.helpers.arrayElement([
        'TTT培训-企业内训师（TTT）',
        '卓越管理-打造高效执行力',
        '卓越领导力-目标管理与计划执行',
        '九型人格与管理应用',
        '深刻理解激励辅导下属的内涵及价值，并积极有效的改变辅导的观念与...',
        '裂变-创新时代卓越领导艺术与实践',
      ]),
      summary: faker.helpers.arrayElement([
        '帮助企业内部培训师充分认识自己的角色和任务，树立培训师的职业形...',
        '向复杂的大型机构客户销售产品和方案的销售方法论',
        '分析众多真实、鲜活的挑战性的销售案例，结合客户购买的6个心理周期...',
        '了解大客户销售中客户的决策方式，购买特点，行为心理，有针对性地...',
        '没有搞不定的订单，只有不会成交的销售。本课程将教给您：用头脑做...',
        '精准销售模式是以企业现有销售团队为基础，通过高层推动，重新梳理...',
      ]),
      avatargroup: getRandomArrayElements(avatargroupList, 2, 5),
      cover: faker.helpers.arrayElement(coverList),
      viewingtimes: faker.number.int({ min: 10, max: 999 }),
      date: dayjs(faker.date.anytime()).format('YYYY-MM-DD'),
    });
  });
  return result;
};

export default defineMock({
  '/api/video/list': ({ query }) => {
    const { page = 1, pageSize = 1 } = query;
    const list = videoList(Number(pageSize));
    return resultSuccess({
      page: Number(page),
      pageSize: Number(pageSize),
      pageCount: 60,
      itemCount: 60 * Number(pageSize),
      list,
    });
  },
});

//从数组中取出指定个数的元素
function getRandomArrayElements(arr, start, end) {
  const count = Math.floor(Math.random() * (end - start) + start);
  const shuffled = arr.slice(0);
  let i = arr.length;
  const min = i - count;
  let temp: any;
  let index: number;
  while (i-- > min) {
    index = Math.floor((i + 1) * Math.random());
    temp = shuffled[index];
    shuffled[index] = shuffled[i];
    shuffled[i] = temp;
  }
  return shuffled.slice(min);
}
