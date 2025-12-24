import { useGlobSetting } from '@/hooks/setting';
import { createInstance } from './alova';
import { defaultConfig } from './config';

const globSetting = useGlobSetting();
const { apiUrl, urlPrefix } = globSetting;

// 创建 Alova 实例， 如果多服务，可以创建多个实例，区分不同配置信息
export const Alova = createInstance({
  ...defaultConfig,
  baseURL: apiUrl,
  urlPrefix,
});
