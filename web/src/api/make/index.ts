// Demo API for AdvancedTable example page.
// This project includes some template/demo pages that expect these functions.
// Provide a small in-memory implementation so `pnpm build` can succeed.

type MakeRow = Record<string, any> & { id: number };

const db: MakeRow[] = Array.from({ length: 23 }).map((_, i) => ({
  id: i + 1,
  name: `演示用户${i + 1}`,
  mobile: 13800000000 + i,
  address: '演示地址',
  date: Date.now(),
  beginTime: Date.now(),
  endTime: Date.now(),
  avatar: '',
}));

function paginate<T>(arr: T[], page: number, pageSize: number) {
  const start = (page - 1) * pageSize;
  return arr.slice(start, start + pageSize);
}

export async function makeList(params: { page?: number; pageSize?: number; [k: string]: any }) {
  const page = Number(params?.page ?? 1) || 1;
  const pageSize = Number(params?.pageSize ?? 10) || 10;

  // very light filtering
  let rows = [...db];
  if (params?.name) {
    const kw = String(params.name);
    rows = rows.filter((r) => String(r.name).includes(kw));
  }
  if (params?.address) {
    const kw = String(params.address);
    rows = rows.filter((r) => String(r.address).includes(kw));
  }

  const itemCount = rows.length;
  const pageCount = Math.max(1, Math.ceil(itemCount / pageSize));
  const list = paginate(rows, Math.min(page, pageCount), pageSize);

  return { page, pageSize, list, pageCount, itemCount };
}

export async function makeAdd(row: Partial<MakeRow>) {
  const id = (db.at(-1)?.id ?? 0) + 1;
  db.push({ ...(row as any), id } as MakeRow);
  return { id };
}

export async function makeEdit(row: Partial<MakeRow> & { id: number }) {
  const idx = db.findIndex((r) => r.id === row.id);
  if (idx >= 0) {
    db[idx] = { ...db[idx], ...(row as any) };
  }
  return true;
}

export async function makeDelete(row: Partial<MakeRow> & { id: number }) {
  const idx = db.findIndex((r) => r.id === row.id);
  if (idx >= 0) {
    db.splice(idx, 1);
  }
  return true;
}
