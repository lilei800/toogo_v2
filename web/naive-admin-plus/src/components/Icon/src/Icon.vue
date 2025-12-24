<template>
  <n-icon :component="iconCompoent" />
</template>

<script lang="ts" setup>
  import { computed } from 'vue';
  import * as AntdIcons from '@vicons/antd';
  import * as Ionicons5Icons from '@vicons/ionicons5';

  const props = defineProps({
    icon: {
      type: String,
      required: true,
    },
  });

  function deletePrefix(name: string) {
    return name.indexOf(':') != -1 ? name.split(':')[1] : name;
  }

  function getPrefix(name: string) {
    return name.indexOf(':') != -1 ? name.split(':')[0] : name;
  }

  const iconCompoent = computed(() => {
    const keyName = getPrefix(props.icon);
    if (keyName === '') {
      console.error('Icons are not within the agreed scope');
      return;
    }
    const iconName = deletePrefix(props.icon);
    if (keyName === 'antd') {
      return AntdIcons[iconName] as any;
    }
    if (keyName === 'ionicons5') {
      return Ionicons5Icons[iconName] as any;
    }
    return [];
  });
</script>
