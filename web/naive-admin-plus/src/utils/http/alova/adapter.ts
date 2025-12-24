import { axiosMockResponse, axiosRequestAdapter } from '@alova/adapter-axios';
import { createAlovaMockAdapter } from '@alova/mock';
import { defaultMockConfig } from './config';
import mocks from './mocks';
import { useLocalSetting } from '@/hooks/setting';

const { useMock, loggerMock } = useLocalSetting();

const mockAdapterOptions = {
  ...defaultMockConfig,
  enable: useMock,
  httpAdapter: axiosRequestAdapter(),
  ...(useMock ? axiosMockResponse : {}),
  ...(loggerMock ? {} : { mockRequestLogger: false }),
};

export const mockAdapter = createAlovaMockAdapter([...mocks], mockAdapterOptions);
