// Demo API used by a number of template/demo pages (table, tableSelect, excel export, etc).
// Provide a simple in-memory implementation so production builds can succeed.

export interface TableItem {
  id: number;
  name: string;
  age: number;
  address: string;
  sex: 'male' | 'female';
  status: 'open' | 'close' | 'refuse';
  avatar?: string;
  account?: string;
  mobile?: string;
  email?: string;
  role?: string;
  createDate?: string;
}

export const sexMap: Record<string, string> = {
  male: '男',
  female: '女',
};

export const statusMap: Record<string, string> = {
  open: '启用',
  close: '禁用',
  refuse: '拒绝',
};

const db: TableItem[] = Array.from({ length: 57 }).map((_, i) => {
  const sex = i % 2 === 0 ? 'male' : 'female';
  const status: TableItem['status'] = i % 7 === 0 ? 'refuse' : i % 3 === 0 ? 'close' : 'open';
  return {
    id: i + 1,
    name: `演示数据${i + 1}`,
    age: 18 + (i % 20),
    address: `演示地址 - ${i + 1}`,
    sex,
    status,
    avatar: '',
    account: `demo_${i + 1}`,
    mobile: `1380000${String(1000 + i).slice(-4)}`,
    email: `demo${i + 1}@example.com`,
    role: i % 4 === 0 ? '管理员' : '用户',
    createDate: new Date(Date.now() - i * 86400000).toISOString().slice(0, 19).replace('T', ' '),
  };
});

function paginate<T>(arr: T[], page: number, pageSize: number) {
  const start = (page - 1) * pageSize;
  return arr.slice(start, start + pageSize);
}

export async function getTableList(params: { page?: number; pageSize?: number; [k: string]: any }) {
  const page = Number(params?.page ?? 1) || 1;
  const pageSize = Number(params?.pageSize ?? 10) || 10;

  let rows = [...db];
  if (params?.name) {
    const kw = String(params.name);
    rows = rows.filter((r) => r.name.includes(kw));
  }
  if (params?.address) {
    const kw = String(params.address);
    rows = rows.filter((r) => r.address.includes(kw));
  }

  const itemCount = rows.length;
  const pageCount = Math.max(1, Math.ceil(itemCount / pageSize));
  const list = paginate(rows, Math.min(page, pageCount), pageSize);
  return { page, pageSize, list, pageCount, itemCount };
}

export async function getTableSelectList(params: {
  page?: number;
  pageSize?: number;
  [k: string]: any;
}) {
  return await getTableList(params);
}
