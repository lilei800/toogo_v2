<template>
  <PageWrapper title="文件管理" showFooter>
    <template #headerContent>
      <PicFile
        ref="picFileRef"
        :loading="loading"
        :fileList="getPaginationList"
        :groupList="groups"
      >
        <template #action-right>
          <n-input type="text" placeholder="搜索文件" class="mr-3">
            <template #prefix>
              <n-icon :component="SearchOutlined" />
            </template>
          </n-input>
        </template>
        <template #upload>
          <n-upload action="https://www.mocky.io/v2/5e4bafc63100007100d8b70f">
            <n-button type="primary">
              上传文件
              <template #icon>
                <n-icon>
                  <UploadOutlined />
                </n-icon>
              </template>
            </n-button>
          </n-upload>
        </template>
      </PicFile>
      <n-divider />
      <div class="flex justify-center mt-6 picker-pagination">
        <n-pagination
          show-size-picker
          :page-sizes="[2, 4, 6]"
          v-model:page="currentPage"
          :pageSize="pageSize"
          :show-total="(total) => `共 ${total} 个`"
          :item-count="getTotal"
          @update:page="handlePageChange"
          @update:page-size="handlePageSizeChange"
        />
      </div>
    </template>
  </PageWrapper>
</template>

<script lang="ts" setup>
  import { ref } from 'vue';
  import { UploadOutlined, SearchOutlined } from '@vicons/antd';
  import { PicFile, PicFileItem } from '@/components/PicFile/index';
  import { usePagination } from '@/hooks/web/usePagination';

  const picFileRef = ref();
  const loading = ref(false);
  const currentPage = ref<number>(1);
  const pageSize = ref<number>(4);

  const groups = ref([
    { name: '图片', id: 1, quantity: 58 },
    { name: '文件', id: 2, quantity: 23 },
    { name: '封面', id: 3, quantity: 16 },
  ]);

  const files = ref<PicFileItem[]>([
    {
      id: 1,
      name: 'avatar-1.jpg',
      url: 'https://assets.naiveadmin.com/assets/avatar/avatar-1.jpg',
      type: 'jpg',
      size: '16.8KB',
      createDate: '2025-02-01 19:15:21',
    },
    {
      id: 2,
      name: 'avatar-2.png',
      url: 'https://assets.naiveadmin.com/assets/avatar/avatar-2.jpg',
      type: 'jpg',
      size: '29.5KB',
      createDate: '2025-01-15 22:18:39',
    },
    {
      id: 3,
      name: 'avatar-3.png',
      url: 'https://assets.naiveadmin.com/assets/avatar/avatar-3.jpg',
      type: 'jpg',
      size: '39.5KB',
      createDate: '2025-01-10 11:12:19',
    },
    {
      id: 4,
      name: 'avatar-4.jpg',
      url: 'https://assets.naiveadmin.com/assets/avatar/avatar-4.jpg',
      type: 'jpg',
      size: '53.5KB',
      createDate: '2025-01-16 13:11:49',
    },
    {
      id: 5,
      name: 'avatar-5.jpg',
      url: 'https://assets.naiveadmin.com/assets/avatar/avatar-5.jpg',
      type: 'jpg',
      size: '36.5KB',
      createDate: '2025-02-15 12:00:05',
    },
    {
      id: 6,
      name: 'avatar-6.jpg',
      url: 'https://assets.naiveadmin.com/assets/avatar/avatar-6.jpg',
      type: 'jpg',
      size: '56.8KB',
      createDate: '2025-01-14 18:16:22',
    },
    {
      id: 7,
      name: 'article-1.jpeg',
      url: 'https://assets.naiveadmin.com/assets/article/1.jpeg',
      type: 'jpeg',
      size: '36.5KB',
      createDate: '2025-01-19 18:19:05',
    },
    {
      id: 8,
      name: 'article-2.jpeg',
      url: 'https://assets.naiveadmin.com/assets/article/2.jpeg',
      type: 'jpeg',
      size: '26.5KB',
      createDate: '2025-01-14 17:12:19',
    },
    {
      id: 9,
      name: 'article-3.jpeg',
      url: 'https://assets.naiveadmin.com/assets/article/3.jpeg',
      type: 'jpeg',
      size: '86.5KB',
      createDate: '2025-01-18 12:12:19',
    },
  ]);

  const getPagination = (list) => {
    return usePagination(list, pageSize.value);
  };

  const { getPaginationList, getTotal, setCurrentPage, setPageSize } = getPagination(files.value);

  function handlePageChange(page: number) {
    currentPage.value = page;
    setCurrentPage(page);
    picFileRef.value.resetCheckedKeys();
  }

  function handlePageSizeChange(size: number) {
    pageSize.value = size;
    setPageSize(size);
    picFileRef.value.resetCheckedKeys();
  }
</script>
