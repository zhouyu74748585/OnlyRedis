<template>
  <NModal v-model:show="visible" preset="card" title="Connection Configuration" style="width: 520px" :bordered="false">
    <n-tabs type="segment" animated>
      <n-tab-pane name="basic" tab="Basic">
        <div class="space-y-3 mt-2">
          <div>
            <label class="text-[13px] text-text-secondary mb-1.5 block">Name</label>
            <input v-model="form.name" class="input-field" placeholder="My Redis Server" />
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="text-[13px] text-text-secondary mb-1.5 block">Host</label>
              <input v-model="form.host" class="input-field" placeholder="127.0.0.1" />
            </div>
            <div>
              <label class="text-[13px] text-text-secondary mb-1.5 block">Port</label>
              <input v-model.number="form.port" class="input-field" placeholder="6379" type="number" />
            </div>
          </div>
          <div>
            <label class="text-[13px] text-text-secondary mb-1.5 block">Password</label>
            <input v-model="form.password" class="input-field" placeholder="Optional" type="password" />
          </div>
          <div>
            <label class="text-[13px] text-text-secondary mb-1.5 block">Database</label>
            <input v-model.number="form.db" class="input-field" placeholder="0" type="number" min="0" max="15" />
          </div>
        </div>
      </n-tab-pane>
      <n-tab-pane name="ssh" tab="SSH Tunnel">
        <div class="space-y-3 mt-2">
          <label class="flex items-center gap-2 cursor-pointer">
            <input v-model="form.sshEnabled" type="checkbox" class="accent-accent-blue" />
            <span class="text-[13px] text-text-secondary">Enable SSH Tunnel</span>
          </label>
          <template v-if="form.sshEnabled">
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="text-[13px] text-text-secondary mb-1.5 block">SSH Host</label>
                <input v-model="form.sshHost" class="input-field" placeholder="192.168.1.1" />
              </div>
              <div>
                <label class="text-[13px] text-text-secondary mb-1.5 block">SSH Port</label>
                <input v-model.number="form.sshPort" class="input-field" placeholder="22" type="number" />
              </div>
            </div>
            <div>
              <label class="text-[13px] text-text-secondary mb-1.5 block">SSH User</label>
              <input v-model="form.sshUser" class="input-field" placeholder="root" />
            </div>
            <div>
              <label class="text-[13px] text-text-secondary mb-1.5 block">SSH Password</label>
              <input v-model="form.sshPassword" class="input-field" placeholder="Password or leave empty for key" type="password" />
            </div>
            <div>
              <label class="text-[13px] text-text-secondary mb-1.5 block">SSH Key File</label>
              <input v-model="form.sshKeyFile" class="input-field" placeholder="~/.ssh/id_rsa" />
            </div>
          </template>
        </div>
      </n-tab-pane>
    </n-tabs>

    <template #footer>
      <div class="flex items-center justify-between">
        <button
          class="px-3 py-1.5 text-[13px] rounded-md border border-blue-200 text-accent-blue hover:bg-blue-50 transition-colors"
          @click="testConnection"
          :disabled="testing"
        >
          {{ testing ? 'Testing...' : 'Test Connection' }}
        </button>
        <div class="flex gap-2">
          <button
            class="px-4 py-1.5 text-[13px] rounded-md border border-gray-300 text-text-secondary hover:bg-gray-100 transition-colors"
            @click="emit('close')"
          >
            Cancel
          </button>
          <button
            class="btn-primary"
            @click="save"
          >
            {{ editData ? 'Update' : 'Create' }}
          </button>
        </div>
      </div>
    </template>
  </NModal>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { NModal, NTabs, NTabPane } from 'naive-ui'
import { useConnectionStore, type RedisConnection } from '@/stores/connection'

const props = defineProps<{ editData?: RedisConnection | null }>()
const emit = defineEmits<{ close: []; save: [cfg: RedisConnection] }>()

const store = useConnectionStore()
const visible = ref(true)
const testing = ref(false)

const form = reactive<RedisConnection>({
  id: '',
  name: '',
  host: '127.0.0.1',
  port: 6379,
  password: '',
  db: 0,
  sshEnabled: false,
  sshHost: '',
  sshPort: 22,
  sshUser: '',
  sshKeyFile: '',
  sshPassword: '',
  timeout: 5,
  retries: 3,
  status: 'disconnected',
})

onMounted(() => {
  if (props.editData) {
    Object.assign(form, props.editData)
  }
})

async function testConnection() {
  testing.value = true
  const result = await store.testConnection({ ...form })
  testing.value = false
  if (result.success) {
    alert('Connection successful!')
  } else {
    alert('Connection failed: ' + (result.error || 'Unknown error'))
  }
}

function save() {
  if (!form.name || !form.host) return
  emit('save', { ...form })
}
</script>
