import type { Ref } from 'vue';
import type { BasicColumn, BasicTableProps, TableActionType } from '../types/table';
import { provide, inject, ComputedRef } from 'vue';

const key = Symbol('s-table');

interface TableContextProps extends BasicTableProps {
  data: any[];
  columns: BasicColumn[];
  printConfig: Recordable;
  checkedRowKeys: Ref<string[] | number[]>;
}

type Instance = TableActionType & {
  wrapRef: Ref<Nullable<HTMLElement>>;
  getBindValues: ComputedRef<Recordable>;
  checkedRowKeys: Ref<string[] | number[]>;
  isShowTable: any;
  tableId: string;
};

type RetInstance = Omit<Instance, 'getBindValues'> & {
  getBindValues: ComputedRef<TableContextProps>;
};

export function createTableContext(instance: Instance) {
  provide(key, instance);
}

export function useTableContext(): RetInstance {
  return inject(key) as RetInstance;
}
