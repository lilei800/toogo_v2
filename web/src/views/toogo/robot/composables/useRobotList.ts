/**
 * 机器人列表逻辑
 */
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useMessage } from 'naive-ui';
import { getRobotList, startRobot, stopRobot, deleteRobot } from '@/api/toogo/robot';

export interface Robot {
  id: number;
  robotName: string;
  symbol: string;
  exchange: string;
  platform: string;
  tradingPair: string;
  status: number;
  leverage: number;
  marginRatio: number;
  marginPercent: number;
  stopLossPercent: number;
  takeProfitRetracePercent: number;
  profitActivatePercent: number;
  totalPnl: number;
  consumedPower: number;
  createdAt: string;
  totalProfit: number;
  autoTradeEnabled: number;
  autoCloseEnabled: number;
  apiConfigId: number;
  scheduleStart?: string;  // 定时启动时间
  scheduleStop?: string;   // 定时停止时间
}

export interface SearchParams {
  status: number | null;
  platform: string | null;
}

export function useRobotList() {
  const message = useMessage();
  
  // 状态
  const loading = ref(false);
  const robotList = ref<Robot[]>([]);
  const total = ref(0);
  const searchParams = ref<SearchParams>({
    status: null,
    platform: null,
  });

  // 计算属性
  const runningCount = computed(() => robotList.value.filter(r => r.status === 2).length);
  const todayPnl = computed(() => robotList.value.reduce((sum, r) => sum + (r.totalPnl || 0), 0));
  const totalPnl = computed(() => robotList.value.reduce((sum, r) => sum + (r.totalProfit || 0), 0));
  const totalPower = computed(() => robotList.value.reduce((sum, r) => sum + (r.consumedPower || 0), 0));

  // 加载数据
  const loadData = async () => {
    loading.value = true;
    try {
      const params: Record<string, any> = {
        page: 1,
        pageSize: 100,
      };
      if (searchParams.value.status !== null) {
        params.status = searchParams.value.status;
      }
      if (searchParams.value.platform) {
        params.platform = searchParams.value.platform;
      }
      
      const res = await getRobotList(params);
      if (res.code === 0) {
        robotList.value = res.data?.list || [];
        total.value = res.data?.total || 0;
      }
    } catch (error) {
      console.error('加载机器人列表失败:', error);
    } finally {
      loading.value = false;
    }
  };

  // 启动机器人
  const handleStartRobot = async (robot: Robot) => {
    try {
      const res = await startRobot({ id: robot.id });
      if (res.code === 0) {
        message.success('启动成功');
        await loadData();
      } else {
        message.error(res.message || '启动失败');
      }
    } catch (error) {
      message.error('启动失败');
    }
  };

  // 停止机器人
  const handleStopRobot = async (robot: Robot) => {
    try {
      const res = await stopRobot({ id: robot.id });
      if (res.code === 0) {
        message.success('停止成功');
        await loadData();
      } else {
        message.error(res.message || '停止失败');
      }
    } catch (error) {
      message.error('停止失败');
    }
  };

  // 删除机器人
  const handleDeleteRobot = async (robot: Robot) => {
    if (robot.status === 2) {
      message.warning('请先停止机器人');
      return;
    }
    try {
      const res = await deleteRobot({ id: robot.id });
      if (res.code === 0) {
        message.success('删除成功');
        await loadData();
      } else {
        message.error(res.message || '删除失败');
      }
    } catch (error) {
      message.error('删除失败');
    }
  };

  // 初始化
  onMounted(() => {
    loadData();
  });

  return {
    // 状态
    loading,
    robotList,
    total,
    searchParams,
    
    // 计算属性
    runningCount,
    todayPnl,
    totalPnl,
    totalPower,
    
    // 方法
    loadData,
    handleStartRobot,
    handleStopRobot,
    handleDeleteRobot,
  };
}

// 状态选项
export const statusOptions = [
  { label: '全部', value: null },
  { label: '待启动', value: 1 },
  { label: '运行中', value: 2 },
  { label: '已停止', value: 3 },
];

// 平台选项
export const platformOptions = [
  { label: '全部', value: null },
  { label: 'Binance', value: 'binance' },
  { label: 'Bitget', value: 'bitget' },
  { label: 'OKX', value: 'okx' },
];

// 获取状态类型
export function getStatusType(status: number): 'success' | 'warning' | 'error' | 'default' {
  switch (status) {
    case 2: return 'success';
    case 1: return 'warning';
    case 3: return 'error';
    default: return 'default';
  }
}

// 获取状态文本
export function getStatusText(status: number): string {
  switch (status) {
    case 1: return '待启动';
    case 2: return '运行中';
    case 3: return '已停止';
    default: return '未知';
  }
}

