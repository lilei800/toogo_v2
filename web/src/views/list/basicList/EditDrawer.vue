<template>
  <DrawerForm
    ref="drawerFormRef"
    :width="450"
    :drawerContent="{
      title: '编辑用户',
    }"
    positiveText="保存"
    :form="form"
    @submit="formSubmit"
  >
    <div class="mt-6">
      <BasicForm @register="register" />
    </div>
  </DrawerForm>
</template>

<script lang="ts" setup>
  import { DrawerForm } from '@/components/DrawerForm/index';
  import { BasicForm, FormSchema, useForm } from '@/components/Form';
  import { ref, nextTick } from 'vue';

  const schemas: FormSchema[] = [
    {
      field: 'name',
      component: 'NInput',
      label: '名称',
      rules: [{ required: true, message: '请输入名称', trigger: ['blur'] }],
    },
    {
      field: 'sex',
      component: 'NSelect',
      label: '性别',
      componentProps: {
        placeholder: '请选择性别',
        options: [
          {
            label: '男',
            value: 'male',
          },
          {
            label: '女',
            value: 'female',
          },
          ,
        ],
      },
    },
    {
      field: 'email',
      component: 'NInput',
      label: '邮箱',
      componentProps: {
        placeholder: '请输入邮箱',
      },
    },
    {
      field: 'city',
      component: 'NInput',
      label: '城市',
      componentProps: {
        placeholder: '请输入城市',
      },
    },
    {
      field: 'status',
      component: 'NSelect',
      label: '状态',
      componentProps: {
        placeholder: '请选择性别',
        options: [
          {
            label: '拒绝',
            value: 'refuse',
          },
          {
            label: '取消',
            value: 'close',
          },
          {
            label: '成功',
            value: 'success',
          },
        ],
      },
    },
  ];

  const drawerFormRef = ref();

  const [register, form] = useForm({
    gridProps: { cols: 1 },
    layout: 'horizontal',
    submitButtonText: '提交预约',
    showActionButtonGroup: false,
    schemas,
  });

  function formSubmit(values, done) {
    console.log('表单值', values);
    // 1、这里做表单 api 提交
    // 2、提交完关闭弹窗
    done(true);
  }

  function openDrawer() {
    nextTick(() => {
      drawerFormRef.value.openDrawer();
    });
  }

  defineExpose({
    openDrawer,
  });
</script>
