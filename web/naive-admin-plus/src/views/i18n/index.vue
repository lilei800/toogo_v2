<script lang="ts" setup>
  import { useLocale } from '@/locales/useLocale';
  import { useI18n } from '@/hooks/web/useI18n';

  const { getLocale } = useLocale();
  const { t } = useI18n();
</script>

<template>
  <PageWrapper :title="t('common.i18nText')">
    <template #headerContent>
      <n-flex gap="middle" vertical>
        <n-text strong>当前语言：{{ getLocale }}</n-text>
        <n-text depth="3"
          >点击右上角的多语言切换按钮，尝试切换不同的语言，下面的内容会自动变化：</n-text
        >
        <n-text depth="3">1. {{ t('common.welcomeText') }}!</n-text>
        <n-text depth="3">2. NaiveUi 组件内部也会自动切换多语言：</n-text>
        <n-date-picker style="width: 180px" />
        <n-text depth="3">3. 菜单及系统页面：</n-text>
        <n-alert type="info">
          <span
            >当前在线预览环境 <n-text strong>多语言配置</n-text>，只对
            <n-text strong>多语言菜单</n-text>，做了配置，其它尚未配置。</span
          >
        </n-alert>
      </n-flex>
    </template>
  </PageWrapper>
</template>
