<template>
  <PageWrapper title="图标选择器" content="通常用于，图标选择，默认装配了 Antd 和 Ionicons5 图标库">
    <n-card :bordered="false" class="mt-3">
      <div style="width: 350px"><IconPicker /></div>
    </n-card>
  </PageWrapper>
</template>

<script lang="ts" setup>
  import { IconPicker } from '@/components/IconPicker';
</script>

<style lang="less" scoped>
  .loading {
    position: relative;
  }
</style>
