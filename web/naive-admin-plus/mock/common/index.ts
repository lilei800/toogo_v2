import { defineMock } from '@alova/mock';
import { resultSuccess } from '../util';

const stateList = [
  {
    label: '规划中',
    value: 1,
  },
  {
    label: '待处理',
    value: 2,
  },
  {
    label: '已完成',
    value: 3,
  },
  {
    label: '已拒绝',
    value: 4,
  },
];

export default defineMock({
  '/api/state_list': () => resultSuccess(stateList),
});
