<template>
  <div class="ranking-strategy">
    <n-card :bordered="false" class="proCard">
      <n-result status="info" title="功能建设中" description="盈利排行策略模板正在开发中，敬请期待">
        <template #icon>
          <div class="building-icon">
            <n-icon size="80" color="#1890ff">
              <ToolOutlined />
            </n-icon>
          </div>
        </template>

        <template #footer>
          <n-space vertical align="center" :size="24">
            <n-card style="max-width: 600px; text-align: left">
              <n-space vertical :size="16">
                <n-text strong style="font-size: 16px">🚀 即将推出的功能：</n-text>

                <n-space vertical :size="12">
                  <n-space align="start">
                    <n-tag type="success" size="small">1</n-tag>
                    <div>
                      <n-text strong>策略盈利排行榜</n-text>
                      <n-text depth="3" style="display: block; font-size: 13px">
                        展示平台上盈利最高的策略模板，按周、月、总盈利排名
                      </n-text>
                    </div>
                  </n-space>

                  <n-space align="start">
                    <n-tag type="success" size="small">2</n-tag>
                    <div>
                      <n-text strong>一键复制策略</n-text>
                      <n-text depth="3" style="display: block; font-size: 13px">
                        看到优秀策略可以一键复制到您的策略模板中使用
                      </n-text>
                    </div>
                  </n-space>

                  <n-space align="start">
                    <n-tag type="success" size="small">3</n-tag>
                    <div>
                      <n-text strong>策略创作者激励</n-text>
                      <n-text depth="3" style="display: block; font-size: 13px">
                        分享您的策略，其他用户使用时您可获得激励奖励
                      </n-text>
                    </div>
                  </n-space>

                  <n-space align="start">
                    <n-tag type="success" size="small">4</n-tag>
                    <div>
                      <n-text strong>策略回测数据</n-text>
                      <n-text depth="3" style="display: block; font-size: 13px">
                        查看每个策略的历史回测表现、最大回撤、夏普比率等
                      </n-text>
                    </div>
                  </n-space>
                </n-space>
              </n-space>
            </n-card>

            <n-space>
              <n-button @click="$router.push('/toogo/strategy/official')">
                <template #icon
                  ><n-icon><StarOutlined /></n-icon
                ></template>
                先看看官方策略
              </n-button>
              <n-button type="primary" @click="$router.push('/toogo/strategy/my')">
                <template #icon
                  ><n-icon><FolderOutlined /></n-icon
                ></template>
                管理我的策略
              </n-button>
            </n-space>
          </n-space>
        </template>
      </n-result>
    </n-card>

    <!-- 预览效果（占位） -->
    <n-card title="📊 盈利排行预览" style="margin-top: 16px" :bordered="false">
      <n-alert type="warning" style="margin-bottom: 16px">
        以下为示例数据，实际功能开发中...
      </n-alert>

      <n-data-table :columns="previewColumns" :data="previewData" :bordered="false" />
    </n-card>
  </div>
</template>

<script lang="ts" setup>
  import { h } from 'vue';
  import { NTag, NButton, NSpace, NText } from 'naive-ui';
  import { ToolOutlined, StarOutlined, FolderOutlined } from '@vicons/antd';

  // 预览列
  const previewColumns = [
    {
      title: '排名',
      key: 'rank',
      width: 80,
      render: (row: any) =>
        h(
          NTag,
          { type: row.rank <= 3 ? 'warning' : 'default', size: 'small' },
          () => `#${row.rank}`,
        ),
    },
    { title: '策略名称', key: 'name' },
    { title: '创建者', key: 'creator' },
    {
      title: '交易对',
      key: 'symbol',
      render: (row: any) => h(NTag, { type: 'info', size: 'small' }, () => row.symbol),
    },
    {
      title: '本月盈利',
      key: 'profit',
      render: (row: any) => h(NText, { type: 'success', strong: true }, () => `+${row.profit}%`),
    },
    { title: '使用人数', key: 'users' },
    {
      title: '操作',
      key: 'action',
      render: () =>
        h(NSpace, {}, () => [
          h(NButton, { size: 'small', disabled: true }, () => '查看详情'),
          h(NButton, { size: 'small', type: 'primary', disabled: true }, () => '一键复制'),
        ]),
    },
  ];

  // 预览数据
  const previewData = [
    {
      rank: 1,
      name: 'BTC趋势跟踪Pro',
      creator: 'TradeM***',
      symbol: 'BTC-USDT',
      profit: '23.5',
      users: 156,
    },
    {
      rank: 2,
      name: 'ETH波段策略',
      creator: 'Crypto***',
      symbol: 'ETH-USDT',
      profit: '18.2',
      users: 89,
    },
    {
      rank: 3,
      name: '多币种平衡组合',
      creator: 'Smart***',
      symbol: '多币种',
      profit: '15.8',
      users: 234,
    },
    {
      rank: 4,
      name: 'SOL高波动策略',
      creator: 'Speed***',
      symbol: 'SOL-USDT',
      profit: '12.3',
      users: 67,
    },
    {
      rank: 5,
      name: 'DOGE网格策略',
      creator: 'Grid***',
      symbol: 'DOGE-USDT',
      profit: '9.7',
      users: 45,
    },
  ];
</script>

<style scoped lang="less">
  .ranking-strategy {
    .building-icon {
      width: 120px;
      height: 120px;
      background: linear-gradient(135deg, #e6f7ff 0%, #bae7ff 100%);
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      margin: 0 auto;
      animation: pulse 2s ease-in-out infinite;
    }

    @keyframes pulse {
      0%,
      100% {
        transform: scale(1);
      }
      50% {
        transform: scale(1.05);
      }
    }
  }
</style>
