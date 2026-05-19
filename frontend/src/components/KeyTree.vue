<template>
  <div class="flex flex-col flex-1 min-h-0">
    <!-- Header -->
    <div class="flex items-center justify-between px-3 py-2">
      <div class="flex items-center gap-2">
        <KeyIcon class="w-4 h-4 text-accent-blue" />
        <span class="text-[13px] font-semibold uppercase text-text-secondary tracking-wider">Keys</span>
      </div>
      <button
        class="w-7 h-7 flex items-center justify-center rounded-md hover:bg-gray-100 transition-colors"
        @click="onRefresh"
        :disabled="!props.connId"
      >
        <RefreshCw class="w-4 h-4 text-text-secondary" :class="{ 'animate-spin': store.rootLoading }" />
      </button>
    </div>

    <!-- DB Selector -->
    <div v-if="props.connId" class="px-2 pb-2">
      <div class="flex items-center gap-2">
        <span class="text-[12px] text-text-muted flex-shrink-0">DB</span>
        <select
          :value="store.currentDb"
          class="flex-1 text-[13px] px-2 py-1 rounded-md border border-gray-300 bg-white text-text-primary outline-none focus:border-accent-blue/40 transition-colors"
          @change="onDbChange(Number(($event.target as HTMLSelectElement).value))"
        >
          <option v-for="db in 16" :key="db - 1" :value="db - 1">Database {{ db - 1 }}</option>
        </select>
      </div>
    </div>

    <!-- Search -->
    <div class="px-2 pb-1">
      <div class="relative">
        <Search class="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-text-muted" />
        <input
          v-model="searchText"
          class="input-field pl-8 text-sm"
          placeholder="Search keys or type command (GET/HGETALL/KEYS...)"
          @keyup.enter="onSearch"
          @focus="showSearchHint = true"
          @blur="hideSearchHint"
        />
      </div>
      <div v-if="showSearchHint" class="mt-1 px-1 text-[11px] text-text-muted leading-relaxed">
        <span class="font-medium">Key:</span> <code class="text-accent-blue">user:*</code> wildcard search
        &nbsp;|&nbsp;
        <span class="font-medium">Cmd:</span> <code class="text-accent-blue">GET mykey</code>,
        <code class="text-accent-blue">HGETALL myhash</code>
        <br>
        <span class="font-medium">Pattern:</span> <code class="text-accent-blue">*</code> any
        &nbsp;<code class="text-accent-blue">?</code> one
        &nbsp;<code class="text-accent-blue">[abc]</code> set
      </div>
    </div>
    <!-- Add Key Button -->
    <div v-if="props.connId" class="px-2 pb-2">
      <button
        class="w-full flex items-center justify-center gap-1.5 py-2 rounded-md border border-dashed border-gray-300 text-[13px] text-text-muted hover:border-accent-blue/40 hover:text-accent-blue transition-colors"
        @click="showAddDialog = true"
      >
        <Plus class="w-3.5 h-3.5" />
        New Key
      </button>
    </div>

    <!-- Key Tree / Search Results -->
    <div class="flex-1 overflow-y-auto px-2 pb-2" @click="closeContextMenu">
      <!-- Empty states -->
      <div v-if="!props.connId" class="flex items-center justify-center h-full text-[13px] text-text-muted">
        Connect to a Redis server first
      </div>
      <div v-else-if="isTreeMode && store.rootLoading && store.rootNodes.length === 0" class="flex items-center justify-center h-full">
        <RefreshCw class="w-5 h-5 text-text-muted animate-spin" />
      </div>
      <div v-else-if="isTreeMode && store.rootNodes.length === 0" class="flex items-center justify-center h-full text-[13px] text-text-muted">
        No keys found
      </div>

      <!-- Tree mode -->
      <div v-else-if="isTreeMode">
        <TreeNode
          v-for="node in store.rootNodes"
          :key="node.fullKey"
          :node="node"
          :depth="0"
          :conn-id="props.connId"
          :selected-key="store.selectedKey"
          @select="onNodeSelect"
          @contextmenu="onNodeContextMenu"
        />
        <div
          v-if="store.rootHasMore"
          class="flex items-center gap-1 py-1 cursor-pointer text-text-muted hover:text-accent-blue transition-colors select-none px-1"
          @click="store.loadRoot(props.connId, true)"
        >
          <span class="text-[12px]">Load more...</span>
        </div>
      </div>

      <!-- Search mode -->
      <div v-else>
        <div v-if="searching" class="flex items-center justify-center h-full">
          <RefreshCw class="w-5 h-5 text-text-muted animate-spin" />
        </div>
        <div v-else-if="store.searchKeys.length === 0" class="flex items-center justify-center h-full text-[13px] text-text-muted">
          No keys matching "{{ searchText }}"
        </div>
        <div v-else class="space-y-0.5">
          <div
            v-for="kn in store.searchKeys"
            :key="kn.key"
            class="flex items-center gap-1.5 px-2 py-1 rounded cursor-pointer transition-colors group"
            :class="store.selectedKey === kn.key ? 'bg-blue-50 text-accent-blue' : 'text-text-secondary hover:bg-gray-100'"
            @click="store.selectKey(kn.key, kn.type)"
            @contextmenu.prevent.stop="onFlatContextMenu($event, kn.key)"
          >
            <component :is="typeIcon(kn.type)" class="w-3.5 h-3.5 flex-shrink-0" :class="typeColor(kn.type)" />
            <span class="text-[13px] truncate flex-1" :title="kn.key">{{ formatKeyName(kn.key) }}</span>
            <span class="text-[11px] text-text-muted flex-shrink-0">{{ kn.type }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Context Menu -->
    <Teleport to="body">
      <div
        v-if="contextMenu.visible"
        class="fixed z-50 bg-white rounded-lg shadow-lg border border-gray-200 py-1 min-w-[160px]"
        :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }"
        @click.stop
      >
        <button
          class="w-full flex items-center gap-2 px-3 py-1.5 text-[13px] text-text-secondary hover:bg-gray-100 transition-colors text-left"
          @click="copyKey"
        >
          <Copy class="w-3.5 h-3.5" />
          Copy Key
        </button>
        <button
          class="w-full flex items-center gap-2 px-3 py-1.5 text-[13px] text-accent-red hover:bg-red-50 transition-colors text-left"
          @click="deleteContextKey"
        >
          <Trash2 class="w-3.5 h-3.5" />
          Delete Key
        </button>
      </div>
    </Teleport>

    <!-- Add Key Dialog -->
    <NModal v-if="showAddDialog" v-model:show="showAddDialog" preset="card" title="New Key" style="width: 420px" :bordered="false">
      <div class="space-y-3">
        <div>
          <label class="text-[13px] text-text-secondary mb-1.5 block">Key Name</label>
          <input v-model="newKeyName" class="input-field" placeholder="my:key:name" />
        </div>
        <div>
          <label class="text-[13px] text-text-secondary mb-1.5 block">Data Type</label>
          <div class="flex gap-2">
            <button
              v-for="t in ['string','hash','list','set','zset']"
              :key="t"
              class="px-3 py-1.5 rounded text-[13px] transition-colors"
              :class="newKeyType === t ? 'bg-blue-50 text-accent-blue border border-blue-200' : 'border border-gray-300 text-text-secondary hover:bg-gray-100'"
              @click="newKeyType = t"
            >
              {{ t.toUpperCase() }}
            </button>
          </div>
        </div>
        <div>
          <label class="text-[13px] text-text-secondary mb-1.5 block">Initial Value</label>
          <textarea
            v-if="newKeyType === 'string'"
            v-model="newKeyValue"
            class="input-field resize-none h-20 font-mono"
            placeholder="Enter string value..."
          />
          <textarea
            v-else-if="newKeyType === 'hash'"
            v-model="newKeyValue"
            class="input-field resize-none h-20 font-mono"
            :placeholder="'field1=value1\nfield2=value2'"
          />
          <textarea
            v-else
            v-model="newKeyValue"
            class="input-field resize-none h-20 font-mono"
            :placeholder="placeholders[newKeyType] || ''"
          />
        </div>
        <div>
          <label class="text-[13px] text-text-secondary mb-1.5 block">TTL (seconds, -1 for persistent)</label>
          <input v-model.number="newKeyTTL" class="input-field" type="number" placeholder="-1" />
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button class="px-4 py-1.5 text-[13px] rounded-md border border-gray-300 text-text-secondary hover:bg-gray-100 transition-colors" @click="showAddDialog = false">Cancel</button>
          <button class="btn-primary text-[13px]" @click="createKey">Create</button>
        </div>
      </template>
    </NModal>

    <!-- Delete Confirm Dialog -->
    <NModal v-if="showDeleteConfirm" v-model:show="showDeleteConfirm" preset="card" title="Delete Key" style="width: 400px" :bordered="false">
      <p class="text-[13px] text-text-secondary">
        Are you sure you want to delete <span class="text-accent-red font-mono">{{ contextMenu.key }}</span>?
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
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { Key as KeyIcon, RefreshCw, Search, Plus, Copy, Trash2, Type, Hash, List, Box, Sliders } from 'lucide-vue-next'
import { NModal } from 'naive-ui'
import { useKeyStore } from '@/stores/key'
import TreeNode from '@/components/TreeNode.vue'

const props = defineProps<{ connId: string }>()
const store = useKeyStore()

const searchText = ref('')
const searching = ref(false)
const showSearchHint = ref(false)
const showAddDialog = ref(false)
const showDeleteConfirm = ref(false)
const newKeyName = ref('')
const newKeyType = ref('string')
const newKeyValue = ref('')
const newKeyTTL = ref(-1)
let searchHintTimer: ReturnType<typeof setTimeout> | null = null

const isTreeMode = computed(() => !searchText.value || searchText.value === '*')

const contextMenu = ref({
  visible: false,
  x: 0,
  y: 0,
  key: '',
})

/** Redis commands that extract a key name and auto-open it */
const KEY_EXTRACT_COMMANDS = ['GET', 'TYPE', 'TTL', 'EXISTS', 'HGETALL', 'LRANGE', 'SMEMBERS', 'ZRANGE']

const placeholders: Record<string, string> = {
  'string': '',
  'hash': 'field1=value1\nfield2=value2',
  'list': 'item1\nitem2\nitem3',
  'set': 'member1\nmember2',
  'zset': 'member1 1.0\nmember2 2.0',
}

watch(() => props.connId, (newId) => {
  if (newId) {
    searchText.value = ''
    store.currentDb = 0
    store.rootNodes = []
    store.loadRoot(newId)
  }
})

async function onDbChange(db: number) {
  await (window.go.main.App as any).SelectDB(props.connId, db)
  store.currentDb = db
  store.rootNodes = []
  store.searchKeys = []
  searchText.value = ''
  await store.loadRoot(props.connId)
}

function onRefresh() {
  if (isTreeMode.value) {
    store.rootNodes = []
    store.loadRoot(props.connId)
  } else {
    doSearch()
  }
}

/** Parse search input into tokens (respecting quotes) */
function tokenize(text: string): string[] {
  const tokens: string[] = []
  const regex = /"([^"]*)"|'([^']*)'|(\S+)/g
  let m: RegExpExecArray | null
  while ((m = regex.exec(text)) !== null) {
    tokens.push(m[1] || m[2] || m[3])
  }
  return tokens
}

/** Check if text has any Redis glob wildcards */
function hasWildcards(text: string): boolean {
  // : is NOT a wildcard in Redis glob, only * ? [ ] are
  return /[*?\[\]]/.test(text)
}

async function onSearch() {
  const text = searchText.value.trim()
  if (!text || text === '*') {
    searchText.value = ''
    store.rootNodes = []
    store.searchKeys = []
    await store.loadRoot(props.connId)
    return
  }
  await doSearch()
}

async function doSearch() {
  const text = searchText.value.trim()
  searching.value = true

  try {
    const tokens = tokenize(text)
    const firstToken = tokens[0]?.toUpperCase() || ''

    // 1. Redis command with key extraction: GET / HGETALL / LRANGE / SMEMBERS / ZRANGE / TYPE / TTL
    if (KEY_EXTRACT_COMMANDS.includes(firstToken) && tokens.length >= 2) {
      const keyName = tokens[1]
      const keyType = await window.go.main.App.GetKeyType(props.connId, keyName)
      if (!keyType || keyType === 'none') {
        store.searchKeys = []
        return
      }
      // Pre-cache and navigate
      store.keyTypes[keyName] = keyType
      store.selectKey(keyName, keyType)
      // Also show in search results
      store.searchKeys = [{ key: keyName, type: keyType, ttl: -2 }]
      return
    }

    // 2. KEYS command: KEYS <pattern>
    if (firstToken === 'KEYS' && tokens.length >= 2) {
      await store.scanKeys(props.connId, tokens[1])
      return
    }

    // 3. SCAN command: SCAN <cursor> [MATCH <pattern>] [COUNT <count>]
    if (firstToken === 'SCAN') {
      const cursor = parseInt(tokens[1]) || 0
      let match = '*'
      let count = 200
      for (let i = 2; i < tokens.length; i++) {
        if (tokens[i].toUpperCase() === 'MATCH' && i + 1 < tokens.length) match = tokens[++i]
        if (tokens[i].toUpperCase() === 'COUNT' && i + 1 < tokens.length) count = parseInt(tokens[++i]) || 200
      }
      await store.scanKeys(props.connId, match)
      return
    }

    // 4. Exact key name (no wildcards): check existence directly
    if (!hasWildcards(text)) {
      const keyType = await window.go.main.App.GetKeyType(props.connId, text)
      if (keyType && keyType !== 'none') {
        store.searchKeys = [{ key: text, type: keyType, ttl: -2 }]
        return
      }
      // Not found, try as pattern anyway
      store.searchKeys = []
      return
    }

    // 5. Wildcard pattern: SCAN with multiple rounds
    await store.scanKeys(props.connId, text)
  } finally {
    searching.value = false
  }
}

function onNodeSelect(key: string, keyType: string) {
  store.selectKey(key, keyType)
}

function onNodeContextMenu(payload: { event: MouseEvent; key: string }) {
  contextMenu.value = {
    visible: true,
    x: payload.event.clientX,
    y: payload.event.clientY,
    key: payload.key,
  }
}

function onFlatContextMenu(e: MouseEvent, key: string) {
  contextMenu.value = {
    visible: true,
    x: e.clientX,
    y: e.clientY,
    key,
  }
}

function closeContextMenu() {
  contextMenu.value.visible = false
}

function hideSearchHint() {
  searchHintTimer = setTimeout(() => { showSearchHint.value = false }, 200)
}

function copyKey() {
  navigator.clipboard.writeText(contextMenu.value.key)
  contextMenu.value.visible = false
}

function deleteContextKey() {
  contextMenu.value.visible = false
  showDeleteConfirm.value = true
}

async function confirmDeleteKey() {
  await window.go.main.App.DeleteKey(props.connId, contextMenu.value.key)
  showDeleteConfirm.value = false
  if (store.openedTabs.includes(contextMenu.value.key)) {
    store.closeTab(contextMenu.value.key)
  }
  if (isTreeMode.value) {
    store.rootNodes = []
    await store.loadRoot(props.connId)
  } else {
    await doSearch()
  }
}

async function createKey() {
  if (!newKeyName.value) return
  const app = window.go.main.App
  const connId = props.connId
  const key = newKeyName.value
  const ttl = newKeyTTL.value

  if (newKeyType.value === 'string') {
    await app.SetStringValue(connId, key, newKeyValue.value, ttl)
  } else if (newKeyType.value === 'hash') {
    // Parse field=value pairs (one per line)
    const lines = newKeyValue.value.split('\n').filter(l => l.trim())
    if (lines.length > 0) {
      // First set an empty string to create the key, then set hash fields
      await app.SetStringValue(connId, key, '', -1)
      await app.DeleteKey(connId, key)
      for (const line of lines) {
        const eq = line.indexOf('=')
        if (eq > 0) {
          const field = line.substring(0, eq).trim()
          const value = line.substring(eq + 1).trim()
          if (field) await app.HashSet(connId, key, field, value)
        }
      }
      if (ttl > 0) await (app as any).SetTTL(connId, key, ttl)
    } else {
      await app.SetStringValue(connId, key, '', ttl)
    }
  } else if (newKeyType.value === 'list') {
    const items = newKeyValue.value.split('\n').filter(l => l.trim())
    if (items.length > 0) {
      for (let i = items.length - 1; i >= 0; i--) {
        await app.ListPush(connId, key, items[i].trim(), true)
      }
      if (ttl > 0) await (app as any).SetTTL(connId, key, ttl)
    } else {
      await app.SetStringValue(connId, key, '', ttl)
    }
  } else if (newKeyType.value === 'set') {
    const members = newKeyValue.value.split('\n').filter(l => l.trim())
    if (members.length > 0) {
      for (const m of members) {
        await app.SetAdd(connId, key, m.trim())
      }
      if (ttl > 0) await (app as any).SetTTL(connId, key, ttl)
    } else {
      await app.SetStringValue(connId, key, '', ttl)
    }
  } else if (newKeyType.value === 'zset') {
    const entries = newKeyValue.value.split('\n').filter(l => l.trim())
    if (entries.length > 0) {
      for (const line of entries) {
        const parts = line.trim().split(/\s+/)
        const member = parts[0]
        const score = parseFloat(parts[1]) || 0
        if (member) await app.ZSetAdd(connId, key, member, score)
      }
      if (ttl > 0) await (app as any).SetTTL(connId, key, ttl)
    } else {
      await app.SetStringValue(connId, key, '', ttl)
    }
  }

  showAddDialog.value = false
  newKeyName.value = ''
  newKeyValue.value = ''
  if (isTreeMode.value) {
    store.rootNodes = []
    await store.loadRoot(props.connId)
  } else {
    await doSearch()
  }
}

function typeIcon(type: string) {
  const map: Record<string, any> = { string: Type, hash: Hash, list: List, set: Box, zset: Sliders }
  return map[type] || KeyIcon
}

function typeColor(type: string) {
  const map: Record<string, string> = { string: 'text-accent-green', hash: 'text-accent-orange', list: 'text-accent-blue', set: 'text-accent-red', zset: 'text-accent-blue' }
  return map[type] || 'text-text-muted'
}

function formatKeyName(key: string) {
  const parts = key.split(':')
  if (parts.length > 4) {
    return parts.slice(0, 3).join(':') + ':...' + parts[parts.length - 1]
  }
  return key
}

function onGlobalClick() {
  contextMenu.value.visible = false
}

onMounted(() => {
  document.addEventListener('click', onGlobalClick)
})

onUnmounted(() => {
  document.removeEventListener('click', onGlobalClick)
  if (searchHintTimer) clearTimeout(searchHintTimer)
})
</script>
