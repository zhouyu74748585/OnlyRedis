<template>
  <div class="flex flex-col h-full">
    <!-- Tabs Bar -->
    <div class="flex items-center bg-bg-secondary border-b border-gray-200 min-h-[36px]">
      <div class="flex items-center flex-1 overflow-x-auto">
        <div
          v-for="tab in keyStore.openedTabs"
          :key="tab"
          class="flex items-center gap-1.5 px-3 py-2 text-[13px] cursor-pointer border-r border-gray-200 transition-colors flex-shrink-0"
          :class="keyStore.activeTab === tab ? 'bg-bg-primary text-accent-blue border-t-2 border-t-accent-blue' : 'text-text-secondary hover:bg-gray-50'"
          @click="keyStore.setActiveTab(tab)"
        >
          <span class="max-w-[160px] truncate">{{ formatTabName(tab) }}</span>
          <button
            class="p-0.5 rounded hover:bg-gray-200 transition-colors"
            @click.stop="keyStore.closeTab(tab)"
          >
            <X class="w-3 h-3" />
          </button>
        </div>
      </div>
      <div v-if="keyStore.activeTab" class="flex items-center gap-1.5 px-2 flex-shrink-0">
        <button class="p-1 rounded hover:bg-gray-200 transition-colors" title="Refresh" @click="refreshKey">
          <RefreshCw class="w-3.5 h-3.5 text-text-secondary" />
        </button>
        <button class="p-1 rounded hover:bg-red-50 transition-colors" title="Delete" @click="deleteCurrentKey">
          <Trash class="w-3.5 h-3.5 text-accent-red/70" />
        </button>
      </div>
    </div>

    <!-- Content Area -->
    <div v-if="!props.connId" class="flex-1 flex items-center justify-center">
      <div class="text-center text-text-muted">
        <Database class="w-12 h-12 mx-auto mb-3 opacity-20" />
        <p class="text-sm">Connect to a Redis server to browse data</p>
      </div>
    </div>
    <div v-else-if="!keyStore.activeTab" class="flex-1 flex items-center justify-center">
      <div class="text-center text-text-muted">
        <MousePointerClick class="w-12 h-12 mx-auto mb-3 opacity-20" />
        <p class="text-sm">Select a key from the left panel</p>
      </div>
    </div>
    <div v-else class="flex-1 overflow-hidden">
      <component
        :is="viewerComponent"
        :conn-id="props.connId"
        :key-name="keyStore.activeTab"
        :key="keyStore.activeTab + ':' + refreshCounter"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { X, RefreshCw, Trash, Database, MousePointerClick } from 'lucide-vue-next'
import { useKeyStore } from '@/stores/key'
import StringViewer from '@/components/StringViewer.vue'
import HashViewer from '@/components/HashViewer.vue'
import ListViewer from '@/components/ListViewer.vue'
import SetViewer from '@/components/SetViewer.vue'
import ZSetViewer from '@/components/ZSetViewer.vue'

const props = defineProps<{ connId: string }>()
const keyStore = useKeyStore()

const refreshCounter = ref(0)

const typeMap: Record<string, any> = {
  string: StringViewer,
  hash: HashViewer,
  list: ListViewer,
  set: SetViewer,
  zset: ZSetViewer,
}

/** Use pre-cached type from store (set by TreeNode on click), fallback to string */
const viewerComponent = computed(() => {
  const keyType = keyStore.keyTypes[keyStore.activeTab] || 'string'
  return typeMap[keyType] || StringViewer
})

function formatTabName(key: string) {
  const parts = key.split(':')
  return parts.length > 3 ? '...' + parts.slice(-2).join(':') : key
}

function refreshKey() {
  // Force re-mount viewer component by incrementing refresh counter
  if (keyStore.activeTab) {
    refreshCounter.value++
  }
}

async function deleteCurrentKey() {
  if (!keyStore.activeTab) return
  await window.go.main.App.DeleteKey(props.connId, keyStore.activeTab)
  keyStore.closeTab(keyStore.activeTab)
}
</script>
