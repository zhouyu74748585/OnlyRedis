<template>
  <div class="flex flex-col h-full">
    <div class="flex items-center justify-between px-3 py-1.5 bg-bg-secondary border-b border-gray-200">
      <div class="flex items-center gap-2">
        <Box class="w-4 h-4 text-accent-red" />
        <span class="text-[13px] text-text-secondary">
          SET ({{ totalMembers }} members<template v-if="searchQuery">, {{ activeDataset.length }} matched<template v-if="searching">...</template></template>)
        </span>
      </div>
      <div class="flex items-center gap-2">
        <button class="px-2 py-1 rounded border border-gray-300 text-text-secondary hover:bg-gray-100 transition-colors" title="Refresh" @click="emit('refresh')">
          <RefreshCw class="w-3.5 h-3.5" />
        </button>
        <button class="px-2 py-1 rounded border border-red-200 text-accent-red/70 hover:bg-red-50 transition-colors" title="Delete Key" @click="emit('delete')">
          <Trash class="w-3.5 h-3.5" />
        </button>
        <button class="btn-primary" @click="showAdd = true">+ Add Member</button>
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

    <!-- Truncated Warning -->
    <div v-if="truncated" class="px-3 py-1.5 bg-yellow-50 border-b border-yellow-200 text-[12px] text-yellow-700">
      ⚠ Showing first {{ pageSize * totalPages }} of {{ totalMembers }} members — data capped at 10,000 to avoid memory overflow
    </div>

    <div class="flex-1 overflow-auto">
      <div v-if="pagedMembers.length === 0" class="flex items-center justify-center h-full text-[13px] text-text-muted">
        Empty set
      </div>
      <div
        v-for="m in pagedMembers"
        :key="m"
        class="flex items-center gap-2 px-3 py-2 border-b border-gray-100 hover:bg-gray-50 transition-colors group"
      >
        <span class="flex-1 text-[13px] text-text-primary truncate font-mono">{{ m }}</span>
        <button
          class="p-1 rounded hover:bg-red-50 transition-colors opacity-0 group-hover:opacity-100"
          @click="removeMember(m)"
        >
          <Trash class="w-3.5 h-3.5 text-accent-red/60" />
        </button>
      </div>
    </div>

    <!-- Pagination (client-side) -->
    <div v-if="totalPages > 1" class="flex items-center justify-between px-3 py-1.5 bg-bg-secondary border-t border-gray-200">
      <span class="text-[12px] text-text-muted">Page {{ currentPage }} / {{ totalPages }}</span>
      <div class="flex gap-1">
        <button
          class="px-2 py-0.5 text-[12px] rounded border border-gray-300 text-text-secondary hover:bg-gray-100 transition-colors disabled:opacity-40"
          :disabled="currentPage <= 1"
          @click="currentPage--"
        >
          Prev
        </button>
        <button
          class="px-2 py-0.5 text-[12px] rounded border border-gray-300 text-text-secondary hover:bg-gray-100 transition-colors disabled:opacity-40"
          :disabled="currentPage >= totalPages"
          @click="currentPage++"
        >
          Next
        </button>
      </div>
    </div>

    <NModal v-if="showAdd" v-model:show="showAdd" preset="card" title="Add Set Member" style="width: 400px" :bordered="false">
      <div>
        <label class="text-[13px] text-text-secondary mb-1.5 block">Member</label>
        <input v-model="newMember" class="input-field" placeholder="member_value" @keyup.enter="addMember" />
      </div>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button class="px-4 py-1.5 text-[13px] rounded-md border border-gray-300 text-text-secondary hover:bg-gray-100 transition-colors" @click="showAdd = false">Cancel</button>
          <button class="btn-primary" @click="addMember">Add</button>
        </div>
      </template>
    </NModal>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue'
import { Box, Trash, Search, X, RefreshCw } from 'lucide-vue-next'

const props = defineProps<{ connId: string; keyName: string }>()
const emit = defineEmits(['refresh', 'delete'])
const allMembers = ref<string[]>([])
const searchMembers = ref<string[]>([])
const totalMembers = ref(0)
const searchQuery = ref('')
const searching = ref(false)
let searchTimer: ReturnType<typeof setTimeout> | null = null
const currentPage = ref(1)
const pageSize = 500
const truncated = ref(false)
const showAdd = ref(false)
const newMember = ref('')

// Active dataset: search results when query present, otherwise full dataset
const activeDataset = computed(() =>
  searchQuery.value ? searchMembers.value : allMembers.value
)

const totalPages = computed(() => Math.max(1, Math.ceil(activeDataset.value.length / pageSize)))
const pagedMembers = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  return activeDataset.value.slice(start, start + pageSize)
})

// Backend search with debounce (300ms)
async function doSetSearch() {
  const q = searchQuery.value.trim()
  if (!q || !props.connId || !props.keyName) {
    searchMembers.value = []
    return
  }
  searching.value = true
  const raw = await window.go.main.App.SetSearch(props.connId, props.keyName, q)
  const result = JSON.parse(raw)
  searchMembers.value = result.members || []
  searching.value = false
}

async function loadMembers() {
  if (!props.connId || !props.keyName) return
  // count=-1 tells backend to fetch ALL members in a SCAN loop
  const raw = await window.go.main.App.SetMembers(props.connId, props.keyName, 0, -1)
  const result = JSON.parse(raw)
  allMembers.value = result.members || []
  totalMembers.value = result.total || 0
  truncated.value = result.truncated === true
}

async function addMember() {
  if (!newMember.value) return
  await window.go.main.App.SetAdd(props.connId, props.keyName, newMember.value)
  showAdd.value = false
  newMember.value = ''
  await loadMembers()
}

async function removeMember(m: string) {
  await window.go.main.App.SetRemove(props.connId, props.keyName, m)
  await loadMembers()
}

// Debounce search input → backend call
watch(searchQuery, () => {
  currentPage.value = 1
  if (searchTimer) clearTimeout(searchTimer)
  if (!searchQuery.value.trim()) {
    searchMembers.value = []
    return
  }
  searching.value = true
  searchTimer = setTimeout(doSetSearch, 300)
})

watch(() => props.keyName, () => {
  searchQuery.value = ''
  currentPage.value = 1
  loadMembers()
})
onMounted(loadMembers)
</script>
