<template>
  <div class="flex flex-col h-screen w-screen overflow-hidden bg-bg-primary">
    <!-- Custom Title Bar (draggable, with centered connection info) -->
    <div
      class="h-[38px] flex-shrink-0 flex items-center border-b border-gray-200 select-none relative"
      style="--wails-draggable: drag"
    >
      <!-- Traffic light spacer (macOS traffic lights are ~70px wide) -->
      <div class="absolute left-0 top-0 w-[70px] h-full" style="--wails-draggable: no-drag" />
      <!-- Centered connection info -->
      <div class="flex-1 flex items-center justify-center">
        <template v-if="activeConn">
          <div class="flex items-center gap-2 text-[12px]">
            <span class="w-2 h-2 rounded-full bg-green-500 flex-shrink-0" />
            <span class="text-text-primary font-medium">{{ activeConn.name }}</span>
            <span class="text-text-muted">·</span>
            <span class="text-text-muted">{{ activeConn.host }}:{{ activeConn.port }}</span>
            <span class="text-text-muted">·</span>
            <span class="text-text-secondary">
              DB <span class="text-accent-blue font-mono">{{ keyStore.currentDb }}</span>
            </span>
          </div>
        </template>
        <span v-else class="text-[12px] text-text-muted">No connection selected</span>
      </div>
    </div>

    <!-- Main Content (3-column layout) -->
    <div class="flex flex-1 min-h-0">
      <!-- Left Sidebar (resizable width) -->
      <div class="flex flex-col border-r border-gray-200 flex-shrink-0" :style="{ width: leftWidth + 'px', minWidth: '200px' }">
        <!-- Connection Panel (fixed height, resizable) -->
        <div :style="{ height: connectionPanelHeight + 'px' }" class="flex-shrink-0 min-h-[80px] overflow-hidden">
          <ConnectionPanel @select="onSelectConnection" />
        </div>
        <!-- Horizontal resize handle -->
        <div
          class="h-[5px] flex-shrink-0 cursor-row-resize z-10 hover:bg-accent-blue/20 transition-colors flex items-center justify-center group"
          @mousedown="startHorizontalResize"
        >
          <div class="w-8 h-0.5 rounded bg-gray-300 group-hover:bg-accent-blue/50 transition-colors" />
        </div>
        <!-- Key Tree -->
        <div class="flex-1 min-h-[60px] overflow-hidden">
          <KeyTree :conn-id="activeConnId" />
        </div>
      </div>

      <!-- Vertical resize handle (left | center) -->
      <div
        class="w-[5px] cursor-col-resize z-10 hover:bg-accent-blue/20 transition-colors flex items-center justify-center group flex-shrink-0"
        @mousedown="startLeftResize"
      >
        <div class="h-8 w-0.5 rounded bg-gray-300 group-hover:bg-accent-blue/50 transition-colors" />
      </div>

      <!-- Center Content -->
      <div class="flex-1 flex flex-col min-w-[200px] relative">
        <DataViewer :conn-id="activeConnId" />

        <!-- Monitor toggle button (on right edge of center) -->
        <div
          class="absolute right-0 top-1/2 -translate-y-1/2 z-20 flex items-center"
          @click="toggleMonitor"
        >
          <button
            class="w-4 h-12 flex items-center justify-center rounded-l-md bg-gray-200 hover:bg-accent-blue/20 border border-r-0 border-gray-300 transition-colors group"
            :title="monitorCollapsed ? 'Show Monitor' : 'Hide Monitor'"
          >
            <ChevronRight
              v-if="monitorCollapsed"
              class="w-3 h-3 text-text-muted group-hover:text-accent-blue transition-colors"
            />
            <ChevronLeft
              v-else
              class="w-3 h-3 text-text-muted group-hover:text-accent-blue transition-colors"
            />
          </button>
        </div>
      </div>

      <!-- Vertical resize handle (center | right) - hidden when collapsed -->
      <div
        v-if="!monitorCollapsed"
        class="w-[5px] cursor-col-resize z-10 hover:bg-accent-blue/20 transition-colors flex items-center justify-center group flex-shrink-0"
        @mousedown="startRightResize"
      >
        <div class="h-8 w-0.5 rounded bg-gray-300 group-hover:bg-accent-blue/50 transition-colors" />
      </div>

      <!-- Right Monitor Panel -->
      <div
        class="border-l border-gray-200 flex-shrink-0 overflow-hidden transition-all duration-200 ease-in-out"
        :class="monitorCollapsed ? 'border-l-0' : ''"
        :style="{
          width: monitorCollapsed ? '0px' : rightWidth + 'px',
          minWidth: monitorCollapsed ? '0px' : '240px',
        }"
      >
        <MonitorPanel v-if="!monitorCollapsed" :conn-id="activeConnId" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ChevronLeft, ChevronRight } from 'lucide-vue-next'
import { useConnectionStore } from '@/stores/connection'
import { useKeyStore } from '@/stores/key'
import ConnectionPanel from '@/components/ConnectionPanel.vue'
import KeyTree from '@/components/KeyTree.vue'
import DataViewer from '@/components/DataViewer.vue'
import MonitorPanel from '@/components/MonitorPanel.vue'

const connectionStore = useConnectionStore()
const keyStore = useKeyStore()

const activeConnId = ref('')

const activeConn = computed(() => connectionStore.connections.find(c => c.id === activeConnId.value) || null)

const leftWidth = ref(260)
const rightWidth = ref(320)
const connectionPanelHeight = ref(200)
const monitorCollapsed = ref(false)
let savedRightWidth = 320

function onSelectConnection(id: string) {
  activeConnId.value = id
}

// Update native window title when connection or DB changes
watch(
  () => [activeConn.value?.name, activeConn.value?.host, activeConn.value?.port, keyStore.currentDb] as const,
  ([name, host, port, db]) => {
    if (name && host) {
      window.go.main.App.SetWindowTitle(`${name} · ${host}:${port} · DB${db}`)
    } else {
      window.go.main.App.SetWindowTitle('onlyRedis')
    }
  }
)

function toggleMonitor() {
  monitorCollapsed.value = !monitorCollapsed.value
  if (monitorCollapsed.value) {
    savedRightWidth = rightWidth.value
    rightWidth.value = 0
  } else {
    rightWidth.value = savedRightWidth
  }
}

// === Horizontal resize (ConnectionPanel ↔ KeyTree) ===
function startHorizontalResize(e: MouseEvent) {
  e.preventDefault()
  const startY = e.clientY
  const startHeight = connectionPanelHeight.value

  function onMove(ev: MouseEvent) {
    const deltaY = ev.clientY - startY
    connectionPanelHeight.value = Math.max(80, Math.min(500, startHeight + deltaY))
  }

  function onUp() {
    document.removeEventListener('mousemove', onMove)
    document.removeEventListener('mouseup', onUp)
    document.body.style.cursor = ''
    document.body.style.userSelect = ''
  }

  document.addEventListener('mousemove', onMove)
  document.addEventListener('mouseup', onUp)
  document.body.style.cursor = 'row-resize'
  document.body.style.userSelect = 'none'
}

// === Vertical resize: left sidebar width ===
function startLeftResize(e: MouseEvent) {
  e.preventDefault()
  const startX = e.clientX
  const startWidth = leftWidth.value

  function onMove(ev: MouseEvent) {
    const delta = ev.clientX - startX
    leftWidth.value = Math.max(200, Math.min(500, startWidth + delta))
  }

  function onUp() {
    document.removeEventListener('mousemove', onMove)
    document.removeEventListener('mouseup', onUp)
    document.body.style.cursor = ''
    document.body.style.userSelect = ''
  }

  document.addEventListener('mousemove', onMove)
  document.addEventListener('mouseup', onUp)
  document.body.style.cursor = 'col-resize'
  document.body.style.userSelect = 'none'
}

// === Vertical resize: right monitor width ===
function startRightResize(e: MouseEvent) {
  e.preventDefault()
  const startX = e.clientX
  const startWidth = rightWidth.value

  function onMove(ev: MouseEvent) {
    const delta = startX - ev.clientX
    rightWidth.value = Math.max(240, Math.min(500, startWidth + delta))
  }

  function onUp() {
    document.removeEventListener('mousemove', onMove)
    document.removeEventListener('mouseup', onUp)
    document.body.style.cursor = ''
    document.body.style.userSelect = ''
  }

  document.addEventListener('mousemove', onMove)
  document.addEventListener('mouseup', onUp)
  document.body.style.cursor = 'col-resize'
  document.body.style.userSelect = 'none'
}
</script>
