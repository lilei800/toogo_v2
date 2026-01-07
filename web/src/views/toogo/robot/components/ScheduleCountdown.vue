<template>
  <n-space :size="4" align="center" v-if="countdownText">
    <n-icon :component="icon" :size="14" />
    <n-text :type="type" :depth="2" style="font-size: 12px">
      {{ countdownText }}
    </n-text>
  </n-space>
</template>

<script setup lang="ts">
  import { ref, computed, onMounted, onUnmounted, watch } from 'vue';
  import { ClockCircleOutlined, AlertOutlined } from '@vicons/antd';

  interface Props {
    scheduleTime?: string;
    mode: 'start' | 'stop';
    robotStatus: number;
  }

  const props = defineProps<Props>();

  const countdown = ref('');
  let timer: ReturnType<typeof setInterval> | null = null;

  // 图标
  const icon = computed(() => {
    return props.mode === 'start' ? ClockCircleOutlined : AlertOutlined;
  });

  // 颜色类型
  const type = computed(() => {
    return props.mode === 'start' ? 'success' : 'warning';
  });

  // 倒计时文本
  const countdownText = computed(() => {
    if (!countdown.value) return '';
    return props.mode === 'start' ? `${countdown.value}后自动启动` : `${countdown.value}后自动停止`;
  });

  // 计算倒计时
  const calculateCountdown = () => {
    // 检查是否应该显示倒计时
    if (!props.scheduleTime) {
      countdown.value = '';
      return;
    }

    // 启动倒计时：只在未运行时显示
    if (props.mode === 'start' && props.robotStatus === 2) {
      countdown.value = '';
      return;
    }

    // 停止倒计时：只在运行中时显示
    if (props.mode === 'stop' && props.robotStatus !== 2) {
      countdown.value = '';
      return;
    }

    // 计算时间差
    const now = Date.now();
    const targetTime = new Date(props.scheduleTime).getTime();
    const diff = targetTime - now;

    // 已过期
    if (diff <= 0) {
      countdown.value = '';
      return;
    }

    // 格式化倒计时
    countdown.value = formatTime(diff);
  };

  // 格式化时间
  const formatTime = (milliseconds: number): string => {
    const seconds = Math.floor(milliseconds / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);

    if (days > 0) {
      return `${days}天${hours % 24}小时`;
    } else if (hours > 0) {
      return `${hours}小时${minutes % 60}分`;
    } else if (minutes > 0) {
      return `${minutes}分${seconds % 60}秒`;
    } else {
      return `${seconds}秒`;
    }
  };

  // 启动定时器
  const startTimer = () => {
    stopTimer();
    calculateCountdown();
    timer = setInterval(calculateCountdown, 1000);
  };

  // 停止定时器
  const stopTimer = () => {
    if (timer) {
      clearInterval(timer);
      timer = null;
    }
  };

  // 监听属性变化
  watch(
    () => [props.scheduleTime, props.robotStatus],
    () => {
      calculateCountdown();
    },
    { immediate: true },
  );

  // 生命周期
  onMounted(() => {
    startTimer();
  });

  onUnmounted(() => {
    stopTimer();
  });
</script>

<style scoped>
  /* 倒计时样式 */
</style>
