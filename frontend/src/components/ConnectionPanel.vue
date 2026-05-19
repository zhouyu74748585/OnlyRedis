<template>
  <div class="flex flex-col h-full border-b border-gray-200">
    <!-- Header -->
    <div class="flex items-center justify-between px-3 py-2">
      <div class="flex items-center gap-2">
        <Database class="w-4 h-4 text-accent-blue" />
        <span class="text-[13px] font-semibold uppercase text-text-secondary tracking-wider">Connections</span>
      </div>
      <button
        class="w-7 h-7 flex items-center justify-center rounded-md hover:bg-gray-100 transition-colors"
        @click="showDialog = true"
      >
        <Plus class="w-4 h-4 text-text-secondary" />
      </button>
    </div>

    <!-- Connection List -->
    <div class="flex-1 overflow-y-auto px-2 space-y-1">
      <div v-if="store.connections.length === 0" class="flex flex-col items-center justify-center h-full text-text-muted gap-2">
        <Plug class="w-8 h-8 opacity-30" />
        <span class="text-[13px] text-text-muted">No connections yet</span>
        <button
          class="text-[13px] text-accent-blue hover:text-accent-blue-hover transition-colors"
          @click="showDialog = true"
        >
          + Add Connection
        </button>
      </div>
      <div
        v-for="conn in store.connections"
        :key="conn.id"
        class="flex items-center gap-2 px-2 py-1.5 rounded-md cursor-pointer transition-all duration-200 group"
        :class="store.activeId === conn.id ? 'bg-blue-50 border border-blue-200' : 'hover:bg-gray-100 border border-transparent'"
        @click="onClick(conn)"
        @contextmenu.prevent="showContextMenu($event, conn)"
      >
        <span
          class="status-dot flex-shrink-0"
          :class="conn.status === 'connected' ? 'connected' : conn.status === 'connecting' ? 'connecting' : 'disconnected'"
          :title="conn.status"
        />
        <div class="flex-1 min-w-0">
          <div class="text-[13px] text-text-primary truncate">{{ conn.name }}</div>
          <div class="text-[12px] text-text-muted truncate">{{ conn.host }}:{{ conn.port }}</div>
        </div>
        <div v-if="store.activeId === conn.id" class="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity">
          <button class="p-0.5 hover:bg-gray-200 rounded" @click.stop="editConnection(conn)">
            <Edit class="w-3.5 h-3.5 text-text-secondary" />
          </button>
          <button class="p-0.5 hover:bg-gray-200 rounded" @click.stop="removeConnection(conn.id)">
            <Trash class="w-3.5 h-3.5 text-accent-red" />
          </button>
        </div>
      </div>
    </div>

    <!-- Connection Dialog -->
    <ConnectionDialog
      v-if="showDialog"
      :edit-data="editData"
      @close="showDialog = false; editData = null"
      @save="onSave"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Database, Plus, Plug, Edit, Trash } from 'lucide-vue-next'
import { useConnectionStore, type RedisConnection } from '@/stores/connection'
import ConnectionDialog from '@/components/ConnectionDialog.vue'

const emit = defineEmits<{ select: [id: string] }>()
const store = useConnectionStore()
const showDialog = ref(false)
const editData = ref<RedisConnection | null>(null)

store.loadConnections()

async function onClick(conn: RedisConnection) {
  const needsConnect = conn.status !== 'connected'
  if (needsConnect) {
    await store.connect(conn.id)
  } else {
    store.setActive(conn.id)
  }
  emit('select', conn.id)
}

function showContextMenu(e: MouseEvent, conn: RedisConnection) {
  // Simple right-click menu will be implemented later
}

function editConnection(conn: RedisConnection) {
  editData.value = { ...conn }
  showDialog.value = true
}

function removeConnection(id: string) {
  store.deleteConnection(id)
}

async function onSave(cfg: RedisConnection) {
  if (editData.value) {
    await store.updateConnection(editData.value.id, cfg)
  } else {
    await store.addConnection(cfg)
  }
  showDialog.value = false
  editData.value = null
}
</script>
