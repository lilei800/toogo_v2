<template>
  <PageWrapper title="卡片" content="BasicCard 基于 Card 和 Collapse 组件扩展" showFooter>
    <BasicCard title="基本使用" tooltip="这是一段提示内容"> 我是卡片内容 </BasicCard>

    <n-card title="手动控制展开/收起" class="mt-3">
      <n-switch v-model:value="show" class="mb-3">
        <template #checked> 收起 </template>
        <template #unchecked> 展开 </template>
      </n-switch>
      <BasicCard title="卡片" :show="show"> 手动控制展开/收起-卡片内容 </BasicCard>
    </n-card>

    <n-card title="触发展开区域" class="mt-3">
      <n-flex vertical class="mb-3">
        <n-flex>
          <n-tag v-model:checked="main" checkable> main </n-tag>
          <n-tag v-model:checked="arrow" checkable> arrow </n-tag>
        </n-flex>
      </n-flex>
      <BasicCard title="卡片" :triggers="triggers"> 触发展开区域-卡片内容 </BasicCard>
    </n-card>

    <n-card title="自定义展开" class="mt-3">
      <div class="mb-3">
        <n-text depth="3">
          如果只是想修改文案，可以通过 属性：spreadText 和 foldText 配置即可
        </n-text>
      </div>

      <BasicCard title="卡片" spreadText="显示内容" foldText="隐藏内容">
        自定义展开文案-卡片内容
        <template #collapse="{ expanded }">
          <div v-if="expanded" class="text-yellow-400"> 折叠 </div>
          <div v-else class="text-green-400"> 打开 </div>
        </template>
      </BasicCard>
    </n-card>

    <n-card title="自定义插槽" class="mt-3">
      <BasicCard title="卡片">
        自定义插槽-卡片内容
        <template #action>
          <n-flex>
            <n-button size="small" type="tertiary">复制</n-button>
            <n-button size="small" type="tertiary">编辑</n-button>
            <n-button size="small" type="tertiary">查看</n-button>
          </n-flex>
        </template>
      </BasicCard>
    </n-card>

    <BasicCard
      title="显示分割条"
      tooltip="这是一段提示内容"
      showDivider
      :showCollapse="false"
      class="mt-3"
      :segmented="{
        content: true,
        footer: 'soft',
      }"
    >
      我是卡片内容
    </BasicCard>
  </PageWrapper>
</template>

<script lang="ts" setup>
  import { BasicCard } from '@/components/Card/index';
  import { ref, computed } from 'vue';

  const show = ref(false);

  const main = ref(true);
  const arrow = ref(true);

  const triggers = computed(() => {
    const areas: Array<'main' | 'arrow'> = [];
    if (main.value) areas.push('main');
    if (arrow.value) areas.push('arrow');
    return areas;
  });
</script>
