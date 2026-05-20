<template>
  <div class="flex flex-col h-full">
    <!-- Toolbar -->
    <div class="flex items-center justify-between px-3 py-1.5 bg-bg-secondary border-b border-gray-200">
      <div class="flex items-center gap-2">
        <Type class="w-4 h-4 text-accent-green" />
        <span class="text-[13px] text-text-secondary">STRING</span>
        <span v-if="ttl !== undefined" class="text-[12px] text-text-muted">
          TTL: {{ formatTTL(ttl) }}
        </span>
      </div>
      <div class="flex items-center gap-2">
        <button
          class="px-3 py-1 text-[13px] rounded border border-gray-300 text-text-secondary hover:bg-gray-100 transition-colors"
          @click="formatJson"
        >
          Format JSON
        </button>
        <button class="px-2 py-1 rounded border border-gray-300 text-text-secondary hover:bg-gray-100 transition-colors" title="Refresh" @click="emit('refresh')">
          <RefreshCw class="w-3.5 h-3.5" />
        </button>
        <button class="px-2 py-1 rounded border border-red-200 text-accent-red/70 hover:bg-red-50 transition-colors" title="Delete Key" @click="emit('delete')">
          <Trash class="w-3.5 h-3.5" />
        </button>
        <button class="btn-primary" @click="saveValue">Save</button>
      </div>
    </div>

    <!-- Editor Area -->
    <div class="flex-1 flex flex-col min-h-0 p-3">
      <!-- TTL Editor -->
      <div class="flex items-center gap-2 mb-2">
        <label class="text-[12px] text-text-muted w-8">TTL:</label>
        <input
          v-model.number="ttlInput"
          class="input-field w-24 py-1"
          type="number"
          placeholder="-1"
          @blur="updateTTL"
        />
        <button
          class="text-[12px] text-accent-blue hover:text-accent-blue-hover transition-colors"
          @click="updateTTL"
        >
          Update
        </button>
      </div>

      <!-- Text Editor -->
      <textarea
        v-model="value"
        class="flex-1 w-full rounded-lg p-3 text-sm resize-none outline-none transition-all duration-200 font-mono"
        style="background: #f9fafb; border: 1px solid #d1d5db; color: #1f2937;"
        placeholder="String value..."
        spellcheck="false"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { Type, RefreshCw, Trash } from 'lucide-vue-next'

const props = defineProps<{ connId: string; keyName: string }>()
const emit = defineEmits(['refresh', 'delete'])

const value = ref('')
const ttl = ref<number | undefined>()
const ttlInput = ref(-1)

async function loadValue() {
  if (!props.connId || !props.keyName) return
  const val = await window.go.main.App.GetStringValue(props.connId, props.keyName)
  value.value = val || ''
  const ttlVal = await window.go.main.App.GetTTL(props.connId, props.keyName)
  ttl.value = ttlVal
  ttlInput.value = ttlVal
}

async function saveValue() {
  await window.go.main.App.SetStringValue(props.connId, props.keyName, value.value, ttlInput.value)
}

async function updateTTL() {
  ttl.value = ttlInput.value
}

function formatJson() {
  try {
    const obj = JSON.parse(value.value)
    value.value = JSON.stringify(obj, null, 2)
  } catch {
    // Not valid JSON, ignore
  }
}

function formatTTL(t: number) {
  if (t === -1) return 'persistent'
  if (t < 60) return `${t}s`
  if (t < 3600) return `${Math.floor(t / 60)}m`
  return `${Math.floor(t / 3600)}h`
}

watch(() => props.keyName, loadValue)
onMounted(loadValue)
</script>
