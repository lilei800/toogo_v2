import { computed, unref } from 'vue';
import { keysOf, Simplify, simplyOmit } from './utils';

export function useOmitProps<T extends object, K extends keyof T>(
  props: T,
  excludeProps: Record<K, any>,
) {
  const excludePropKeys = keysOf(excludeProps);
  return computed(() => {
    const unwrappedProps = unref(props);
    return simplyOmit(unwrappedProps, excludePropKeys) as Simplify<Omit<T, K>>;
  });
}
