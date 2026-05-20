<template>
  <div class="flex flex-col h-full">
    <div class="flex items-center justify-between px-3 py-1.5 bg-bg-secondary border-b border-gray-200">
      <div class="flex items-center gap-2">
        <Sliders class="w-4 h-4 text-accent-blue" />
        <span class="text-[13px] text-text-secondary">
          ZSET ({{ totalEntries }} entries<template v-if="searchQuery">, {{ activeEntries.length }} matched<template v-if="searching">...</template></template>)
        </span>
      </div>
      <div class="flex items-center gap-2">
        <button class="px-2 py-1 rounded border border-gray-300 text-text-secondary hover:bg-gray-100 transition-colors" title="Refresh" @click="emit('refresh')">
          <RefreshCw class="w-3.5 h-3.5" />
        </button>
        <button class="px-2 py-1 rounded border border-red-200 text-accent-red/70 hover:bg-red-50 transition-colors" title="Delete Key" @click="emit('delete')">
          <Trash class="w-3.5 h-3.5" />
        </button>
        <button class="btn-primary" @click="showAdd = true">+ Add Entry</button>
      </div>
    </div>

    <!-- Search Bar -->
    <div class="flex items-center gap-1.5 px-3 py-1 bg-bg-secondary border-b border-gray-200">
      <Search class="w-3.5 h-3.5 text-text-muted flex-shrink-0" />
      <input
        v-model="searchQuery"
        class="flex-1 text-[12px] bg-transparent outline-none text-text-primary placeholder-text-muted"
        placeholder="Search member..."
      />
      <button v-if="searchQuery" class="p-0.5 rounded hover:bg-gray-200 transition-colors" @click="searchQuery = ''">
        <X class="w-3 h-3 text-text-muted" />
      </button>
    </div>

    <div class="flex-1 overflow-auto">
      <table class="w-full text-[13px]">
        <thead class="sticky top-0 bg-bg-secondary">
          <tr class="border-b border-gray-200">
            <th class="text-left px-3 py-2 text-text-muted font-medium w-[80px]">Score</th>
            <th class="text-left px-3 py-2 text-text-muted font-medium">Member</th>
            <th class="text-right px-3 py-2 text-text-muted font-medium w-[60px]">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="activeEntries.length === 0">
            <td colspan="3" class="text-center py-8 text-text-muted">No entries</td>
          </tr>
          <tr
            v-for="e in activeEntries"
            :key="e.member"
            class="border-b border-gray-100 hover:bg-gray-50 transition-colors group"
          >
            <td class="px-3 py-1.5 text-accent-blue font-mono">{{ e.score }}</td>
            <td class="px-3 py-1.5 text-text-primary font-mono">{{ e.member }}</td>
            <td class="px-3 py-1.5 text-right">
              <button
                class="p-1 rounded hover:bg-red-50 transition-colors opacity-0 group-hover:opacity-100"
                @click="removeEntry(e.member)"
              >
                <Trash class="w-3.5 h-3.5 text-accent-red/60" />
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1" class="flex items-center justify-between px-3 py-1.5 bg-bg-secondary border-t border-gray-200">
      <span class="text-[12px] text-text-muted">Page {{ currentPage }} / {{ totalPages }}</span>
      <div class="flex gap-1">
        <button
          class="px-2 py-0.5 text-[12px] rounded border border-gray-300 text-text-secondary hover:bg-gray-100 transition-colors disabled:opacity-40"
          :disabled="currentPage <= 1"
          @click="goPage(currentPage - 1)"
        >
          Prev
        </button>
        <button
          class="px-2 py-0.5 text-[12px] rounded border border-gray-300 text-text-secondary hover:bg-gray-100 transition-colors disabled:opacity-40"
          :disabled="currentPage >= totalPages"
          @click="goPage(currentPage + 1)"
        >
          Next
        </button>
      </div>
    </div>

    <NModal v-if="showAdd" v-model:show="showAdd" preset="card" title="Add ZSet Entry" style="width: 400px" :bordered="false">
      <div class="space-y-3">
        <div>
          <label class="text-[13px] text-text-secondary mb-1.5 block">Member</label>
          <input v-model="newMember" class="input-field" placeholder="member_value" />
        </div>
        <div>
          <label class="text-[13px] text-text-secondary mb-1.5 block">Score</label>
          <input v-model.number="newScore" class="input-field" placeholder="0.0" type="number" step="any" />
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button class="px-4 py-1.5 text-[13px] rounded-md border border-gray-300 text-text-secondary hover:bg-gray-100 transition-colors" @click="showAdd = false">Cancel</button>
          <button class="btn-primary" @click="addEntry">Add</button>
        </div>
      </template>
    </NModal>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue'
import { Sliders, Trash, Search, X, RefreshCw } from 'lucide-vue-next'

const props = defineProps<{ connId: string; keyName: string }>()
const emit = defineEmits(['refresh', 'delete'])

interface ZSetEntry {
  member: string
  score: number
}

const entries = ref<ZSetEntry[]>([])
const searchEntries = ref<ZSetEntry[]>([])
const totalEntries = ref(0)
const currentPage = ref(1)
const pageSize = 200
const searchQuery = ref('')
const searching = ref(false)
let searchTimer: ReturnType<typeof setTimeout> | null = null
const showAdd = ref(false)
const newMember = ref('')
const newScore = ref(0)

// Active dataset: search results or current page
const activeEntries = computed(() =>
  searchQuery.value ? searchEntries.value : entries.value
)

const totalPages = computed(() => {
  if (searchQuery.value) return 1 // show all search results, no pagination
  return Math.max(1, Math.ceil(totalEntries.value / pageSize))
})

// Backend search via ZSCAN MATCH
async function doZSetSearch() {
  const q = searchQuery.value.trim()
  if (!q || !props.connId || !props.keyName) {
    searchEntries.value = []
    return
  }
  searching.value = true
  const raw = await window.go.main.App.ZSetSearch(props.connId, props.keyName, q)
  const result = JSON.parse(raw)
  searchEntries.value = result.entries || []
  searching.value = false
}

// True server-side pagination via ZRANGE + ZCARD
async function loadEntries() {
  if (!props.connId || !props.keyName) return
  const start = (currentPage.value - 1) * pageSize
  const stop = currentPage.value * pageSize - 1
  const raw = await window.go.main.App.ZSetRange(props.connId, props.keyName, start, stop)
  const result = JSON.parse(raw)
  entries.value = result.entries || []
  totalEntries.value = result.total || 0
}

function goPage(page: number) {
  currentPage.value = page
  loadEntries()
}

async function addEntry() {
  if (!newMember.value) return
  await window.go.main.App.ZSetAdd(props.connId, props.keyName, newMember.value, newScore.value)
  showAdd.value = false
  newMember.value = ''
  newScore.value = 0
  await loadEntries()
}

async function removeEntry(member: string) {
  await window.go.main.App.ZSetRemove(props.connId, props.keyName, member)
  await loadEntries()
}

// Debounce search input → backend call
watch(searchQuery, () => {
  currentPage.value = 1
  if (searchTimer) clearTimeout(searchTimer)
  if (!searchQuery.value.trim()) {
    searchEntries.value = []
    return
  }
  searching.value = true
  searchTimer = setTimeout(doZSetSearch, 300)
})

watch(() => props.keyName, () => {
  searchQuery.value = ''
  currentPage.value = 1
  loadEntries()
})
onMounted(loadEntries)
</script>
