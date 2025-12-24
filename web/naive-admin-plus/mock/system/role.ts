import { defineMock } from '@alova/mock';
import { resultSuccess, doCustomTimes } from '../util';
import { faker } from '@faker-js/faker';
import dayjs from 'dayjs';
function getMenuKeys() {
  const keys = ['dashboard', 'console', 'workplace', 'basic-form', 'step-form', 'detail'];
  const newKeys: string[] = [];
  doCustomTimes(parseInt((Math.random() * 6).toString()), () => {
    const key: string = keys[Math.floor(Math.random() * keys.length)];
    newKeys.push(key);
  });
  return Array.from(new Set(newKeys));
}

const roleList = (pageSize) => {
  const result: any[] = [];
  doCustomTimes(pageSize, () => {
    result.push({
      id: faker.string.numeric(4),
      name: faker.person.firstName(),
      //explain: faker.string.numeric(4),
      explain: faker.lorem.paragraph(1),
      isDefault: faker.helpers.arrayElement([true, false]),
      menu_keys: getMenuKeys(),
      createDate: dayjs(faker.date.anytime()).format('YYYY-MM-DD HH:mm'),
      status: faker.helpers.arrayElement(['normal', 'enable', 'disable']),
    });
  });
  return result;
};

export default defineMock({
  '/api/role/list': ({ query }) => {
    const { page = 1, pageSize = 10 } = query;
    const list = roleList(Number(pageSize));
    return resultSuccess({
      page: Number(page),
      pageSize: Number(pageSize),
      pageCount: 30,
      total: 30 * Number(pageSize),
      list,
    });
  },
});
