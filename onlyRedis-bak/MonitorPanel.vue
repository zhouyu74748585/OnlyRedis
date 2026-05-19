<template>
  <div class="flex flex-col h-full">
    <!-- Header -->
    <div class="flex items-center justify-between px-3 py-2 border-b border-gray-200">
      <div class="flex items-center gap-2">
        <Activity class="w-4 h-4 text-accent-blue" />
        <span class="text-[13px] font-semibold uppercase text-text-secondary tracking-wider">Monitor</span>
      </div>
      <div class="flex items-center gap-2">
          <button
            v-if="!store.isMonitoring && props.connId"
            class="w-6 h-6 flex items-center justify-center rounded hover:bg-green-50 transition-colors"
            @click="startMonitor"
          >
          <Play class="w-3.5 h-3.5 text-accent-green" />
        </button>
          <button
            v-else-if="store.isMonitoring"
            class="w-6 h-6 flex items-center justify-center rounded hover:bg-red-50 transition-colors"
            @click="stopMonitor"
          >
          <Pause class="w-3.5 h-3.5 text-accent-red" />
        </button>
      </div>
    </div>

    <div v-if="!props.connId" class="flex-1 flex items-center justify-center text-[13px] text-text-muted">
      Connect to a server
    </div>
    <div v-else-if="!store.data" class="flex-1 flex flex-col items-center justify-center gap-2 text-text-muted">
      <BarChart class="w-10 h-10 opacity-20" />
      <span class="text-[13px]">Click play to start monitoring</span>
    </div>
    <template v-else>
      <!-- Metric Cards -->
      <div class="grid grid-cols-2 gap-2 p-3">
        <div class="glass-panel p-3">
          <div class="text-[12px] text-text-muted mb-1">Memory</div>
          <div class="text-base font-semibold text-accent-blue">{{ store.data.usedMemoryHuman }}</div>
          <div class="text-[12px] text-text-muted mt-1">Peak: {{ store.data.peakMemoryHuman }}</div>
        </div>
        <div class="glass-panel p-3">
          <div class="text-[12px] text-text-muted mb-1">Clients</div>
          <div class="text-base font-semibold text-accent-green">{{ store.data.connectedClients }}</div>
          <div class="text-[12px] text-text-muted mt-1">Blocked: {{ store.data.blockedClients }}</div>
        </div>
        <div class="glass-panel p-3">
          <div class="text-[12px] text-text-muted mb-1">Hit Rate</div>
          <div class="text-base font-semibold" :class="store.data.hitRate > 90 ? 'text-accent-green' : 'text-accent-orange'">
            {{ store.data.hitRate?.toFixed(1) }}%
          </div>
        </div>
        <div class="glass-panel p-3">
          <div class="text-[12px] text-text-muted mb-1">QPS</div>
          <div class="text-base font-semibold text-accent-blue">{{ store.data.qps?.toLocaleString() }}</div>
          <div class="text-[12px] text-text-muted mt-1">{{ store.data.opsPerSec?.toLocaleString() }} ops/s</div>
        </div>
      </div>

      <!-- QPS Chart -->
      <div class="px-3 mb-2">
        <div class="text-[12px] text-text-muted mb-1.5">QPS Trend</div>
        <div class="h-[140px]">
          <v-chart :option="qpsChartOption" autoresize />
        </div>
      </div>

      <!-- Info Stats -->
      <div class="flex-1 overflow-auto px-3 pb-2">
        <div class="text-[12px] text-text-muted mb-1.5">Server Info</div>
        <div class="space-y-0.5">
          <div class="flex justify-between text-[13px] py-0.5 border-b border-gray-100">
            <span class="text-text-muted">Keys</span>
            <span class="text-text-secondary">{{ store.data.keysCount?.toLocaleString() }}</span>
          </div>
          <div class="flex justify-between text-[13px] py-0.5 border-b border-gray-100">
            <span class="text-text-muted">Expires</span>
            <span class="text-text-secondary">{{ store.data.expiresCount?.toLocaleString() }}</span>
          </div>
          <div class="flex justify-between text-[13px] py-0.5 border-b border-gray-100">
            <span class="text-text-muted">Avg TTL</span>
            <span class="text-text-secondary">{{ store.data.avgTTL?.toLocaleString() }}ms</span>
          </div>
          <div class="flex justify-between text-[13px] py-0.5 border-b border-gray-100">
            <span class="text-text-muted">CPU Sys</span>
            <span class="text-text-secondary">{{ store.data.cpuSys?.toFixed(2) }}</span>
          </div>
          <div class="flex justify-between text-[13px] py-0.5 border-b border-gray-100">
            <span class="text-text-muted">CPU User</span>
            <span class="text-text-secondary">{{ store.data.cpuUser?.toFixed(2) }}</span>
          </div>
          <div class="flex justify-between text-[13px] py-0.5 border-b border-gray-100">
            <span class="text-text-muted">Net In</span>
            <span class="text-text-secondary">{{ formatBytes(store.data.netInputRate) }}/s</span>
          </div>
          <div class="flex justify-between text-[13px] py-0.5 border-b border-gray-100">
            <span class="text-text-muted">Net Out</span>
            <span class="text-text-secondary">{{ formatBytes(store.data.netOutputRate) }}/s</span>
          </div>
          <div class="flex justify-between text-[13px] py-0.5">
            <span class="text-text-muted">Uptime</span>
            <span class="text-text-secondary">{{ formatUptime(store.data.uptimeInSeconds) }}</span>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onUnmounted, watch } from 'vue'
import { Activity, Play, Pause, BarChart } from 'lucide-vue-next'
import { useMonitorStore } from '@/stores/monitor'
import VChart from 'vue-echarts'
import 'echarts'

const props = defineProps<{ connId: string }>()
const store = useMonitorStore()

let monitorTimer: ReturnType<typeof setInterval> | null = null

const qpsChartOption = computed(() => ({
  backgroundColor: 'transparent',
  grid: { top: 5, right: 10, bottom: 20, left: 40 },
  xAxis: {
    type: 'category',
    data: store.qpsHistory.map((p) => p.time),
    axisLine: { lineStyle: { color: '#d1d5db' } },
    axisLabel: { color: '#6b7280', fontSize: 11 },
  },
  yAxis: {
    type: 'value',
    splitLine: { lineStyle: { color: 'rgba(0,0,0,0.06)' } },
    axisLabel: { color: '#6b7280', fontSize: 11 },
  },
  series: [{
    data: store.qpsHistory.map((p) => p.qps),
    type: 'line',
    smooth: true,
    symbol: 'none',
    lineStyle: { color: '#60a5fa', width: 2 },
    areaStyle: { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: 'rgba(96,165,250,0.2)' }, { offset: 1, color: 'rgba(96,165,250,0)' }] } },
  }],
}))

async function startMonitor() {
  await store.startMonitoring(props.connId)
  monitorTimer = setInterval(() => {
    store.fetchData(props.connId)
  }, store.interval * 1000)
  store.fetchData(props.connId)
}

async function stopMonitor() {
  if (monitorTimer) {
    clearInterval(monitorTimer)
    monitorTimer = null
  }
  await store.stopMonitoring(props.connId)
}

function formatBytes(bytes: number): string {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

function formatUptime(seconds: number): string {
  const d = Math.floor(seconds / 86400)
  const h = Math.floor((seconds % 86400) / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  if (d > 0) return `${d}d ${h}h`
  if (h > 0) return `${h}h ${m}m`
  return `${m}m`
}

onUnmounted(() => {
  if (monitorTimer) clearInterval(monitorTimer)
})
</script>
