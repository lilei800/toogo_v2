import { defineMock } from '@alova/mock';
import { faker } from '@faker-js/faker';
import { resultSuccess, doCustomTimes } from '../util';

const categoryList = (pageSize) => {
  const result: any[] = [];
  doCustomTimes(pageSize, () => {
    result.push({
      id: faker.string.numeric(4),
      name: `分类${faker.helpers.arrayElement([
        '零',
        '一',
        '二',
        '三',
        '四',
        '五',
        '六',
        '七',
        '八',
        '九',
        '十',
      ])}`,
    });
  });
  return result;
};

export default defineMock({
  '/api/category/list': ({ query }) => {
    const { page = 1, pageSize = 20 } = query;
    const list = categoryList(Number(pageSize));
    return resultSuccess({
      page: Number(page),
      pageSize: Number(pageSize),
      pageCount: 60,
      itemCount: 60 * Number(pageSize),
      list,
    });
  },
});
