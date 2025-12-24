// 过滤滤掉不需要打印的列
export function filterColumns(columns) {
  if (!columns) {
    return [];
  }
  return columns.filter(
    (item) => !['selection', 'action'].includes(item.key || item.type) && item.print !== false,
  );
}

export function createSerial(data, checkedRowKeys, rowKey) {
  const safeData = data || [];
  const safeCheckedRowKeys = checkedRowKeys || [];
  const safeRowKey = rowKey || 'id'; // 提供一个默认值

  const checkedRowKeySet = new Set(safeCheckedRowKeys);

  const list = safeData.map((item, index) => {
    return { ...item, serial: index + 1 };
  });

  return safeCheckedRowKeys.length > 0
    ? list.filter((item) => checkedRowKeySet.has(item[safeRowKey]))
    : list;
}
