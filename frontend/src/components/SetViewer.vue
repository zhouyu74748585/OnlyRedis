<template>
  <div class="flex flex-col h-full">
    <div class="flex items-center justify-between px-3 py-1.5 bg-bg-secondary border-b border-gray-200">
      <div class="flex items-center gap-2">
        <Box class="w-4 h-4 text-accent-red" />
        <span class="text-[13px] text-text-secondary">SET ({{ members.length }} members)</span>
      </div>
      <button class="btn-primary" @click="showAdd = true">+ Add Member</button>
    </div>

    <div class="flex-1 overflow-auto">
      <div v-if="members.length === 0" class="flex items-center justify-center h-full text-[13px] text-text-muted">
        Empty set
      </div>
      <div
        v-for="m in members"
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
import { ref, watch, onMounted } from 'vue'
import { Box, Trash } from 'lucide-vue-next'

const props = defineProps<{ connId: string; keyName: string }>()
const members = ref<string[]>([])
const showAdd = ref(false)
const newMember = ref('')

async function loadMembers() {
  if (!props.connId || !props.keyName) return
  const raw = await window.go.main.App.SetMembers(props.connId, props.keyName, 0, 500)
  const result = JSON.parse(raw)
  members.value = result.members || []
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

watch(() => props.keyName, loadMembers)
onMounted(loadMembers)
</script>
