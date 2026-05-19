<template>
  <div class="flex flex-col h-full">
    <!-- Toolbar -->
    <div class="flex items-center justify-between px-3 py-1.5 bg-bg-secondary border-b border-gray-200">
      <div class="flex items-center gap-2">
        <Hash class="w-4 h-4 text-accent-orange" />
        <span class="text-[13px] text-text-secondary">HASH ({{ totalFields }} fields)</span>
      </div>
      <button class="btn-primary" @click="showAdd = true">+ Add Field</button>
    </div>

    <!-- Table -->
    <div class="flex-1 overflow-auto">
      <table class="w-full text-[13px]" style="table-layout: fixed">
        <thead class="sticky top-0 bg-bg-secondary z-10">
          <tr class="border-b border-gray-200">
            <th class="text-left px-3 py-2 text-text-muted font-medium relative" :style="{ width: fieldWidth + 'px' }">
              Field
              <!-- Resize handle -->
              <div
                class="absolute right-0 top-0 bottom-0 w-[5px] cursor-col-resize hover:bg-accent-blue/20 transition-colors flex items-center justify-center group"
                @mousedown.stop="startFieldResize"
              >
                <div class="h-4 w-0.5 rounded bg-gray-300 group-hover:bg-accent-blue/50 transition-colors" />
              </div>
            </th>
            <th class="text-left px-3 py-2 text-text-muted font-medium">Value</th>
            <th v-if="supportFieldTTL" class="text-right px-3 py-2 text-text-muted font-medium" style="width: 72px">TTL</th>
            <th class="text-right px-3 py-2 text-text-muted font-medium" style="width: 60px">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="fields.length === 0">
            <td :colspan="supportFieldTTL ? 4 : 3" class="text-center py-8 text-text-muted">No fields</td>
          </tr>
          <tr
            v-for="(item, idx) in fields"
            :key="item.field"
            class="border-b border-gray-100 hover:bg-gray-50 transition-colors"
          >
            <td class="px-3 py-1.5 text-text-primary font-mono truncate" :style="{ maxWidth: fieldWidth + 'px' }">
              <span :title="item.field">{{ item.field }}</span>
            </td>
            <td class="px-3 py-1.5">
              <span
                class="block w-full text-text-secondary cursor-pointer truncate hover:text-accent-blue hover:bg-blue-50 rounded px-1 py-0.5 transition-colors"
                :title="'Click to edit: ' + item.value"
                @click="openEditDialog(item)"
              >{{ item.value }}</span>
            </td>
            <td v-if="supportFieldTTL" class="px-3 py-1.5 text-right text-[12px] font-mono" :class="item.ttl >= 0 ? 'text-accent-green' : 'text-text-muted'">
              {{ formatTTL(item.ttl) }}
            </td>
            <td class="px-3 py-1.5 text-right">
              <button
                class="p-1 rounded hover:bg-red-50 transition-colors"
                @click="deleteField(item.field)"
              >
                <Trash class="w-3 h-3 text-accent-red/60" />
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Pagination -->
    <div v-if="totalFields > 100" class="flex items-center justify-between px-3 py-1.5 bg-bg-secondary border-t border-gray-200">
      <span class="text-[12px] text-text-muted">Page {{ currentPage }}</span>
      <div class="flex gap-1">
        <button
          class="px-2 py-0.5 text-[12px] rounded border border-gray-300 text-text-secondary hover:bg-gray-100 transition-colors"
          :disabled="currentPage <= 1"
          @click="prevPage"
        >
          Prev
        </button>
        <button
          class="px-2 py-0.5 text-[12px] rounded border border-gray-300 text-text-secondary hover:bg-gray-100 transition-colors"
          :disabled="fields.length < 100"
          @click="nextPage"
        >
          Next
        </button>
      </div>
    </div>

    <!-- Add Field Dialog -->
    <NModal v-if="showAdd" v-model:show="showAdd" preset="card" title="Add Hash Field" style="width: 400px" :bordered="false">
      <div class="space-y-3">
        <div>
          <label class="text-[13px] text-text-secondary mb-1.5 block">Field</label>
          <input v-model="newField" class="input-field" placeholder="field_name" />
        </div>
        <div>
          <label class="text-[13px] text-text-secondary mb-1.5 block">Value</label>
          <input v-model="newValue" class="input-field" placeholder="value" />
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button class="px-4 py-1.5 text-[13px] rounded-md border border-gray-300 text-text-secondary hover:bg-gray-100 transition-colors" @click="showAdd = false">Cancel</button>
          <button class="btn-primary" @click="addField">Add</button>
        </div>
      </template>
    </NModal>

    <!-- Edit Value Dialog -->
    <NModal v-model:show="showEdit" preset="card" :title="'Edit Field: ' + editingField" style="width: 720px" :bordered="false">
      <div class="flex flex-col" style="min-height: 360px">
        <div class="flex items-center justify-between mb-2">
          <span class="text-[12px] text-text-muted">Value for field <span class="font-mono text-text-primary">{{ editingField }}</span></span>
          <button
            class="px-3 py-1 text-[12px] rounded border border-gray-300 text-text-secondary hover:bg-gray-100 transition-colors"
            @click="formatEditJson"
          >
            Format JSON
          </button>
        </div>
        <textarea
          v-model="editValue"
          class="flex-1 w-full rounded-lg p-3 text-sm resize-none outline-none transition-all duration-200 font-mono"
          style="min-height: 320px; background: #f9fafb; border: 1px solid #d1d5db; color: #1f2937;"
          placeholder="Edit value..."
          spellcheck="false"
        />
      </div>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button
            class="px-4 py-1.5 text-[13px] rounded-md border border-gray-300 text-text-secondary hover:bg-gray-100 transition-colors"
            @click="showEdit = false"
          >Cancel</button>
          <button class="btn-primary" @click="saveEditField">Save</button>
        </div>
      </template>
    </NModal>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue'
import { Hash, Trash } from 'lucide-vue-next'
import { useConnectionStore } from '../stores/connection'

const props = defineProps<{ connId: string; keyName: string }>()

const connectionStore = useConnectionStore()

const fields = ref<{ field: string; value: string; ttl: number }[]>([])
const totalFields = ref(0)
const currentPage = ref(1)
const showAdd = ref(false)
const newField = ref('')
const newValue = ref('')

// Edit value dialog state
const showEdit = ref(false)
const editingField = ref('')
const editValue = ref('')

const fieldWidth = ref(200)

let nextCursor = 0

// Check if Redis version >= 7.4.0 (HTTL support)
const version = computed(() => connectionStore.redisVersions[props.connId] || '')
const supportFieldTTL = computed(() => {
  const v = version.value
  if (!v) return false
  const parts = v.split('.').map(Number)
  if (parts.length < 2) return false
  // Compare major.minor: >= 7.4
  if (parts[0] > 7) return true
  if (parts[0] === 7 && parts[1] >= 4) return true
  return false
})

// === Field column resize ===
function startFieldResize(e: MouseEvent) {
  e.preventDefault()
  e.stopPropagation()
  const startX = e.clientX
  const startWidth = fieldWidth.value

  function onMove(ev: MouseEvent) {
    const delta = ev.clientX - startX
    fieldWidth.value = Math.max(80, Math.min(500, startWidth + delta))
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

function formatTTL(ttl: number): string {
  if (ttl === -2) return '-'
  if (ttl === -1) return '∞'
  return `${ttl}s`
}

async function loadFields() {
  if (!props.connId || !props.keyName) return
  const withTTL = supportFieldTTL.value
  const raw = await window.go.main.App.HashScan(props.connId, props.keyName, nextCursor, 100, withTTL)
  const result = JSON.parse(raw)
  fields.value = result.fields || []
  totalFields.value = result.total || 0
  nextCursor = result.cursor || 0
}

async function updateField(field: string, value: string) {
  await window.go.main.App.HashSet(props.connId, props.keyName, field, value)
}

// === Edit Value Dialog ===
function openEditDialog(item: { field: string; value: string }) {
  editingField.value = item.field
  editValue.value = item.value
  showEdit.value = true
}

async function saveEditField() {
  await window.go.main.App.HashSet(props.connId, props.keyName, editingField.value, editValue.value)
  showEdit.value = false
  // Refresh the fields to show updated value
  await loadFields()
}

function formatEditJson() {
  try {
    const obj = JSON.parse(editValue.value)
    editValue.value = JSON.stringify(obj, null, 2)
  } catch {
    // Not valid JSON, ignore
  }
}

async function deleteField(field: string) {
  await window.go.main.App.HashDel(props.connId, props.keyName, field)
  await loadFields()
}

async function addField() {
  if (!newField.value) return
  await window.go.main.App.HashSet(props.connId, props.keyName, newField.value, newValue.value)
  showAdd.value = false
  newField.value = ''
  newValue.value = ''
  await loadFields()
}

function prevPage() {
  if (currentPage.value > 1) {
    currentPage.value--
    nextCursor = 0
    loadFields()
  }
}

function nextPage() {
  currentPage.value++
  loadFields()
}

watch(() => props.keyName, () => {
  nextCursor = 0
  currentPage.value = 1
  fieldWidth.value = 200
  loadFields()
})
onMounted(loadFields)
</script>
