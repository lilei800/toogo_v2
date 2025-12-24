<template>
  <div class="mb-4 text-center">
    <n-radio-group
      name="themeTypeGroup"
      v-model:value="settingStore.themeType"
      @update:value="themeTypeChange"
    >
      <n-radio-button key="pure" value="pure" label="纯色主题" />
      <n-radio-button key="skin" value="skin" label="主题皮肤" />
    </n-radio-group>
  </div>

  <div class="pt-4 drawer-setting-item align-items-top" v-if="settingStore.themeType === 'pure'">
    <div class="flex justify-center">
      <n-color-picker
        class="color-picker-item"
        :swatches="appThemeList"
        :default-value="designStore.appTheme"
        :actions="['confirm']"
        @confirm="colorPickerComplete"
      />
    </div>
  </div>

  <n-grid x-gap="12" y-gap="12" :cols="3" class="py-3" v-if="settingStore.themeType === 'skin'">
    <n-gi>
      <n-card
        embedded
        :class="{ 'skin-card-on': settingStore.themeSkin === 'blue-sky' }"
        class="skin-card"
        content-style="padding: 8px 0;"
        hoverable
        @click="setSkin('blue-sky')"
      >
        <div class="flex flex-col justify-center text-center">
          <img src="@/assets/images/skins/blue-sky-small.png" />
          <n-text class="pt-2">蓝色天空</n-text>
        </div>
      </n-card>
    </n-gi>
    <n-gi>
      <n-card
        :class="{ 'skin-card-on': settingStore.themeSkin === 'blue-christmas' }"
        class="skin-card"
        embedded
        content-style="padding: 8px;"
        hoverable
        @click="setSkin('blue-christmas')"
      >
        <div class="flex flex-col justify-center text-center">
          <img src="@/assets/images/skins/blue-christmas-small.png" />
          <n-text class="pt-2">蓝色圣诞</n-text>
        </div>
      </n-card>
    </n-gi>
    <n-gi>
      <n-card
        :class="{ 'skin-card-on': settingStore.themeSkin === 'blue-lattice' }"
        class="skin-card"
        embedded
        content-style="padding: 8px;"
        hoverable
        @click="setSkin('blue-lattice')"
      >
        <div class="flex flex-col justify-center text-center">
          <img src="@/assets/images/skins/blue-lattice-small.png" />
          <n-text class="pt-2">彩色云母</n-text>
        </div>
      </n-card>
    </n-gi>
    <n-gi>
      <n-card
        :class="{ 'skin-card-on': settingStore.themeSkin === 'pink-romantic' }"
        class="skin-card"
        embedded
        content-style="padding: 8px;"
        hoverable
        @click="setSkin('pink-romantic')"
      >
        <div class="flex flex-col justify-center text-center">
          <img src="@/assets/images/skins/pink-romantic-small.jpg" />
          <n-text class="pt-2">粉色浪漫</n-text>
        </div>
      </n-card>
    </n-gi>
    <n-gi>
      <n-card
        :class="{ 'skin-card-on': settingStore.themeSkin === 'green-mountain' }"
        class="skin-card"
        embedded
        content-style="padding: 8px;"
        hoverable
        @click="setSkin('green-mountain')"
      >
        <div class="flex flex-col justify-center text-center">
          <img src="@/assets/images/skins/green-mountain-small.jpg" />
          <n-text class="pt-2">翡翠绿峰</n-text>
        </div>
      </n-card>
    </n-gi>
    <n-gi>
      <n-card
        :class="{ 'skin-card-on': settingStore.themeSkin === 'paint-splashing' }"
        class="skin-card"
        embedded
        content-style="padding: 8px;"
        hoverable
        @click="setSkin('paint-splashing')"
      >
        <div class="flex flex-col justify-center text-center">
          <img src="@/assets/images/skins/paint-splashing-small.jpg" />
          <n-text class="pt-2">流光溢彩</n-text>
        </div>
      </n-card>
    </n-gi>
    <n-gi>
      <n-card
        :class="{ 'skin-card-on': settingStore.themeSkin === 'orange-bubble' }"
        class="skin-card"
        embedded
        content-style="padding: 8px;"
        hoverable
        @click="setSkin('orange-bubble')"
      >
        <div class="flex flex-col justify-center text-center">
          <img src="@/assets/images/skins/orange-bubble-small.jpg" />
          <n-text class="pt-2">香橙泡泡</n-text>
        </div>
      </n-card>
    </n-gi>
    <n-gi>
      <n-card
        :class="{ 'skin-card-on': settingStore.themeSkin === 'star-jellyfish' }"
        class="skin-card"
        embedded
        content-style="padding: 8px;"
        hoverable
        @click="setSkin('star-jellyfish')"
      >
        <div class="flex flex-col justify-center text-center">
          <img src="@/assets/images/skins/star-jellyfish-small.jpg" />
          <n-text class="pt-2">深空水母</n-text>
        </div>
      </n-card>
    </n-gi>
    <n-gi>
      <n-card
        :class="{ 'skin-card-on': settingStore.themeSkin === 'star-neon' }"
        class="skin-card"
        embedded
        content-style="padding: 8px;"
        hoverable
        @click="setSkin('star-neon')"
      >
        <div class="flex flex-col justify-center text-center">
          <img src="@/assets/images/skins/star-neon-small.jpg" />
          <n-text class="pt-2">星光霓虹</n-text>
        </div>
      </n-card>
    </n-gi>
  </n-grid>
</template>

<script lang="ts" setup>
  import { appThemeList } from '@/settings/designSetting';
  import { useDesignSettingStore } from '@/store/modules/designSetting';
  import { useProjectSettingStore } from '@/store/modules/projectSetting';
  import { ThemeSkin } from '/#/config';

  const designStore = useDesignSettingStore();
  const settingStore = useProjectSettingStore();

  const skinMapColour = {
    'blue-sky': '#358cf1',
    'blue-christmas': '#358cf1',
    'blue-lattice': '#1d63b9',
    'pink-romantic': '#ff677b',
    'green-mountain': '#4e9c48',
    'paint-splashing': '#09a7e1',
    'orange-bubble': '#eb591d',
    'star-jellyfish': '#3376c1',
    'star-neon': '#c6372e',
  };

  function colorPickerComplete(value) {
    designStore.appTheme = value;
  }

  function setSkin(value: ThemeSkin) {
    settingStore.themeSkin = value;
    designStore.appTheme = skinMapColour[value];
  }

  function themeTypeChange(value) {
    const htmlRoot = document.getElementById('htmlRoot');
    htmlRoot && htmlRoot.setAttribute('data-theme', 'light');
    designStore.setDarkTheme(false);
    settingStore.navTheme = 'light';
    if (value === 'skin') {
      settingStore.headerLucency = true;
      settingStore.sidebarLucency = true;
      settingStore.containerLucency = true;
    } else {
      settingStore.headerLucency = false;
      settingStore.sidebarLucency = false;
      settingStore.containerLucency = false;
    }
  }
</script>

<style lang="less" scoped>
  .color-picker-item {
    width: 50%;
    margin: auto;

    :deep(.n-color-picker-trigger) {
      border-radius: 0.5rem;
    }
  }
  .skin-card {
    &:hover {
      cursor: pointer;
      opacity: 0.9;
    }
    &-on {
      border-color: var(--n-color-target);
    }
    img {
      width: 70px;
      height: 50px;
      border-radius: 5px;
      display: inline-block;
      margin: auto;
    }
  }
</style>
