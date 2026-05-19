<template>
  <div class="flex flex-col h-full">
    <!-- Toolbar -->
    <div class="flex items-center justify-between px-3 py-1.5 bg-bg-secondary border-b border-gray-200">
      <div class="flex items-center gap-2">
        <List class="w-4 h-4 text-accent-blue" />
        <span class="text-[13px] text-text-secondary">LIST ({{ items.length }} elements)</span>
      </div>
      <div class="flex items-center gap-2">
        <button class="px-2 py-1 text-[13px] rounded border border-gray-300 text-text-secondary hover:bg-gray-100 transition-colors" @click="pushItem(true)">LPUSH</button>
        <button class="px-2 py-1 text-[13px] rounded border border-gray-300 text-text-secondary hover:bg-gray-100 transition-colors" @click="pushItem(false)">RPUSH</button>
      </div>
    </div>

    <!-- List Items -->
    <div class="flex-1 overflow-auto">
      <div v-if="items.length === 0" class="flex items-center justify-center h-full text-[13px] text-text-muted">
        Empty list
      </div>
      <div
        v-for="(item, idx) in items"
        :key="idx"
        class="flex items-center gap-2 px-3 py-2 border-b border-gray-100 hover:bg-gray-50 transition-colors group"
      >
        <span class="text-[12px] text-text-muted w-8 text-right flex-shrink-0">{{ idx }}</span>
        <span class="flex-1 text-[13px] text-text-primary truncate font-mono">{{ item }}</span>
        <button
          class="p-1 rounded hover:bg-red-50 transition-colors opacity-0 group-hover:opacity-100"
          @click="removeItem(item, 1)"
        >
          <Trash class="w-3.5 h-3.5 text-accent-red/60" />
        </button>
      </div>
    </div>

    <!-- Quick Add -->
    <div class="flex items-center gap-2 px-3 py-1.5 bg-bg-secondary border-t border-gray-200">
      <input
        v-model="newItem"
        class="input-field flex-1 py-1.5"
        placeholder="New element..."
        @keyup.enter="pushItem(false)"
      />
      <button class="btn-primary" @click="pushItem(false)">Push</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { List, Trash } from 'lucide-vue-next'

const props = defineProps<{ connId: string; keyName: string }>()

const items = ref<string[]>([])
const newItem = ref('')

async function loadItems() {
  if (!props.connId || !props.keyName) return
  const raw = await window.go.main.App.ListRange(props.connId, props.keyName, 0, 199)
  items.value = JSON.parse(raw)
}

async function pushItem(left: boolean) {
  if (!newItem.value) return
  await window.go.main.App.ListPush(props.connId, props.keyName, newItem.value, left)
  newItem.value = ''
  await loadItems()
}

async function removeItem(value: string, count: number) {
  await window.go.main.App.ListRemove(props.connId, props.keyName, value, count)
  await loadItems()
}

watch(() => props.keyName, loadItems)
onMounted(loadItems)
</script>
