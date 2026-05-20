<template>
  <div class="flex flex-col h-full">
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

    <!-- Verifying key existence (clears old content immediately) -->
    <div v-else-if="verifying" class="flex-1 flex items-center justify-center">
      <div class="text-center text-text-muted">
        <RefreshCw class="w-6 h-6 mx-auto mb-2 text-accent-blue animate-spin" />
        <p class="text-[13px]">Loading...</p>
      </div>
    </div>

    <!-- Key not found -->
    <div v-else-if="keyNotFound" class="flex-1 flex items-center justify-center">
      <div class="text-center max-w-[320px]">
        <AlertTriangle class="w-10 h-10 mx-auto mb-3 text-accent-orange opacity-80" />
        <p class="text-sm text-text-secondary mb-1">Key not found</p>
        <p class="text-[13px] text-text-muted mb-4 break-all">
          <span class="font-mono text-accent-red">{{ keyStore.activeTab }}</span>
        </p>
        <button
          class="px-4 py-1.5 text-[13px] rounded border border-gray-300 text-text-secondary hover:bg-gray-100 transition-colors"
          @click="refreshTree"
        >
          <RefreshCw class="w-3.5 h-3.5 inline mr-1.5" />
          Refresh Tree
        </button>
      </div>
    </div>

    <!-- Normal viewer -->
    <div v-else class="flex-1 overflow-hidden">
      <component
        :is="viewerComponent"
        :conn-id="props.connId"
        :key-name="keyStore.activeTab"
        :key="keyStore.activeTab + ':' + refreshCounter"
        @refresh="refreshKey"
        @delete="deleteCurrentKey"
      />
    </div>

    <!-- Delete Confirm Dialog -->
    <NModal v-if="showDeleteConfirm" v-model:show="showDeleteConfirm" preset="card" title="Delete Key" style="width: 400px" :bordered="false">
      <p class="text-[13px] text-text-secondary">
        Are you sure you want to delete <span class="text-accent-red font-mono">{{ keyStore.activeTab }}</span>?
      </p>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button class="px-4 py-1.5 text-[13px] rounded-md border border-gray-300 text-text-secondary hover:bg-gray-100 transition-colors" @click="showDeleteConfirm = false">Cancel</button>
          <button class="btn-danger" @click="confirmDeleteKey">Delete</button>
        </div>
      </template>
    </NModal>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, onMounted } from 'vue'
import { Database, MousePointerClick, RefreshCw, AlertTriangle } from 'lucide-vue-next'
import { useKeyStore } from '@/stores/key'
import StringViewer from '@/components/StringViewer.vue'
import HashViewer from '@/components/HashViewer.vue'
import ListViewer from '@/components/ListViewer.vue'
import SetViewer from '@/components/SetViewer.vue'
import ZSetViewer from '@/components/ZSetViewer.vue'

const props = defineProps<{ connId: string }>()
const keyStore = useKeyStore()

const refreshCounter = ref(0)
const showDeleteConfirm = ref(false)
const verifying = ref(false)
const keyNotFound = ref(false)

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

/** Verify a key exists in the current DB, return its type or empty on failure */
async function verifyKey(): Promise<string> {
  if (!keyStore.activeTab || !props.connId) return ''
  const keyType = await window.go.main.App.GetKeyType(props.connId, keyStore.activeTab)
  if (!keyType || keyType === 'none') return ''
  keyStore.keyTypes[keyStore.activeTab] = keyType
  return keyType
}

/** Watch activeTab change: clear old content → verify → show viewer or error */
watch(() => keyStore.activeTab, async (newKey) => {
  if (!newKey || !props.connId) {
    keyNotFound.value = false
    return
  }
  verifying.value = true
  keyNotFound.value = false
  const keyType = await verifyKey()
  if (!keyType) {
    keyNotFound.value = true
  }
  verifying.value = false
})

function refreshKey() {
  // Force re-mount viewer component and re-verify key existence
  if (!keyStore.activeTab || !props.connId) return
  verifying.value = true
  keyNotFound.value = false
  verifyKey().then((keyType) => {
    if (keyType) {
      refreshCounter.value++
    } else {
      keyNotFound.value = true
    }
    verifying.value = false
  })
}

async function refreshTree() {
  keyNotFound.value = false
  if (props.connId) {
    keyStore.rootNodes = []
    await keyStore.loadRoot(props.connId)
  }
}

function deleteCurrentKey() {
  if (!keyStore.activeTab) return
  showDeleteConfirm.value = true
}

async function confirmDeleteKey() {
  if (!keyStore.activeTab) return
  await window.go.main.App.DeleteKey(props.connId, keyStore.activeTab)
  keyStore.closeTab(keyStore.activeTab)
  showDeleteConfirm.value = false
}

// Verify current key on initial mount (in case activeTab was already set)
onMounted(() => {
  if (keyStore.activeTab && props.connId) {
    verifying.value = true
    verifyKey().then((keyType) => {
      if (!keyType) keyNotFound.value = true
      verifying.value = false
    })
  }
})
</script>
