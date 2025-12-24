<template>
  <!-- 顶部 -->
  <div class="pic-files-top">
    <n-flex align="center">
      <div><slot name="upload"></slot></div>
      <n-divider vertical />
      <n-h3 class="mt-0 mb-0">文件（共{{ getTotalNum }}个）</n-h3>
      <n-text depth="3" v-if="tips">{{ tips }}</n-text>
    </n-flex>
  </div>
  <!-- 分组 -->
  <n-el class="pic-files-group" v-if="showGroup && groupList.length">
    <n-flex align="center" justify="space-between">
      <div class="flex items-center pic-files-group-left">
        <span class="mr-4">分组</span>
        <ul class="flex">
          <li
            class="mx-2 cursor-pointer"
            :class="{ 'group-on': groupId === item.id }"
            v-for="(item, index) in groupList"
            :key="index"
            @click="handleGroupToggle(item.id)"
          >
            <n-text depth="3">{{ item.name }}（{{ item.quantity }}）</n-text>
          </li>
        </ul>
        <n-divider vertical class="mr-6" />
        <n-button text type="primary" @click="handleGroupAdd"> 新建分组 </n-button>
      </div>
    </n-flex>
  </n-el>

  <!-- 全选 -->
  <div class="mt-3">
    <n-alert :bordered="false" :show-icon="false" type="default">
      <n-flex align="center" justify="space-between">
        <div>
          <n-checkbox
            :indeterminate="getIndeterminate"
            v-model:checked="checkAll"
            @update:checked="handleCheckAll"
          >
            <span v-if="!checkAll && getCheckNum === 0">全选</span>
            <span v-else
              >已选择 <strong> {{ getCheckNum }} </strong> 个文件</span
            >
          </n-checkbox>
          <!-- 判断插槽 action-left 是否存在 -->
          <template v-if="$slots['action-left']">
            <slot name="action-left"></slot>
          </template>
          <!-- 不存在 -->
          <template v-else>
            <n-button size="small" ghost type="error" v-if="getCheckNum > 0" @click="handleDelete"
              >删除文件</n-button
            >
          </template>
        </div>
        <!-- 右侧操作区域 -->
        <div class="flex items-center">
          <slot name="action-right"></slot>
          <n-tabs type="segment" animated v-model:value="displayType">
            <n-tab name="grid">
              <n-flex align="center" :size="[3, 0]">
                <n-icon>
                  <GridOutline />
                </n-icon>
                <span>网格</span>
              </n-flex>
            </n-tab>
            <n-tab name="table">
              <n-flex align="center" :size="[3, 0]">
                <n-icon>
                  <TableOutlined />
                </n-icon>
                <span>表格</span>
              </n-flex>
            </n-tab>
          </n-tabs>
        </div>
      </n-flex>
    </n-alert>
  </div>
  <!-- 文件内容区域 -->
  <n-spin :show="_loading">
    <div
      class="pic-files-container"
      @mousedown="startSelection"
      @mousemove="updateSelection"
      @mouseup="endSelection"
    >
      <div v-show="isSelecting" class="selection-box" :style="selectionBoxStyle"></div>
      <ul class="pic-files-picker" v-if="displayType === 'grid'">
        <li
          class="pic-files-picker-item"
          v-for="(item, index) in fileList"
          :key="index"
          :class="{ 'checked-item': checkedKeys.has(item.id) }"
          :data-id="item.id"
        >
          <div class="picker-item-file">
            <!-- 图片 -->
            <n-image v-if="isType(item.type, 'image')" :src="item.url" />
            <!-- pdf -->
            <a v-else-if="item.type === 'pdf'" :href="item.url" target="_blank">
              <img src="@/assets/images/ic-pdf.svg" />
            </a>
            <!-- 视频 -->
            <a v-else-if="isType(item.type, 'video')" :href="item.url" target="_blank">
              <img src="@/assets/images/ic-video.svg" />
            </a>
            <!-- 音频 -->
            <a v-else-if="isType(item.type, 'audio')" :href="item.url" target="_blank">
              <img src="@/assets/images/ic-video.svg" />
            </a>
            <!-- 文档 -->
            <a v-else-if="isType(item.type, 'document')" :href="item.url" target="_blank">
              <img src="@/assets/images/ic-video.svg" />
            </a>
            <!-- 其他文件 -->
            <a :href="item.url" v-else target="_blank">
              <img src="@/assets/images/ic-file.svg" />
            </a>
          </div>
          <strong>{{ item.name }}</strong>
          <div class="flex items-center justify-end item-operate" @click="checkedFile(item.id)">
            <n-icon
              class="check-icon"
              :size="24"
              :color="
                checkedKeys.has(item.id) ? 'rgba(53, 140, 241, 1)' : 'rgba(53, 140, 241, 0.6)'
              "
              style="border-radius: 50%"
              :style="
                checkedKeys.has(item.id)
                  ? 'background: rgba(255, 255, 255, 0.5)'
                  : 'background: rgba(255, 255, 255, 0.5)'
              "
            >
              <CheckmarkCircleSharp />
            </n-icon>
          </div>
        </li>
      </ul>
      <n-data-table
        v-else
        class="pic-files-table"
        :row-key="(row) => row.id"
        :columns="columns"
        :data="fileList"
        :pagination="false"
        :bordered="false"
        :row-props="rowProps"
        v-model:checked-row-keys="getTableCheckedKeys"
        @update:checked-row-keys="handleTableCheck"
      />
    </div>
  </n-spin>
</template>

<script lang="ts" setup>
  import { NButton, useDialog } from 'naive-ui';
  import { ref, computed, watch } from 'vue';
  import { CheckmarkCircleSharp, GridOutline } from '@vicons/ionicons5';
  import { TableOutlined } from '@vicons/antd';
  import { picFileProps } from './props';

  const dialog = useDialog();
  const checkAll = ref(false);
  const groupId = ref(null);
  const displayType = ref('grid');
  const checkedKeys = ref(new Set<number | string>());
  const isSelecting = ref(false);
  const selectionBox = ref({ x: 0, y: 0, width: 0, height: 0 });
  const startPoint = ref({ x: 0, y: 0 });
  const columns = [
    {
      type: 'selection',
    },
    {
      title: '文件名',
      key: 'name',
    },
    {
      title: '文件类型',
      key: 'type',
      width: 150,
    },
    {
      title: '大小',
      key: 'size',
      width: 150,
      defaultSortOrder: 'ascend',
      sorter: 'default',
    },
    {
      title: '创建时间',
      key: 'createDate',
      defaultSortOrder: 'ascend',
      sorter: 'default',
      width: 160,
    },
  ];

  const emit = defineEmits(['delete', 'group-change', 'group-add']);

  const props = defineProps({
    ...picFileProps,
  });

  const _loading = ref(props.loading);

  watch(
    () => props.loading,
    (value) => {
      _loading.value = value ?? false;
    },
    { immediate: true },
  );

  const selectionBoxStyle = computed(() => ({
    left: `${selectionBox.value.x}px`,
    top: `${selectionBox.value.y}px`,
    width: `${selectionBox.value.width}px`,
    height: `${selectionBox.value.height}px`,
  }));

  const getTableCheckedKeys = computed(() => {
    return [...checkedKeys.value];
  });

  const getCheckNum = computed(() => {
    return checkedKeys.value.size;
  });

  const getTotalNum = computed(() => {
    return props.fileList.length;
  });

  const getIndeterminate = computed(() => {
    if (checkAll.value) return false;

    if (checkedKeys.value.size > 0 && checkedKeys.value.size < props.fileList.length) return true;

    return false;
  });

  function rowProps(row) {
    return {
      'data-id': row.id,
    };
  }

  function handleTableCheck(keys) {
    checkedKeys.value = new Set(keys);
    updateCheckAll();
  }

  function handleDelete() {
    if (!checkedKeys.value.size) return;

    const nameStr = props.fileList
      .filter((item) => checkedKeys.value.has(item.id))
      .map((re) => re.name)
      .join('、');

    dialog.warning({
      title: '提示',
      content: `确定要删除 ${nameStr} 吗？`,
      positiveText: '确定',
      negativeText: '取消',
      showIcon: false,
      draggable: true,
      style: 'width: 380px',
      onPositiveClick: () => {
        // 执行删除逻辑
        emit('delete', checkedKeys.value);
      },
    });
  }

  function updateCheckAll() {
    checkAll.value = checkedKeys.value?.size === props.fileList.length;
  }

  function handleGroupToggle(id) {
    groupId.value = id;
    emit('group-change', id);
  }

  function handleGroupAdd() {
    emit('group-add');
  }

  function handleCheckAll() {
    if (checkedKeys.value?.size > 0 && checkedKeys.value?.size === props.fileList.length) {
      checkedKeys.value.clear();
      checkAll.value = false;
      return;
    }
    checkedKeys.value = new Set(props.fileList?.map((item) => item.id));
    checkAll.value = true;
  }

  function checkedFile(id: number | string) {
    if (checkedKeys.value.has(id)) {
      checkedKeys.value.delete(id);
    } else {
      checkedKeys.value.add(id);
    }
    updateCheckAll();
  }

  function isType(type, category) {
    const types = {
      document: ['docx', 'doc', 'pdf', 'xlsx', 'xls', 'pptx', 'ppt', 'csv', 'txt'],
      audio: ['mp3', 'wav', 'aac', 'flac', 'ogg'],
      video: ['mp4', 'avi', 'mkv', 'mov', 'wmv', 'webm', 'm4v', 'ogv'],
      image: ['jpg', 'png', 'gif', 'jpeg', 'bmp', 'tiff', 'webp', 'svg'],
    };
    return types[category].includes(type);
  }

  function startSelection(event: MouseEvent) {
    isSelecting.value = true;
    startPoint.value = { x: event.clientX, y: event.clientY };
    selectionBox.value = { x: event.clientX, y: event.clientY, width: 0, height: 0 };
  }

  function updateSelection(event: MouseEvent) {
    if (!isSelecting.value) return;
    const currentPoint = { x: event.clientX, y: event.clientY };
    selectionBox.value = {
      x: Math.min(startPoint.value.x, currentPoint.x),
      y: Math.min(startPoint.value.y, currentPoint.y),
      width: Math.abs(startPoint.value.x - currentPoint.x),
      height: Math.abs(startPoint.value.y - currentPoint.y),
    };
    updateSelectedItems();
  }

  function endSelection() {
    isSelecting.value = false;
  }

  function updateSelectedItems() {
    const items = document.querySelectorAll('.pic-files-picker-item, .pic-files-table tr');
    items.forEach((item) => {
      const rect = item.getBoundingClientRect();
      const id = Number(item.getAttribute('data-id'));
      if (
        rect.left < selectionBox.value.x + selectionBox.value.width &&
        rect.right > selectionBox.value.x &&
        rect.top < selectionBox.value.y + selectionBox.value.height &&
        rect.bottom > selectionBox.value.y
      ) {
        checkedKeys.value.add(id);
      } else {
        checkedKeys.value.delete(id);
      }
    });
    updateCheckAll();
  }

  function resetCheckedKeys() {
    checkedKeys.value.clear();
  }

  defineExpose({
    resetCheckedKeys,
  });
</script>

<style lang="less" scoped>
  .pic-files {
    &-group {
      margin: 5px 0;
      &-left {
        ul li {
          line-height: 30px;
          padding: 0 15px;
          border-radius: 15px;
          border: 1px solid transparent;
          text-align: center;
        }
        .group-on {
          border-color: var(--primary-color);
          span {
            color: var(--primary-color);
          }
        }
      }
    }
  }

  .pic-files-container {
    padding-top: 20px;
    min-height: 450px;
    position: relative;
    user-select: none;
    .pic-files-picker {
      display: flex;
      flex-wrap: wrap;
      &-item {
        width: 120px;
        margin: 0 15px 15px 0;
        position: relative;
        float: none;
        text-align: center;
        padding: 10px;
        border-radius: 5px;
        strong {
          display: block;
          width: auto;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
          word-wrap: normal;
          font-weight: 400;
          padding: 8px 0 0;
          line-height: 20px;
        }
        &:hover {
          background-color: rgba(53, 140, 241, 0.2);
          .item-operate {
            display: flex !important;
          }
          .item-operate-right {
            display: block;
          }
          strong {
            color: var(--primary-color);
          }
        }
        .picker-item-file {
          display: block;
          width: 90px;
          height: 90px;
          border-radius: 3px;
          overflow: hidden;
          position: relative;
          margin: auto;
          img {
            height: 100%;
          }
        }
        .item-operate {
          padding: 5px 6px;
          position: absolute;
          top: 0;
          left: 0;
          width: 100%;
          display: none !important;
          :deep(.check-icon) {
            cursor: pointer;
            &:hover {
              color: rgba(53, 140, 241, 1) !important;
            }
          }
          .operate-circle {
            background-color: rgba(0, 0, 0, 0.3);
            width: 30px;
            height: 30px;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
          }
          &-right {
            display: none;
          }
        }
      }
      .checked-item {
        background-color: rgba(53, 140, 241, 0.2);
        .item-operate {
          display: flex !important;
        }
      }
    }
  }

  .selection-box {
    position: fixed;
    border: 1px dashed #3399ff;
    background-color: rgba(51, 153, 255, 0.2);
    pointer-events: none;
    z-index: 99;
  }
</style>
