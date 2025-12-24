import { defineMock } from '@alova/mock';
import { resultSuccess } from '../util';

const classifyList = [
  {
    label: '新品',
    value: 'new',
  },
  {
    label: '爆款',
    value: 'hot',
  },
  {
    label: '推荐',
    value: 'rec',
  },
  {
    label: '促销',
    value: 'promotion',
  },
];

export default defineMock({
  '/api/classifyList': () => resultSuccess(classifyList),
});
