<template>
  <div class="volatility-config-page">
    <n-card title="波动率配置" :bordered="false" class="proCard">
      <template #header-extra>
        <n-space>
          <n-button @click="openBatchEditModal()">
            <template #icon><n-icon :component="CopyOutlined" /></template>
            批量设置
          </n-button>
          <n-button @click="openEditModal()" type="primary">
            <template #icon><n-icon :component="PlusOutlined" /></template>
            新增配置
          </n-button>
        </n-space>
      </template>

      <!-- 筛选 -->
      <n-space style="margin-bottom: 16px">
        <n-input
          v-model:value="searchParams.symbol"
          placeholder="搜索交易对"
          clearable
          style="width: 200px"
          @keyup.enter="loadData"
        >
          <template #prefix><n-icon :component="SearchOutlined" /></template>
        </n-input>
        <n-select
          v-model:value="searchParams.isActive"
          :options="activeOptions"
          placeholder="状态"
          clearable
          style="width: 120px"
          @update:value="loadData"
        />
        <n-button @click="resetSearch">
          <template #icon><n-icon :component="ReloadOutlined" /></template>
          重置
        </n-button>
      </n-space>

      <!-- 市场状态说明卡片（适配新算法） -->
      <n-alert type="info" style="margin-bottom: 16px">
        <div class="market-state-tips">
          <div class="tip-title">
            <n-icon :component="InfoCircleOutlined" style="margin-right: 6px" />
            <strong>新算法市场状态判断规则</strong>
          </div>
          <div class="tip-content">
            <div class="tip-row">
              <n-tag type="info" size="small" :bordered="false">低波动市场</n-tag>
              <span>有效波动不足 → V < LowV</span>
            </div>
            <div class="tip-row">
              <n-tag type="warning" size="small" :bordered="false">震荡市场</n-tag>
              <span>有效波动足但不单边 → 其他情况（中等波动）</span>
            </div>
            <div class="tip-row">
              <n-tag type="error" size="small" :bordered="false">高波动市场</n-tag>
              <span>有效波动很大但乱扫 → V ≥ HighV 且 D < 0.4</span>
            </div>
            <div class="tip-row">
              <n-tag type="success" size="small" :bordered="false">趋势市场</n-tag>
              <span>有效波动足且单边 → V ≥ TrendV 且 D ≥ DThreshold</span>
            </div>
            <div
              class="tip-formula"
              style="
                margin-top: 12px;
                padding: 10px;
                background: #f8f9fa;
                border-left: 3px solid #18a058;
                border-radius: 4px;
                font-size: 12px;
              "
            >
              <div style="margin-bottom: 6px; font-weight: 600; color: #333">计算公式：</div>
              <div style="margin-bottom: 4px">V = (H - L) / delta（波动强度）</div>
              <div>D = (P - L) / (H - L) 或 (H - P) / (H - L)（方向一致性，0-1之间）</div>
            </div>
          </div>
          <div class="tip-note">
            <n-text depth="3" style="font-size: 12px">
              💡
              提示：新算法使用单根K线计算，响应快速。系统会根据交易对优先使用特定配置，未配置的交易对将使用全局配置。
            </n-text>
          </div>
        </div>
      </n-alert>

      <!-- 批量操作栏 -->
      <div
        v-if="selectedRowKeys.length > 0"
        style="margin-bottom: 16px; padding: 12px; background: #f0f9ff; border-radius: 4px"
      >
        <n-space align="center" justify="space-between">
          <n-text>
            已选择 <strong>{{ selectedRowKeys.length }}</strong> 项
          </n-text>
          <n-space>
            <n-button size="small" @click="batchEnable">
              <template #icon><n-icon :component="CheckCircleOutlined" /></template>
              批量启用
            </n-button>
            <n-button size="small" @click="batchDisable">
              <template #icon><n-icon :component="CloseCircleOutlined" /></template>
              批量禁用
            </n-button>
            <n-button size="small" type="error" @click="batchDelete">
              <template #icon><n-icon :component="DeleteOutlined" /></template>
              批量删除
            </n-button>
            <n-button size="small" text @click="selectedRowKeys = []"> 取消选择 </n-button>
          </n-space>
        </n-space>
      </div>

      <!-- 数据表格 -->
      <n-data-table
        :columns="columns"
        :data="list"
        :loading="loading"
        :pagination="pagination"
        :row-key="(row) => row.id"
        :row-selection="rowSelection"
        @update:page="handlePageChange"
        @update:page-size="handlePageSizeChange"
      />
    </n-card>

    <!-- 编辑弹窗 -->
    <n-modal
      v-model:show="showEditModal"
      :title="editForm.id ? '编辑波动率配置' : '新增波动率配置'"
      preset="card"
      style="width: 1200px; max-width: 95vw"
      :mask-closable="false"
      class="volatility-config-modal"
    >
      <n-form
        ref="formRef"
        :model="editForm"
        :rules="rules"
        label-placement="top"
        label-width="auto"
      >
        <!-- 基本信息 -->
        <n-card title="基本信息" size="small" :bordered="false" style="margin-bottom: 20px">
          <n-grid :cols="1" :x-gap="16">
            <n-gi>
              <n-form-item label="交易对" path="symbol">
                <n-select
                  v-model:value="editForm.symbol"
                  :options="symbolSelectOptions"
                  filterable
                  tag
                  clearable
                  placeholder="选择或输入交易对（留空为全局配置，将应用于所有未配置的交易对）"
                  @search="handleSymbolSearch"
                  style="width: 100%"
                />
                <template #feedback>
                  <n-text depth="3" style="font-size: 12px"
                    >留空表示全局配置，将作为未配置交易对的默认值</n-text
                  >
                </template>
              </n-form-item>
            </n-gi>
          </n-grid>
        </n-card>

        <!-- 市场状态阈值 -->
        <n-card title="市场状态阈值" size="small" :bordered="false" style="margin-bottom: 20px">
          <template #header-extra>
            <n-tooltip trigger="hover">
              <template #trigger>
                <n-icon :component="QuestionCircleOutlined" style="cursor: help; color: #999" />
              </template>
              这些阈值用于新算法判断市场状态：V = (H-L)/delta，D = 方向一致性
            </n-tooltip>
          </template>

          <n-grid :cols="2" :x-gap="20" :y-gap="16">
            <n-gi>
              <n-form-item label="低波动阈值 (LowV)" path="lowVolatilityThreshold">
                <n-input-number
                  v-model:value="editForm.lowVolatilityThreshold"
                  :min="0.1"
                  :max="5"
                  :precision="2"
                  :step="0.1"
                  style="width: 100%"
                  placeholder="建议: 1.0"
                />
                <template #feedback>
                  <n-text depth="3" style="font-size: 12px">V < LowV → 低波动市场</n-text>
                </template>
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="高波动阈值 (HighV)" path="highVolatilityThreshold">
                <n-input-number
                  v-model:value="editForm.highVolatilityThreshold"
                  :min="0.5"
                  :max="10"
                  :precision="2"
                  :step="0.1"
                  style="width: 100%"
                  placeholder="建议: 2.0"
                />
                <template #feedback>
                  <n-text depth="3" style="font-size: 12px"
                    >V ≥ HighV && D < 0.4 → 高波动市场</n-text
                  >
                </template>
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="趋势阈值 (TrendV)" path="trendStrengthThreshold">
                <n-input-number
                  v-model:value="editForm.trendStrengthThreshold"
                  :min="0.5"
                  :max="5"
                  :precision="2"
                  :step="0.1"
                  style="width: 100%"
                  placeholder="建议: 1.2"
                />
                <template #feedback>
                  <n-text depth="3" style="font-size: 12px"
                    >V ≥ TrendV && D ≥ DThreshold → 趋势市场</n-text
                  >
                </template>
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="方向一致性阈值 (DThreshold)" path="dThreshold">
                <n-input-number
                  v-model:value="editForm.dThreshold"
                  :min="0.1"
                  :max="1"
                  :precision="2"
                  :step="0.05"
                  style="width: 100%"
                  placeholder="建议: 0.7"
                />
                <template #feedback>
                  <n-text depth="3" style="font-size: 12px">趋势判断的方向一致性阈值（0-1）</n-text>
                </template>
              </n-form-item>
            </n-gi>
          </n-grid>

          <!-- 阈值验证提示 -->
          <n-alert
            v-if="editForm.lowVolatilityThreshold >= editForm.highVolatilityThreshold"
            type="error"
            style="margin-top: 16px"
          >
            低波动阈值应小于高波动阈值
          </n-alert>
          <n-alert
            v-if="editForm.trendStrengthThreshold > editForm.highVolatilityThreshold"
            type="warning"
            style="margin-top: 12px"
          >
            趋势阈值通常应 ≤ 高波动阈值
          </n-alert>
        </n-card>

        <!-- 各周期Delta值 -->
        <n-card
          title="各周期Delta值（波动点数阈值）"
          size="small"
          :bordered="false"
          style="margin-bottom: 20px"
        >
          <template #header-extra>
            <n-tooltip trigger="hover" style="max-width: 500px">
              <template #trigger>
                <n-icon :component="QuestionCircleOutlined" style="cursor: help; color: #999" />
              </template>
              <div style="line-height: 1.6">
                <div style="margin-bottom: 6px; font-weight: 600">Delta值说明：</div>
                <div style="margin-bottom: 4px"
                  ><strong>含义</strong
                  >：在该周期内，被认为"仍然属于低波动"的价格波动幅度（USDT）</div
                >
                <div style="margin-bottom: 4px"
                  ><strong>公式</strong>：V = (当前K线最高价 - 当前K线最低价) ÷ Delta</div
                >
                <div style="margin-bottom: 4px"
                  ><strong>判断</strong>：V &lt; 1为低波动，V ≈ 1为震荡，V &gt; 1为高波动或趋势</div
                >
                <div style="margin-bottom: 4px"><strong>设置原则</strong>：</div>
                <ul style="margin: 4px 0; padding-left: 18px; font-size: 12px">
                  <li
                    >周期越长，Delta值越大（1m:10-50, 5m:20-100, 15m:50-300, 30m:100-500,
                    1h:200-1000）</li
                  >
                  <li
                    >价格越高，Delta值可以越大（BTCUSDT: 10-1000, ETHUSDT: 5-800, 小币种:
                    0.1-8）</li
                  >
                  <li>根据历史K线平均波动设置，或根据实际效果调整</li>
                </ul>
                <div
                  style="
                    margin-top: 6px;
                    padding: 6px;
                    background: #f0f9ff;
                    border-radius: 4px;
                    font-size: 12px;
                  "
                >
                  <strong>示例（BTCUSDT）</strong>：15分钟Delta=300，K线波动5 USDT，V=5÷300≈0.016
                  &lt; 1 → 低波动
                </div>
              </div>
            </n-tooltip>
          </template>

          <n-grid :cols="3" :x-gap="20" :y-gap="16">
            <n-gi>
              <n-form-item label="1分钟周期" path="delta1m">
                <n-input-number
                  v-model:value="editForm.delta1m"
                  :min="0.1"
                  :max="1000"
                  :precision="2"
                  :step="5"
                  style="width: 100%"
                  placeholder="建议: 10-50 (BTCUSDT)"
                />
                <template #feedback>
                  <n-text depth="3" style="font-size: 12px"
                    >1分钟周期正常波动基准（USDT），参考范围：BTCUSDT 10-50</n-text
                  >
                </template>
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="5分钟周期" path="delta5m">
                <n-input-number
                  v-model:value="editForm.delta5m"
                  :min="0.1"
                  :max="1000"
                  :precision="2"
                  :step="5"
                  style="width: 100%"
                  placeholder="建议: 20-100 (BTCUSDT)"
                />
                <template #feedback>
                  <n-text depth="3" style="font-size: 12px"
                    >5分钟周期正常波动基准（USDT），参考范围：BTCUSDT 20-100</n-text
                  >
                </template>
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="15分钟周期" path="delta15m">
                <n-input-number
                  v-model:value="editForm.delta15m"
                  :min="0.1"
                  :max="1000"
                  :precision="2"
                  :step="10"
                  style="width: 100%"
                  placeholder="建议: 50-300 (BTCUSDT)"
                />
                <template #feedback>
                  <n-text depth="3" style="font-size: 12px"
                    >15分钟周期正常波动基准（USDT），参考范围：BTCUSDT 50-300</n-text
                  >
                </template>
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="30分钟周期" path="delta30m">
                <n-input-number
                  v-model:value="editForm.delta30m"
                  :min="0.1"
                  :max="1000"
                  :precision="2"
                  :step="10"
                  style="width: 100%"
                  placeholder="建议: 100-500 (BTCUSDT)"
                />
                <template #feedback>
                  <n-text depth="3" style="font-size: 12px"
                    >30分钟周期正常波动基准（USDT），参考范围：BTCUSDT 100-500</n-text
                  >
                </template>
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="1小时周期" path="delta1h">
                <n-input-number
                  v-model:value="editForm.delta1h"
                  :min="0.1"
                  :max="2000"
                  :precision="2"
                  :step="20"
                  style="width: 100%"
                  placeholder="建议: 200-1000 (BTCUSDT)"
                />
                <template #feedback>
                  <n-text depth="3" style="font-size: 12px"
                    >1小时周期正常波动基准（USDT），参考范围：BTCUSDT 200-1000</n-text
                  >
                </template>
              </n-form-item>
            </n-gi>
          </n-grid>
        </n-card>

        <!-- 时间周期权重 -->
        <n-card title="时间周期权重" size="small" :bordered="false" style="margin-bottom: 20px">
          <template #header-extra>
            <n-tooltip trigger="hover">
              <template #trigger>
                <n-icon :component="QuestionCircleOutlined" style="cursor: help; color: #999" />
              </template>
              不同时间周期的权重，用于综合判断市场状态。建议合计为1.0
            </n-tooltip>
          </template>

          <n-grid :cols="3" :x-gap="20" :y-gap="20">
            <n-gi>
              <n-form-item label="1分钟周期权重">
                <div class="weight-slider-container">
                  <n-slider
                    v-model:value="editForm.weight1m"
                    :min="0"
                    :max="1"
                    :step="0.01"
                    :tooltip="false"
                    style="width: 100%"
                  />
                  <div class="weight-value">
                    <n-input-number
                      v-model:value="editForm.weight1m"
                      :min="0"
                      :max="1"
                      :precision="2"
                      :step="0.01"
                      size="small"
                      style="width: 100px"
                    />
                    <span class="weight-percent">{{ (editForm.weight1m * 100).toFixed(0) }}%</span>
                  </div>
                </div>
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="5分钟周期权重">
                <div class="weight-slider-container">
                  <n-slider
                    v-model:value="editForm.weight5m"
                    :min="0"
                    :max="1"
                    :step="0.01"
                    :tooltip="false"
                    style="width: 100%"
                  />
                  <div class="weight-value">
                    <n-input-number
                      v-model:value="editForm.weight5m"
                      :min="0"
                      :max="1"
                      :precision="2"
                      :step="0.01"
                      size="small"
                      style="width: 100px"
                    />
                    <span class="weight-percent">{{ (editForm.weight5m * 100).toFixed(0) }}%</span>
                  </div>
                </div>
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="15分钟周期权重">
                <div class="weight-slider-container">
                  <n-slider
                    v-model:value="editForm.weight15m"
                    :min="0"
                    :max="1"
                    :step="0.01"
                    :tooltip="false"
                    style="width: 100%"
                  />
                  <div class="weight-value">
                    <n-input-number
                      v-model:value="editForm.weight15m"
                      :min="0"
                      :max="1"
                      :precision="2"
                      :step="0.01"
                      size="small"
                      style="width: 100px"
                    />
                    <span class="weight-percent">{{ (editForm.weight15m * 100).toFixed(0) }}%</span>
                  </div>
                </div>
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="30分钟周期权重">
                <div class="weight-slider-container">
                  <n-slider
                    v-model:value="editForm.weight30m"
                    :min="0"
                    :max="1"
                    :step="0.01"
                    :tooltip="false"
                    style="width: 100%"
                  />
                  <div class="weight-value">
                    <n-input-number
                      v-model:value="editForm.weight30m"
                      :min="0"
                      :max="1"
                      :precision="2"
                      :step="0.01"
                      size="small"
                      style="width: 100px"
                    />
                    <span class="weight-percent">{{ (editForm.weight30m * 100).toFixed(0) }}%</span>
                  </div>
                </div>
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="1小时周期权重">
                <div class="weight-slider-container">
                  <n-slider
                    v-model:value="editForm.weight1h"
                    :min="0"
                    :max="1"
                    :step="0.01"
                    :tooltip="false"
                    style="width: 100%"
                  />
                  <div class="weight-value">
                    <n-input-number
                      v-model:value="editForm.weight1h"
                      :min="0"
                      :max="1"
                      :precision="2"
                      :step="0.01"
                      size="small"
                      style="width: 100px"
                    />
                    <span class="weight-percent">{{ (editForm.weight1h * 100).toFixed(0) }}%</span>
                  </div>
                </div>
              </n-form-item>
            </n-gi>
          </n-grid>

          <div
            style="
              margin-top: 16px;
              padding: 14px;
              background: linear-gradient(135deg, #667eea15 0%, #764ba215 100%);
              border-radius: 8px;
              border: 1px solid #e0e0e0;
            "
          >
            <n-space justify="space-between" align="center">
              <n-text
                :type="weightSum === 1 ? 'success' : weightSum > 1 ? 'error' : 'warning'"
                strong
                style="font-size: 14px"
              >
                权重合计: {{ weightSum.toFixed(2) }}
                <span v-if="weightSum === 1" style="color: #18a058">✓ 完美</span>
                <span v-else-if="weightSum > 1" style="color: #d03050">⚠ 超出1.0</span>
                <span v-else style="color: #f0a020">⚠ 建议调整为1.0</span>
              </n-text>
              <n-button
                size="small"
                type="primary"
                ghost
                @click="autoBalanceWeights"
                :disabled="weightSum === 1"
              >
                自动平衡
              </n-button>
            </n-space>
          </div>
        </n-card>

        <!-- 其他设置和预览 -->
        <n-grid :cols="2" :x-gap="20">
          <n-gi>
            <n-card title="其他设置" size="small" :bordered="false">
              <n-form-item label="启用状态" label-placement="top">
                <n-switch
                  v-model:value="editForm.isActive"
                  :checked-value="1"
                  :unchecked-value="0"
                />
                <template #feedback>
                  <n-text depth="3" style="font-size: 12px">禁用后，该配置将不会被使用</n-text>
                </template>
              </n-form-item>
            </n-card>
          </n-gi>
          <n-gi>
            <n-card title="配置预览" size="small" :bordered="false">
              <n-space vertical size="small">
                <div>
                  <n-text depth="3" style="font-size: 12px; margin-right: 8px">阈值:</n-text>
                  <n-tag size="small" type="info"
                    >LowV:{{ editForm.lowVolatilityThreshold.toFixed(2) }}</n-tag
                  >
                  <n-tag size="small" type="warning"
                    >HighV:{{ editForm.highVolatilityThreshold.toFixed(2) }}</n-tag
                  >
                  <n-tag size="small" type="success"
                    >TrendV:{{ editForm.trendStrengthThreshold.toFixed(2) }}</n-tag
                  >
                  <n-tag size="small" type="default">D:{{ editForm.dThreshold.toFixed(2) }}</n-tag>
                </div>
                <div>
                  <n-text depth="3" style="font-size: 12px; margin-right: 8px">Delta:</n-text>
                  <n-text style="font-size: 12px">
                    1m:{{ editForm.delta1m.toFixed(1) }} 5m:{{ editForm.delta5m.toFixed(1) }} 15m:{{
                      editForm.delta15m.toFixed(1)
                    }}
                    30m:{{ editForm.delta30m.toFixed(1) }} 1h:{{ editForm.delta1h.toFixed(1) }}
                  </n-text>
                </div>
                <div>
                  <n-text depth="3" style="font-size: 12px; margin-right: 8px">权重:</n-text>
                  <n-text
                    :type="weightSum === 1 ? 'success' : 'warning'"
                    strong
                    style="font-size: 13px"
                  >
                    {{ weightSum.toFixed(2) }} {{ weightSum === 1 ? '✓' : '⚠' }}
                  </n-text>
                </div>
              </n-space>
            </n-card>
          </n-gi>
        </n-grid>
      </n-form>

      <template #action>
        <n-space justify="end">
          <n-button @click="showEditModal = false">取消</n-button>
          <n-button @click="copyFromGlobal" v-if="editForm.id === 0 && !editForm.symbol">
            从全局配置复制
          </n-button>
          <n-button type="primary" @click="handleSave" :loading="saving">保存</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 批量编辑弹窗 -->
    <n-modal
      v-model:show="showBatchEditModal"
      title="批量设置波动率配置"
      preset="card"
      style="width: 900px"
      :mask-closable="false"
      class="volatility-config-modal"
    >
      <n-alert type="warning" style="margin-bottom: 16px">
        将为选中的交易对批量设置相同的配置参数。如果交易对已存在配置，将被更新；不存在则创建新配置。
      </n-alert>

      <n-form
        ref="batchFormRef"
        :model="batchEditForm"
        :rules="batchRules"
        label-placement="left"
        label-width="140"
      >
        <n-card title="基本信息" size="small" :bordered="false" style="margin-bottom: 16px">
          <n-form-item label="选择交易对" path="symbols">
            <n-select
              v-model:value="batchEditForm.symbols"
              :options="symbolSelectOptions"
              multiple
              filterable
              tag
              placeholder="选择或输入多个交易对（支持手动输入）"
              @search="handleSymbolSearch"
              style="width: 100%"
            />
            <template #feedback>
              <n-text depth="3" style="font-size: 12px">
                已选择 {{ batchEditForm.symbols.length }} 个交易对
              </n-text>
            </template>
          </n-form-item>
        </n-card>

        <n-card title="市场状态阈值" size="small" :bordered="false" style="margin-bottom: 16px">
          <template #header-extra>
            <n-tooltip trigger="hover">
              <template #trigger>
                <n-icon :component="QuestionCircleOutlined" style="cursor: help; color: #999" />
              </template>
              这些阈值用于新算法判断市场状态：V = (H-L)/delta，D = 方向一致性
            </n-tooltip>
          </template>

          <n-grid :cols="4" :x-gap="16">
            <n-gi>
              <n-form-item label="低波动阈值 LowV">
                <n-input-number
                  v-model:value="batchEditForm.lowVolatilityThreshold"
                  :min="0.1"
                  :max="5"
                  :precision="2"
                  :step="0.1"
                  style="width: 100%"
                  placeholder="建议: 1.0"
                />
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="高波动阈值 HighV">
                <n-input-number
                  v-model:value="batchEditForm.highVolatilityThreshold"
                  :min="0.5"
                  :max="10"
                  :precision="2"
                  :step="0.1"
                  style="width: 100%"
                  placeholder="建议: 2.0"
                />
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="趋势阈值 TrendV">
                <n-input-number
                  v-model:value="batchEditForm.trendStrengthThreshold"
                  :min="0.5"
                  :max="5"
                  :precision="2"
                  :step="0.1"
                  style="width: 100%"
                  placeholder="建议: 1.2"
                />
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="方向一致性 DThreshold">
                <n-input-number
                  v-model:value="batchEditForm.dThreshold"
                  :min="0.1"
                  :max="1"
                  :precision="2"
                  :step="0.05"
                  style="width: 100%"
                  placeholder="建议: 0.7"
                />
              </n-form-item>
            </n-gi>
          </n-grid>

          <!-- 阈值验证提示 -->
          <n-alert
            v-if="batchEditForm.lowVolatilityThreshold >= batchEditForm.highVolatilityThreshold"
            type="error"
            style="margin-top: 12px"
          >
            低波动阈值应小于高波动阈值
          </n-alert>
          <n-alert
            v-if="batchEditForm.trendStrengthThreshold > batchEditForm.highVolatilityThreshold"
            type="warning"
            style="margin-top: 12px"
          >
            趋势阈值通常应 ≤ 高波动阈值
          </n-alert>
        </n-card>

        <n-card title="各周期Delta值" size="small" :bordered="false" style="margin-bottom: 16px">
          <template #header-extra>
            <n-tooltip trigger="hover" style="max-width: 500px">
              <template #trigger>
                <n-icon :component="QuestionCircleOutlined" style="cursor: help; color: #999" />
              </template>
              <div style="line-height: 1.6">
                <div style="margin-bottom: 6px; font-weight: 600">Delta值说明：</div>
                <div style="margin-bottom: 4px"
                  ><strong>含义</strong
                  >：在该周期内，被认为"仍然属于低波动"的价格波动幅度（USDT）</div
                >
                <div style="margin-bottom: 4px"
                  ><strong>公式</strong>：V = (当前K线最高价 - 当前K线最低价) ÷ Delta</div
                >
                <div style="margin-bottom: 4px"
                  ><strong>判断</strong>：V &lt; 1为低波动，V ≈ 1为震荡，V &gt; 1为高波动或趋势</div
                >
                <div style="margin-bottom: 4px"><strong>设置原则</strong>：</div>
                <ul style="margin: 4px 0; padding-left: 18px; font-size: 12px">
                  <li
                    >周期越长，Delta值越大（1m:10-50, 5m:20-100, 15m:50-300, 30m:100-500,
                    1h:200-1000）</li
                  >
                  <li
                    >价格越高，Delta值可以越大（BTCUSDT: 10-1000, ETHUSDT: 5-800, 小币种:
                    0.1-8）</li
                  >
                  <li>根据历史K线平均波动设置，或根据实际效果调整</li>
                </ul>
                <div
                  style="
                    margin-top: 6px;
                    padding: 6px;
                    background: #f0f9ff;
                    border-radius: 4px;
                    font-size: 12px;
                  "
                >
                  <strong>示例（BTCUSDT）</strong>：15分钟Delta=300，K线波动5 USDT，V=5÷300≈0.016
                  &lt; 1 → 低波动
                </div>
              </div>
            </n-tooltip>
          </template>

          <n-grid :cols="5" :x-gap="16">
            <n-gi>
              <n-form-item label="1分钟 Delta">
                <n-input-number
                  v-model:value="batchEditForm.delta1m"
                  :min="0.1"
                  :max="1000"
                  :precision="2"
                  :step="5"
                  style="width: 100%"
                  placeholder="建议: 10-50 (BTCUSDT)"
                />
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="5分钟 Delta">
                <n-input-number
                  v-model:value="batchEditForm.delta5m"
                  :min="0.1"
                  :max="1000"
                  :precision="2"
                  :step="5"
                  style="width: 100%"
                  placeholder="建议: 20-100 (BTCUSDT)"
                />
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="15分钟 Delta">
                <n-input-number
                  v-model:value="batchEditForm.delta15m"
                  :min="0.1"
                  :max="1000"
                  :precision="2"
                  :step="10"
                  style="width: 100%"
                  placeholder="建议: 50-300 (BTCUSDT)"
                />
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="30分钟 Delta">
                <n-input-number
                  v-model:value="batchEditForm.delta30m"
                  :min="0.1"
                  :max="1000"
                  :precision="2"
                  :step="10"
                  style="width: 100%"
                  placeholder="建议: 100-500 (BTCUSDT)"
                />
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="1小时 Delta">
                <n-input-number
                  v-model:value="batchEditForm.delta1h"
                  :min="0.1"
                  :max="2000"
                  :precision="2"
                  :step="20"
                  style="width: 100%"
                  placeholder="建议: 200-1000 (BTCUSDT)"
                />
              </n-form-item>
            </n-gi>
          </n-grid>

          <n-divider title-placement="left">
            <n-space align="center" :size="8">
              <span>时间周期权重</span>
              <n-tooltip trigger="hover">
                <template #trigger>
                  <n-icon :component="QuestionCircleOutlined" style="cursor: help; color: #999" />
                </template>
                不同时间周期的权重，用于综合判断市场状态。建议合计为1.0
              </n-tooltip>
            </n-space>
          </n-divider>

          <n-grid :cols="5" :x-gap="16">
            <n-gi>
              <n-form-item label="1分钟">
                <div class="weight-slider-container">
                  <n-slider
                    v-model:value="batchEditForm.weight1m"
                    :min="0"
                    :max="1"
                    :step="0.01"
                    :tooltip="false"
                    style="width: 100%"
                  />
                  <div class="weight-value">
                    <n-input-number
                      v-model:value="batchEditForm.weight1m"
                      :min="0"
                      :max="1"
                      :precision="2"
                      :step="0.01"
                      size="small"
                      style="width: 80px"
                    />
                    <span class="weight-percent"
                      >{{ (batchEditForm.weight1m * 100).toFixed(0) }}%</span
                    >
                  </div>
                </div>
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="5分钟">
                <div class="weight-slider-container">
                  <n-slider
                    v-model:value="batchEditForm.weight5m"
                    :min="0"
                    :max="1"
                    :step="0.01"
                    :tooltip="false"
                    style="width: 100%"
                  />
                  <div class="weight-value">
                    <n-input-number
                      v-model:value="batchEditForm.weight5m"
                      :min="0"
                      :max="1"
                      :precision="2"
                      :step="0.01"
                      size="small"
                      style="width: 80px"
                    />
                    <span class="weight-percent"
                      >{{ (batchEditForm.weight5m * 100).toFixed(0) }}%</span
                    >
                  </div>
                </div>
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="15分钟">
                <div class="weight-slider-container">
                  <n-slider
                    v-model:value="batchEditForm.weight15m"
                    :min="0"
                    :max="1"
                    :step="0.01"
                    :tooltip="false"
                    style="width: 100%"
                  />
                  <div class="weight-value">
                    <n-input-number
                      v-model:value="batchEditForm.weight15m"
                      :min="0"
                      :max="1"
                      :precision="2"
                      :step="0.01"
                      size="small"
                      style="width: 80px"
                    />
                    <span class="weight-percent"
                      >{{ (batchEditForm.weight15m * 100).toFixed(0) }}%</span
                    >
                  </div>
                </div>
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="30分钟">
                <div class="weight-slider-container">
                  <n-slider
                    v-model:value="batchEditForm.weight30m"
                    :min="0"
                    :max="1"
                    :step="0.01"
                    :tooltip="false"
                    style="width: 100%"
                  />
                  <div class="weight-value">
                    <n-input-number
                      v-model:value="batchEditForm.weight30m"
                      :min="0"
                      :max="1"
                      :precision="2"
                      :step="0.01"
                      size="small"
                      style="width: 80px"
                    />
                    <span class="weight-percent"
                      >{{ (batchEditForm.weight30m * 100).toFixed(0) }}%</span
                    >
                  </div>
                </div>
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="1小时">
                <div class="weight-slider-container">
                  <n-slider
                    v-model:value="batchEditForm.weight1h"
                    :min="0"
                    :max="1"
                    :step="0.01"
                    :tooltip="false"
                    style="width: 100%"
                  />
                  <div class="weight-value">
                    <n-input-number
                      v-model:value="batchEditForm.weight1h"
                      :min="0"
                      :max="1"
                      :precision="2"
                      :step="0.01"
                      size="small"
                      style="width: 80px"
                    />
                    <span class="weight-percent"
                      >{{ (batchEditForm.weight1h * 100).toFixed(0) }}%</span
                    >
                  </div>
                </div>
              </n-form-item>
            </n-gi>
          </n-grid>

          <div
            style="
              margin-top: 12px;
              padding: 12px;
              background: linear-gradient(135deg, #667eea15 0%, #764ba215 100%);
              border-radius: 6px;
              border: 1px solid #e0e0e0;
            "
          >
            <n-space justify="space-between" align="center">
              <n-text
                :type="batchWeightSum === 1 ? 'success' : batchWeightSum > 1 ? 'error' : 'warning'"
                strong
                style="font-size: 14px"
              >
                权重合计: {{ batchWeightSum.toFixed(2) }}
                <span v-if="batchWeightSum === 1" style="color: #18a058">✓ 完美</span>
                <span v-else-if="batchWeightSum > 1" style="color: #d03050">⚠ 超出1.0</span>
                <span v-else style="color: #f0a020">⚠ 建议调整为1.0</span>
              </n-text>
              <n-button
                size="small"
                type="primary"
                ghost
                @click="autoBalanceBatchWeights"
                :disabled="batchWeightSum === 1"
              >
                自动平衡
              </n-button>
            </n-space>
          </div>
        </n-card>

        <n-card title="其他设置" size="small" :bordered="false">
          <n-form-item label="启用状态">
            <n-switch
              v-model:value="batchEditForm.isActive"
              :checked-value="1"
              :unchecked-value="0"
            />
          </n-form-item>
        </n-card>
      </n-form>

      <template #action>
        <n-space justify="end">
          <n-button @click="showBatchEditModal = false">取消</n-button>
          <n-button type="primary" @click="handleBatchSave" :loading="saving">批量保存</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive, onMounted, h, computed } from 'vue';
  import { useMessage, useDialog, NButton, NTag, NSpace, NIcon } from 'naive-ui';
  import {
    PlusOutlined,
    EditOutlined,
    DeleteOutlined,
    SearchOutlined,
    ReloadOutlined,
    CopyOutlined,
    InfoCircleOutlined,
    QuestionCircleOutlined,
    CheckCircleOutlined,
    CloseCircleOutlined,
  } from '@vicons/antd';
  import { ToogoVolatilityConfigApi } from '@/api/toogo';

  const message = useMessage();
  const dialog = useDialog();

  // 默认50个主流币种
  const defaultSymbols = [
    'BTCUSDT',
    'ETHUSDT',
    'BNBUSDT',
    'XRPUSDT',
    'SOLUSDT',
    'DOGEUSDT',
    'ADAUSDT',
    'TRXUSDT',
    'AVAXUSDT',
    'LINKUSDT',
    'DOTUSDT',
    'MATICUSDT',
    'LTCUSDT',
    'SHIBUSDT',
    'BCHUSDT',
    'ATOMUSDT',
    'UNIUSDT',
    'XLMUSDT',
    'ETCUSDT',
    'NEARUSDT',
    'APTUSDT',
    'FILUSDT',
    'ICPUSDT',
    'VETUSDT',
    'HBARUSDT',
    'SANDUSDT',
    'MANAUSDT',
    'AAVEUSDT',
    'AXSUSDT',
    'EOSUSDT',
    'THETAUSDT',
    'FTMUSDT',
    'ALGOUSDT',
    'XTZUSDT',
    'FLOWUSDT',
    'GRTUSDT',
    'CHZUSDT',
    'APEUSDT',
    'LDOUSDT',
    'ARBUSDT',
    'OPUSDT',
    'INJUSDT',
    'RUNEUSDT',
    'MKRUSDT',
    'SNXUSDT',
    'COMPUSDT',
    'CRVUSDT',
    '1INCHUSDT',
    'ENSUSDT',
    'GMXUSDT',
  ];

  // 数据
  const loading = ref(false);
  const saving = ref(false);
  const list = ref<any[]>([]);
  const showEditModal = ref(false);
  const showBatchEditModal = ref(false);
  const formRef = ref<any>(null);
  const batchFormRef = ref<any>(null);
  const selectedRowKeys = ref<number[]>([]);

  // 交易对选项
  const symbolSelectOptions = ref(defaultSymbols.map((s) => ({ label: s, value: s })));

  // 搜索参数
  const searchParams = reactive({
    symbol: '',
    isActive: undefined as number | undefined,
  });

  // 分页
  const pagination = reactive({
    page: 1,
    pageSize: 20,
    total: 0,
    showSizePicker: true,
    pageSizes: [10, 20, 50, 100],
  });

  // 启用状态选项
  const activeOptions = [
    { label: '启用', value: 1 },
    { label: '禁用', value: 0 },
  ];

  // 编辑表单（适配新算法）
  const editForm = reactive({
    id: 0,
    symbol: null as string | null,
    highVolatilityThreshold: 2.0, // HighV
    lowVolatilityThreshold: 1.0, // LowV
    trendStrengthThreshold: 1.2, // TrendV
    dThreshold: 0.7, // DThreshold
    delta1m: 2.0,
    delta5m: 2.0,
    delta15m: 3.0,
    delta30m: 3.0,
    delta1h: 5.0,
    weight1m: 0.2,
    weight5m: 0.25,
    weight15m: 0.25,
    weight30m: 0.2,
    weight1h: 0.1,
    isActive: 1,
  });

  // 批量编辑表单（适配新算法）
  const batchEditForm = reactive({
    symbols: [] as string[],
    highVolatilityThreshold: 2.0, // HighV
    lowVolatilityThreshold: 1.0, // LowV
    trendStrengthThreshold: 1.2, // TrendV
    dThreshold: 0.7, // DThreshold
    delta1m: 2.0,
    delta5m: 2.0,
    delta15m: 3.0,
    delta30m: 3.0,
    delta1h: 5.0,
    weight1m: 0.2,
    weight5m: 0.25,
    weight15m: 0.25,
    weight30m: 0.2,
    weight1h: 0.1,
    isActive: 1,
  });

  // 计算权重总和
  const weightSum = computed(() => {
    return (
      Math.round(
        (editForm.weight1m +
          editForm.weight5m +
          editForm.weight15m +
          editForm.weight30m +
          editForm.weight1h) *
          100,
      ) / 100
    );
  });

  const batchWeightSum = computed(() => {
    return (
      Math.round(
        (batchEditForm.weight1m +
          batchEditForm.weight5m +
          batchEditForm.weight15m +
          batchEditForm.weight30m +
          batchEditForm.weight1h) *
          100,
      ) / 100
    );
  });

  // 表单验证规则（适配新算法）
  const rules = {
    highVolatilityThreshold: [
      { required: true, type: 'number', message: '请输入高波动阈值HighV', trigger: 'blur' },
    ],
    lowVolatilityThreshold: [
      { required: true, type: 'number', message: '请输入低波动阈值LowV', trigger: 'blur' },
    ],
    trendStrengthThreshold: [
      { required: true, type: 'number', message: '请输入趋势阈值TrendV', trigger: 'blur' },
    ],
    dThreshold: [
      {
        required: true,
        type: 'number',
        message: '请输入方向一致性阈值DThreshold',
        trigger: 'blur',
      },
    ],
    delta1m: [{ required: true, type: 'number', message: '请输入1分钟delta', trigger: 'blur' }],
    delta5m: [{ required: true, type: 'number', message: '请输入5分钟delta', trigger: 'blur' }],
    delta15m: [{ required: true, type: 'number', message: '请输入15分钟delta', trigger: 'blur' }],
    delta30m: [{ required: true, type: 'number', message: '请输入30分钟delta', trigger: 'blur' }],
    delta1h: [{ required: true, type: 'number', message: '请输入1小时delta', trigger: 'blur' }],
  };

  const batchRules = {
    symbols: [
      { required: true, type: 'array', message: '请选择至少一个交易对', trigger: 'change' },
    ],
  };

  // 表格行选择配置
  const rowSelection = {
    type: 'checkbox',
    showCheckedAll: true,
    onSelect: (row: any, selected: boolean) => {
      if (selected) {
        if (!selectedRowKeys.value.includes(row.id)) {
          selectedRowKeys.value.push(row.id);
        }
      } else {
        const index = selectedRowKeys.value.indexOf(row.id);
        if (index > -1) {
          selectedRowKeys.value.splice(index, 1);
        }
      }
    },
    onSelectAll: (selected: boolean) => {
      if (selected) {
        selectedRowKeys.value = list.value.map((row) => row.id);
      } else {
        selectedRowKeys.value = [];
      }
    },
    getCheckboxProps: (row: any) => ({
      disabled: row.symbol === null, // 全局配置不能批量操作
    }),
  };

  // 表格列（优化：使用中文标题，删除无用列）
  const columns = [
    {
      title: '交易对',
      key: 'symbol',
      width: 140,
      fixed: 'left',
      render: (row: any) =>
        row.symbol
          ? h(
              NTag,
              { type: 'primary', size: 'small', bordered: false },
              { default: () => row.symbol },
            )
          : h(
              NTag,
              { type: 'default', size: 'small', bordered: false },
              { default: () => '全局配置' },
            ),
    },
    {
      title: '低波动阈值',
      key: 'lowVolatilityThreshold',
      width: 110,
      render: (row: any) =>
        h(
          'span',
          { style: { fontWeight: 500, color: '#2080f0' } },
          row.lowVolatilityThreshold?.toFixed(2) || '1.00',
        ),
    },
    {
      title: '高波动阈值',
      key: 'highVolatilityThreshold',
      width: 110,
      render: (row: any) =>
        h(
          'span',
          { style: { fontWeight: 500, color: '#f0a020' } },
          `${row.highVolatilityThreshold?.toFixed(2) || '2.00'}`,
        ),
    },
    {
      title: '趋势阈值',
      key: 'trendStrengthThreshold',
      width: 110,
      render: (row: any) =>
        h(
          'span',
          { style: { fontWeight: 500, color: '#18a058' } },
          row.trendStrengthThreshold?.toFixed(2) || '1.20',
        ),
    },
    {
      title: '方向一致性阈值',
      key: 'dThreshold',
      width: 130,
      render: (row: any) =>
        h(
          'span',
          { style: { fontWeight: 500, color: '#722ed1' } },
          row.dThreshold?.toFixed(2) || '0.70',
        ),
    },
    {
      title: 'Delta值',
      key: 'deltas',
      width: 220,
      render: (row: any) => {
        const deltas = [
          { label: '1分钟', value: row.delta1m ?? 2.0 },
          { label: '5分钟', value: row.delta5m ?? 2.0 },
          { label: '15分钟', value: row.delta15m ?? 3.0 },
          { label: '30分钟', value: row.delta30m ?? 3.0 },
          { label: '1小时', value: row.delta1h ?? 5.0 },
        ];
        return h(
          'div',
          { style: { fontSize: '12px', lineHeight: '1.8' } },
          deltas.map((d, idx) =>
            h(
              'span',
              {
                key: idx,
                style: {
                  display: 'inline-block',
                  marginRight: '8px',
                  marginBottom: '4px',
                  padding: '2px 8px',
                  background: '#f5f5f5',
                  color: '#666',
                  borderRadius: '3px',
                  border: '1px solid #e0e0e0',
                },
              },
              `${d.label}: ${d.value.toFixed(1)}`,
            ),
          ),
        );
      },
    },
    {
      title: '周期权重',
      key: 'weights',
      width: 300,
      render: (row: any) => {
        const weights = [
          { label: '1分钟', value: row.weight1m ?? 0.1 },
          { label: '5分钟', value: row.weight5m ?? 0.2 },
          { label: '15分钟', value: row.weight15m ?? 0.25 },
          { label: '30分钟', value: row.weight30m ?? 0.25 },
          { label: '1小时', value: row.weight1h ?? 0.2 },
        ];
        const sum = weights.reduce((acc, w) => acc + w.value, 0);
        const sumFixed = Math.round(sum * 100) / 100;
        return h('div', { style: { fontSize: '12px', lineHeight: '1.8' } }, [
          h(
            'div',
            { style: { marginBottom: '6px' } },
            weights.map((w, idx) =>
              h(
                'span',
                {
                  key: idx,
                  style: {
                    display: 'inline-block',
                    marginRight: '8px',
                    marginBottom: '4px',
                    padding: '2px 8px',
                    background: '#f5f5f5',
                    color: '#666',
                    borderRadius: '3px',
                    border: '1px solid #e0e0e0',
                  },
                },
                `${w.label}: ${(w.value * 100).toFixed(0)}%`,
              ),
            ),
          ),
          h(
            'div',
            {
              style: {
                fontSize: '12px',
                color: sumFixed === 1 ? '#18a058' : '#f0a020',
                fontWeight: 500,
                padding: '2px 8px',
                background: sumFixed === 1 ? '#f0f9f4' : '#fff9e6',
                borderRadius: '3px',
                display: 'inline-block',
                border: `1px solid ${sumFixed === 1 ? '#b3e5d1' : '#ffe0b2'}`,
              },
            },
            `合计: ${sumFixed.toFixed(2)}`,
          ),
        ]);
      },
    },
    {
      title: '状态',
      key: 'isActive',
      width: 90,
      render: (row: any) =>
        h(
          NTag,
          {
            type: row.isActive === 1 ? 'success' : 'default',
            size: 'small',
            bordered: false,
            style: { fontWeight: 500 },
          },
          { default: () => (row.isActive === 1 ? '启用' : '禁用') },
        ),
    },
    {
      title: '操作',
      key: 'actions',
      width: 180,
      fixed: 'right',
      render: (row: any) =>
        h(
          NSpace,
          { size: 'small' },
          {
            default: () => [
              h(
                NButton,
                {
                  size: 'small',
                  type: 'primary',
                  onClick: () => openEditModal(row),
                  style: { minWidth: '60px' },
                },
                { default: () => '编辑', icon: () => h(NIcon, { component: EditOutlined }) },
              ),
              h(
                NButton,
                {
                  size: 'small',
                  onClick: () => copyConfig(row),
                  style: { minWidth: '60px' },
                },
                { default: () => '复制', icon: () => h(NIcon, { component: CopyOutlined }) },
              ),
              h(
                NButton,
                {
                  size: 'small',
                  type: 'error',
                  onClick: () => handleDelete(row),
                  style: { minWidth: '60px' },
                },
                { default: () => '删除', icon: () => h(NIcon, { component: DeleteOutlined }) },
              ),
            ],
          },
        ),
    },
  ];

  // 加载数据
  const loadData = async () => {
    loading.value = true;
    try {
      const params: any = {
        page: pagination.page || 1,
        pageSize: pagination.pageSize || 20,
      };
      if (searchParams.symbol) params.symbol = searchParams.symbol;
      if (searchParams.isActive !== undefined && searchParams.isActive !== null) {
        params.isActive = searchParams.isActive;
      }

      const res = await ToogoVolatilityConfigApi.list(params);
      console.log('列表API响应:', res);

      // 处理多种可能的响应格式
      if (res) {
        // 格式1: { code: 0, data: { list: [], total: 0 } }
        if (res.code === 0 && res.data) {
          list.value = res.data.list || [];
          pagination.total = res.data.total || 0;
          console.log('列表数据加载成功:', list.value.length, '条，总数:', pagination.total);
        }
        // 格式2: { list: [], total: 0, page: 1 } - 直接返回数据
        else if (res.list !== undefined) {
          list.value = res.list || [];
          pagination.total = res.total || 0;
          console.log(
            '列表数据加载成功（格式2）:',
            list.value.length,
            '条，总数:',
            pagination.total,
          );
        }
        // 格式3: 错误响应
        else if (res.code !== 0) {
          message.error(res.msg || '获取列表失败');
          console.error('获取列表失败:', res);
        }
      }
    } catch (error: any) {
      console.error('获取列表异常:', error);
      message.error(error.message || '获取列表失败');
    } finally {
      loading.value = false;
    }
  };

  // 交易对搜索（支持手动输入）
  const handleSymbolSearch = (query: string) => {
    if (query) {
      const upperQuery = query.toUpperCase();
      if (!symbolSelectOptions.value.find((opt) => opt.value === upperQuery)) {
        symbolSelectOptions.value = [
          { label: upperQuery, value: upperQuery },
          ...defaultSymbols
            .filter((s) => s.includes(upperQuery))
            .map((s) => ({ label: s, value: s })),
        ];
      }
    } else {
      symbolSelectOptions.value = defaultSymbols.map((s) => ({ label: s, value: s }));
    }
  };

  // 重置搜索
  const resetSearch = () => {
    searchParams.symbol = '';
    searchParams.isActive = undefined;
    pagination.page = 1;
    loadData();
  };

  // 打开编辑弹窗
  const openEditModal = (row?: any) => {
    if (row) {
      Object.assign(editForm, {
        id: row.id,
        symbol: row.symbol || null,
        highVolatilityThreshold: row.highVolatilityThreshold || 2.0,
        lowVolatilityThreshold: row.lowVolatilityThreshold || 1.0,
        trendStrengthThreshold: row.trendStrengthThreshold || 1.2,
        dThreshold: row.dThreshold ?? 0.7,
        delta1m: row.delta1m ?? 2.0,
        delta5m: row.delta5m ?? 2.0,
        delta15m: row.delta15m ?? 3.0,
        delta30m: row.delta30m ?? 3.0,
        delta1h: row.delta1h ?? 5.0,
        weight1m: row.weight1m ?? 0.2,
        weight5m: row.weight5m ?? 0.25,
        weight15m: row.weight15m ?? 0.25,
        weight30m: row.weight30m ?? 0.2,
        weight1h: row.weight1h ?? 0.1,
        isActive: row.isActive ?? 1,
      });
    } else {
      Object.assign(editForm, {
        id: 0,
        symbol: null,
        highVolatilityThreshold: 2.0,
        lowVolatilityThreshold: 1.0,
        trendStrengthThreshold: 1.2,
        dThreshold: 0.7,
        delta1m: 2.0,
        delta5m: 2.0,
        delta15m: 3.0,
        delta30m: 3.0,
        delta1h: 5.0,
        weight1m: 0.2,
        weight5m: 0.25,
        weight15m: 0.25,
        weight30m: 0.2,
        weight1h: 0.1,
        isActive: 1,
      });
    }
    showEditModal.value = true;
  };

  // 打开批量编辑弹窗
  const openBatchEditModal = () => {
    Object.assign(batchEditForm, {
      symbols: [],
      highVolatilityThreshold: 2.0,
      lowVolatilityThreshold: 1.0,
      trendStrengthThreshold: 1.2,
      dThreshold: 0.7,
      delta1m: 2.0,
      delta5m: 2.0,
      delta15m: 3.0,
      delta30m: 3.0,
      delta1h: 5.0,
      weight1m: 0.2,
      weight5m: 0.25,
      weight15m: 0.25,
      weight30m: 0.2,
      weight1h: 0.1,
      isActive: 1,
    });
    showBatchEditModal.value = true;
  };

  // 保存
  const handleSave = async () => {
    if (!formRef.value) return;
    await formRef.value.validate(async (errors: any) => {
      if (errors) return;

      // 验证阈值
      if (editForm.lowVolatilityThreshold >= editForm.highVolatilityThreshold) {
        message.error('低波动阈值应小于高波动阈值');
        return;
      }

      saving.value = true;
      try {
        // 处理symbol：空字符串转为null，非空字符串转为大写
        let symbolValue: string | null = null;
        if (editForm.symbol && editForm.symbol.trim() !== '') {
          symbolValue = editForm.symbol.trim().toUpperCase();
        }

        const data: any = {
          symbol: symbolValue,
          highVolatilityThreshold: editForm.highVolatilityThreshold,
          lowVolatilityThreshold: editForm.lowVolatilityThreshold,
          trendStrengthThreshold: editForm.trendStrengthThreshold,
          dThreshold: editForm.dThreshold,
          delta1m: editForm.delta1m,
          delta5m: editForm.delta5m,
          delta15m: editForm.delta15m,
          delta30m: editForm.delta30m,
          delta1h: editForm.delta1h,
          weight1m: editForm.weight1m,
          weight5m: editForm.weight5m,
          weight15m: editForm.weight15m,
          weight30m: editForm.weight30m,
          weight1h: editForm.weight1h,
          isActive: editForm.isActive,
        };

        let res;
        if (editForm.id > 0) {
          data.id = editForm.id;
          res = await ToogoVolatilityConfigApi.update(data);
        } else {
          res = await ToogoVolatilityConfigApi.create(data);
        }

        if (res.code === 0) {
          message.success(editForm.id > 0 ? '更新成功' : '创建成功');
          showEditModal.value = false;
          loadData();
        } else {
          // 检查是否是配置已存在的错误
          const errorMsg = res.msg || '保存失败';
          if (errorMsg.includes('已存在')) {
            // 尝试从错误信息中提取ID
            const idMatch = errorMsg.match(/ID:\s*(\d+)/);
            if (idMatch) {
              const existingId = parseInt(idMatch[1]);
              // 询问用户是否要编辑现有配置
              dialog.warning({
                title: '配置已存在',
                content: h('div', [
                  h(
                    'p',
                    { style: { marginBottom: '8px' } },
                    `交易对 ${symbolValue || '全局'} 的配置已存在（ID: ${existingId}）。`,
                  ),
                  h('p', { style: { color: '#666', fontSize: '13px' } }, '您可以选择：'),
                  h(
                    'ul',
                    {
                      style: {
                        margin: '8px 0',
                        paddingLeft: '20px',
                        color: '#666',
                        fontSize: '13px',
                      },
                    },
                    [h('li', '编辑现有配置（推荐）'), h('li', '删除现有配置后重新创建')],
                  ),
                ]),
                positiveText: '编辑现有配置',
                negativeText: '取消',
                onPositiveClick: async () => {
                  // 重新加载数据并打开编辑弹窗
                  await loadData();
                  const existingRow = list.value.find((item) => item.id === existingId);
                  if (existingRow) {
                    openEditModal(existingRow);
                    message.info('已打开现有配置的编辑窗口');
                  } else {
                    message.warning('未找到现有配置，请刷新页面后重试');
                  }
                },
              });
            } else {
              message.error(errorMsg);
            }
          } else {
            message.error(errorMsg);
          }
          console.error('保存失败:', res);
        }
      } catch (error: any) {
        console.error('保存异常:', error);

        // 检查是否是配置已存在的错误
        const errorMsg = error.message || '保存失败';
        if (errorMsg.includes('已存在')) {
          // 尝试从错误信息中提取交易对和ID
          const symbolMatch = errorMsg.match(/交易对\s+([A-Z]+)\s+的配置已存在/);
          const idMatch = errorMsg.match(/ID:\s*(\d+)/);
          const existingSymbol = symbolMatch ? symbolMatch[1] : editForm.symbol || '全局';

          if (idMatch) {
            const existingId = parseInt(idMatch[1]);
            // 询问用户是否要编辑现有配置
            dialog.warning({
              title: '配置已存在',
              content: h('div', [
                h(
                  'p',
                  { style: { marginBottom: '8px' } },
                  `交易对 ${existingSymbol} 的配置已存在（ID: ${existingId}）。`,
                ),
                h('p', { style: { color: '#666', fontSize: '13px' } }, '您可以选择：'),
                h(
                  'ul',
                  {
                    style: {
                      margin: '8px 0',
                      paddingLeft: '20px',
                      color: '#666',
                      fontSize: '13px',
                    },
                  },
                  [h('li', '编辑现有配置（推荐）'), h('li', '删除现有配置后重新创建')],
                ),
              ]),
              positiveText: '编辑现有配置',
              negativeText: '取消',
              onPositiveClick: async () => {
                // 重新加载数据并打开编辑弹窗
                await loadData();
                const existingRow = list.value.find((item) => item.id === existingId);
                if (existingRow) {
                  openEditModal(existingRow);
                  message.info('已打开现有配置的编辑窗口');
                } else {
                  message.warning('未找到现有配置，请刷新页面后重试');
                }
              },
            });
          } else {
            message.error(errorMsg);
          }
        } else {
          message.error(errorMsg);
        }
      } finally {
        saving.value = false;
      }
    });
  };

  // 批量保存
  const handleBatchSave = async () => {
    if (!batchFormRef.value) return;
    await batchFormRef.value.validate(async (errors: any) => {
      if (errors) return;

      if (batchEditForm.symbols.length === 0) {
        message.warning('请选择至少一个交易对');
        return;
      }

      saving.value = true;
      try {
        const res = await ToogoVolatilityConfigApi.batchEdit({
          symbols: batchEditForm.symbols,
          highVolatilityThreshold: batchEditForm.highVolatilityThreshold,
          lowVolatilityThreshold: batchEditForm.lowVolatilityThreshold,
          trendStrengthThreshold: batchEditForm.trendStrengthThreshold,
          dThreshold: batchEditForm.dThreshold,
          delta1m: batchEditForm.delta1m,
          delta5m: batchEditForm.delta5m,
          delta15m: batchEditForm.delta15m,
          delta30m: batchEditForm.delta30m,
          delta1h: batchEditForm.delta1h,
          weight1m: batchEditForm.weight1m,
          weight5m: batchEditForm.weight5m,
          weight15m: batchEditForm.weight15m,
          weight30m: batchEditForm.weight30m,
          weight1h: batchEditForm.weight1h,
          isActive: batchEditForm.isActive,
        });

        if (res.code === 0) {
          message.success(`成功为 ${batchEditForm.symbols.length} 个交易对设置配置`);
          showBatchEditModal.value = false;
          loadData();
        } else {
          message.error(res.msg || '批量保存失败');
        }
      } catch (error: any) {
        message.error(error.message || '批量保存失败');
      } finally {
        saving.value = false;
      }
    });
  };

  // 删除
  const handleDelete = (row: any) => {
    dialog.warning({
      title: '确认删除',
      content: `确定要删除 ${row.symbol || '全局'} 配置吗？`,
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          const res = await ToogoVolatilityConfigApi.delete({ id: row.id });
          if (res.code === 0) {
            message.success('删除成功');
            loadData();
          } else {
            message.error(res.msg || '删除失败');
          }
        } catch (error: any) {
          message.error(error.message || '删除失败');
        }
      },
    });
  };

  // 分页
  const handlePageChange = (page: number) => {
    pagination.page = page;
    loadData();
  };

  const handlePageSizeChange = (pageSize: number) => {
    pagination.pageSize = pageSize;
    pagination.page = 1;
    loadData();
  };

  // 自动平衡权重
  const autoBalanceWeights = () => {
    const currentSum = weightSum.value;
    if (currentSum === 0) {
      // 如果总和为0，使用默认权重
      editForm.weight1m = 0.1;
      editForm.weight5m = 0.2;
      editForm.weight15m = 0.25;
      editForm.weight30m = 0.25;
      editForm.weight1h = 0.2;
    } else {
      // 按比例缩放，使总和为1
      editForm.weight1m = Math.round((editForm.weight1m / currentSum) * 100) / 100;
      editForm.weight5m = Math.round((editForm.weight5m / currentSum) * 100) / 100;
      editForm.weight15m = Math.round((editForm.weight15m / currentSum) * 100) / 100;
      editForm.weight30m = Math.round((editForm.weight30m / currentSum) * 100) / 100;
      editForm.weight1h = Math.round((editForm.weight1h / currentSum) * 100) / 100;
    }
    message.success('权重已自动平衡');
  };

  // 批量编辑的自动平衡
  const autoBalanceBatchWeights = () => {
    const currentSum = batchWeightSum.value;
    if (currentSum === 0) {
      batchEditForm.weight1m = 0.1;
      batchEditForm.weight5m = 0.2;
      batchEditForm.weight15m = 0.25;
      batchEditForm.weight30m = 0.25;
      batchEditForm.weight1h = 0.2;
    } else {
      batchEditForm.weight1m = Math.round((batchEditForm.weight1m / currentSum) * 100) / 100;
      batchEditForm.weight5m = Math.round((batchEditForm.weight5m / currentSum) * 100) / 100;
      batchEditForm.weight15m = Math.round((batchEditForm.weight15m / currentSum) * 100) / 100;
      batchEditForm.weight30m = Math.round((batchEditForm.weight30m / currentSum) * 100) / 100;
      batchEditForm.weight1h = Math.round((batchEditForm.weight1h / currentSum) * 100) / 100;
    }
    message.success('权重已自动平衡');
  };

  // 复制配置
  const copyConfig = (row: any) => {
    Object.assign(editForm, {
      id: 0,
      symbol: null, // 复制时清空交易对，让用户选择
      highVolatilityThreshold: row.highVolatilityThreshold || 2.0,
      lowVolatilityThreshold: row.lowVolatilityThreshold || 1.0,
      trendStrengthThreshold: row.trendStrengthThreshold || 1.2,
      dThreshold: row.dThreshold ?? 0.7,
      delta1m: row.delta1m ?? 2.0,
      delta5m: row.delta5m ?? 2.0,
      delta15m: row.delta15m ?? 3.0,
      delta30m: row.delta30m ?? 3.0,
      delta1h: row.delta1h ?? 5.0,
      weight1m: row.weight1m ?? 0.2,
      weight5m: row.weight5m ?? 0.25,
      weight15m: row.weight15m ?? 0.25,
      weight30m: row.weight30m ?? 0.2,
      weight1h: row.weight1h ?? 0.1,
      isActive: row.isActive ?? 1,
    });
    showEditModal.value = true;
    message.info('已复制配置，请选择交易对后保存');
  };

  // 从全局配置复制
  const copyFromGlobal = () => {
    const globalConfig = list.value.find((item) => item.symbol === null);
    if (globalConfig) {
      Object.assign(editForm, {
        highVolatilityThreshold: globalConfig.highVolatilityThreshold || 2.0,
        lowVolatilityThreshold: globalConfig.lowVolatilityThreshold || 1.0,
        trendStrengthThreshold: globalConfig.trendStrengthThreshold || 1.2,
        dThreshold: globalConfig.dThreshold ?? 0.7,
        delta1m: globalConfig.delta1m ?? 2.0,
        delta5m: globalConfig.delta5m ?? 2.0,
        delta15m: globalConfig.delta15m ?? 3.0,
        delta30m: globalConfig.delta30m ?? 3.0,
        delta1h: globalConfig.delta1h ?? 5.0,
        weight1m: globalConfig.weight1m ?? 0.2,
        weight5m: globalConfig.weight5m ?? 0.25,
        weight15m: globalConfig.weight15m ?? 0.25,
        weight30m: globalConfig.weight30m ?? 0.2,
        weight1h: globalConfig.weight1h ?? 0.1,
      });
      message.success('已从全局配置复制参数');
    } else {
      message.warning('未找到全局配置');
    }
  };

  // 获取市场状态预览类型
  const getMarketStatePreview = (type: string) => {
    switch (type) {
      case 'high_vol':
        return 'warning';
      case 'low_vol':
        return 'info';
      case 'trend':
        return 'success';
      default:
        return 'default';
    }
  };

  // 批量启用
  const batchEnable = async () => {
    if (selectedRowKeys.value.length === 0) {
      message.warning('请先选择要启用的配置');
      return;
    }
    try {
      saving.value = true;
      // 这里需要后端支持批量更新接口，暂时逐个更新
      for (const id of selectedRowKeys.value) {
        const row = list.value.find((item) => item.id === id);
        if (row && row.symbol !== null) {
          await ToogoVolatilityConfigApi.update({
            id,
            symbol: row.symbol,
            highVolatilityThreshold: row.highVolatilityThreshold,
            lowVolatilityThreshold: row.lowVolatilityThreshold,
            trendStrengthThreshold: row.trendStrengthThreshold,
            dThreshold: row.dThreshold ?? 0.7,
            delta1m: row.delta1m ?? 2.0,
            delta5m: row.delta5m ?? 2.0,
            delta15m: row.delta15m ?? 3.0,
            delta30m: row.delta30m ?? 3.0,
            delta1h: row.delta1h ?? 5.0,
            weight1m: row.weight1m,
            weight5m: row.weight5m,
            weight15m: row.weight15m,
            weight30m: row.weight30m,
            weight1h: row.weight1h,
            isActive: 1,
          });
        }
      }
      message.success(`成功启用 ${selectedRowKeys.value.length} 个配置`);
      selectedRowKeys.value = [];
      loadData();
    } catch (error: any) {
      message.error(error.message || '批量启用失败');
    } finally {
      saving.value = false;
    }
  };

  // 批量禁用
  const batchDisable = async () => {
    if (selectedRowKeys.value.length === 0) {
      message.warning('请先选择要禁用的配置');
      return;
    }
    try {
      saving.value = true;
      for (const id of selectedRowKeys.value) {
        const row = list.value.find((item) => item.id === id);
        if (row && row.symbol !== null) {
          await ToogoVolatilityConfigApi.update({
            id,
            symbol: row.symbol,
            highVolatilityThreshold: row.highVolatilityThreshold,
            lowVolatilityThreshold: row.lowVolatilityThreshold,
            trendStrengthThreshold: row.trendStrengthThreshold,
            dThreshold: row.dThreshold ?? 0.7,
            delta1m: row.delta1m ?? 2.0,
            delta5m: row.delta5m ?? 2.0,
            delta15m: row.delta15m ?? 3.0,
            delta30m: row.delta30m ?? 3.0,
            delta1h: row.delta1h ?? 5.0,
            weight1m: row.weight1m,
            weight5m: row.weight5m,
            weight15m: row.weight15m,
            weight30m: row.weight30m,
            weight1h: row.weight1h,
            isActive: 0,
          });
        }
      }
      message.success(`成功禁用 ${selectedRowKeys.value.length} 个配置`);
      selectedRowKeys.value = [];
      loadData();
    } catch (error: any) {
      message.error(error.message || '批量禁用失败');
    } finally {
      saving.value = false;
    }
  };

  // 批量删除
  const batchDelete = () => {
    if (selectedRowKeys.value.length === 0) {
      message.warning('请先选择要删除的配置');
      return;
    }
    dialog.warning({
      title: '确认批量删除',
      content: `确定要删除选中的 ${selectedRowKeys.value.length} 个配置吗？`,
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          saving.value = true;
          for (const id of selectedRowKeys.value) {
            await ToogoVolatilityConfigApi.delete({ id });
          }
          message.success(`成功删除 ${selectedRowKeys.value.length} 个配置`);
          selectedRowKeys.value = [];
          loadData();
        } catch (error: any) {
          message.error(error.message || '批量删除失败');
        } finally {
          saving.value = false;
        }
      },
    });
  };

  // 初始化
  onMounted(() => {
    loadData();
  });
</script>

<style scoped lang="less">
  .volatility-config-page {
    .market-state-tips {
      .tip-title {
        font-weight: 600;
        margin-bottom: 12px;
        display: flex;
        align-items: center;
        font-size: 15px;
        color: #333;
      }
      .tip-content {
        display: flex;
        flex-direction: column;
        gap: 10px;
        margin-bottom: 12px;
      }
      .tip-row {
        display: flex;
        align-items: center;
        gap: 12px;
        font-size: 13px;
        padding: 8px 12px;
        background: #fafafa;
        border-radius: 6px;
        transition: all 0.3s;
        &:hover {
          background: #f0f0f0;
        }
        span {
          flex: 1;
          color: #666;
        }
      }
      .tip-formula {
        margin-top: 12px;
        padding: 12px;
        background: #f8f9fa;
        border-left: 3px solid #18a058;
        border-radius: 6px;
        font-size: 12px;
        line-height: 1.8;
      }
      .tip-note {
        margin-top: 12px;
        padding-top: 12px;
        border-top: 1px solid #e0e0e0;
      }
    }
    .config-preview {
      margin-top: 8px;
    }
    .weight-slider-container {
      display: flex;
      flex-direction: column;
      gap: 12px;
      .weight-value {
        display: flex;
        align-items: center;
        gap: 12px;
        .weight-percent {
          font-size: 13px;
          color: #666;
          font-weight: 600;
          min-width: 45px;
          padding: 4px 10px;
          background: #f5f5f5;
          border-radius: 4px;
          text-align: center;
        }
      }
    }
  }

  :deep(.volatility-config-modal) {
    .n-card {
      box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
      border-radius: 8px;
    }
    .n-card-header {
      padding: 18px 24px;
      border-bottom: 1px solid #f0f0f0;
      font-weight: 600;
      font-size: 15px;
      background: #fafafa;
    }
    .n-card-body {
      padding: 24px;
    }
    .n-form-item {
      margin-bottom: 0;
    }
    .n-form-item-label {
      font-weight: 500;
      color: #333;
      font-size: 13px;
      margin-bottom: 8px;
    }
    .n-input-number {
      width: 100%;
    }
    .n-grid {
      margin-bottom: 0;
    }
  }
</style>
