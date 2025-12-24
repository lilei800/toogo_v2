import { defineMock } from '@alova/mock';
import { resultSuccess } from '../util';
import { faker } from '@faker-js/faker';
import dayjs from 'dayjs';

const dictionaryList = [
  {
    id: faker.string.numeric(4),
    label: '预约事项',
    key: 'makeMatter',
    children: [
      {
        id: faker.string.numeric(4),
        label: '初次预约',
        key: 'theMake',
      },
      {
        id: faker.string.numeric(4),
        label: '多次预约',
        key: 'towMake',
      },
    ],
  },
  {
    id: faker.string.numeric(4),
    label: '注册来源',
    key: 'registeredSource',
  },
];

const dictionaryItems = () => {
  return [
    {
      key: 'registeredSource',
      values: [
        {
          id: faker.string.numeric(4),
          value: 'baidu',
          label: '百度',
          order: faker.helpers.arrayElement([1, 2, 3, 4, 5, 6, 7, 8, 9, 10]),
          create_date: dayjs(faker.date.anytime()).format('YYYY-MM-DD HH:mm'),
        },
        {
          id: faker.string.numeric(4),
          value: 'weibo',
          label: '微博',
          order: faker.helpers.arrayElement([1, 2, 3, 4, 5, 6, 7, 8, 9, 10]),
          create_date: dayjs(faker.date.anytime()).format('YYYY-MM-DD HH:mm'),
        },
        {
          id: faker.string.numeric(4),
          value: 'weixin',
          label: '微信',
          order: faker.helpers.arrayElement([1, 2, 3, 4, 5, 6, 7, 8, 9, 10]),
          create_date: dayjs(faker.date.anytime()).format('YYYY-MM-DD HH:mm'),
        },
      ],
    },
    {
      key: 'theMake',
      parentKey: 'makeMatter',
      values: [
        {
          id: faker.string.numeric(4),
          value: 'examine',
          label: '检查',
          order: faker.helpers.arrayElement([1, 2, 3, 4, 5, 6, 7, 8, 9, 10]),
          create_date: dayjs(faker.date.anytime()).format('YYYY-MM-DD HH:mm'),
        },
        {
          id: faker.string.numeric(4),
          value: 'tooth',
          label: '拔牙',
          order: faker.helpers.arrayElement([1, 2, 3, 4, 5, 6, 7, 8, 9, 10]),
          create_date: dayjs(faker.date.anytime()).format('YYYY-MM-DD HH:mm'),
        },
      ],
    },
    {
      key: 'towMake',
      parentKey: 'makeMatter',
      values: [
        {
          id: faker.string.numeric(4),
          value: 'take',
          label: '拆线',
          order: faker.helpers.arrayElement([1, 2, 3, 4, 5, 6, 7, 8, 9, 10]),
          create_date: dayjs(faker.date.anytime()).format('YYYY-MM-DD HH:mm'),
        },
        {
          id: faker.string.numeric(4),
          value: 'periodontal',
          label: '牙周',
          order: faker.helpers.arrayElement([1, 2, 3, 4, 5, 6, 7, 8, 9, 10]),
          create_date: dayjs(faker.date.anytime()).format('YYYY-MM-DD HH:mm'),
        },
      ],
    },
  ];
};

const dictionaryInfo = (_, key: string) => {
  const list: any[] = [];
  dictionaryItems().forEach((item: any) => {
    if (item.key === key || item.parentKey === key) {
      list.push(item as any);
    }
  });
  const valuesList: any[] = [];
  list.forEach((item: any) => {
    item.values.map((values) => {
      valuesList.push(values);
    });
  });
  return valuesList;
};

export default defineMock({
  //字典列表
  '/api/dictionary/list': () => resultSuccess(dictionaryList),
  //字典详情
  '/api/dictionary/info': ({ query }) => {
    const { page = 1, pageSize = 10, key, keywords } = query;
    let list = dictionaryInfo(Number(pageSize), key);
    //实现搜索筛选
    if (keywords) {
      list = list.filter((item) => {
        return item.label.indexOf(keywords) != -1;
      });
    }
    const count = list.length > Number(pageSize) ? Math.ceil(list.length / Number(pageSize)) : 0;
    return resultSuccess({
      page: Number(page),
      pageSize: Number(pageSize),
      pageCount: count,
      itemCount: count * Number(pageSize),
      list,
    });
  },
});
