<template>
  <div class="flex flex-col h-full">
    <div class="flex items-center justify-between px-3 py-1.5 bg-bg-secondary border-b border-gray-200">
      <div class="flex items-center gap-2">
        <Sliders class="w-4 h-4 text-accent-blue" />
        <span class="text-[13px] text-text-secondary">ZSET ({{ entries.length }} entries)</span>
      </div>
      <button class="btn-primary" @click="showAdd = true">+ Add Entry</button>
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
          <tr v-if="entries.length === 0">
            <td colspan="3" class="text-center py-8 text-text-muted">No entries</td>
          </tr>
          <tr
            v-for="e in entries"
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
import { ref, watch, onMounted } from 'vue'
import { Sliders, Trash } from 'lucide-vue-next'
import { NModal } from 'naive-ui'

const props = defineProps<{ connId: string; keyName: string }>()

interface ZSetEntry {
  member: string
  score: number
}

const entries = ref<ZSetEntry[]>([])
const showAdd = ref(false)
const newMember = ref('')
const newScore = ref(0)

async function loadEntries() {
  if (!props.connId || !props.keyName) return
  const raw = await window.go.main.App.ZSetScan(props.connId, props.keyName, 0, 200)
  const result = JSON.parse(raw)
  entries.value = result.entries || []
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

watch(() => props.keyName, loadEntries)
onMounted(loadEntries)
</script>
