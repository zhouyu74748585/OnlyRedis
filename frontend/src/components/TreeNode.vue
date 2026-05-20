<template>
  <div>
    <!-- Node row -->
    <div
      class="flex items-center gap-1 py-1 rounded cursor-pointer transition-colors select-none tree-node"
      :class="isSelected ? 'bg-blue-50 text-accent-blue' : 'text-text-secondary hover:bg-gray-100'"
      :style="{ paddingLeft: (depth * 16 + 4) + 'px', paddingRight: '4px' }"
      @click="onClick"
      @contextmenu.prevent.stop="onCtxMenu"
    >
      <!-- Expand/collapse toggle -->
      <Loader
        v-if="!node.isLeaf && node.loading"
        class="w-3.5 h-3.5 text-text-muted flex-shrink-0 animate-spin"
      />
      <ChevronRight
        v-else-if="!node.isLeaf"
        class="w-3.5 h-3.5 text-text-muted flex-shrink-0 transition-transform"
        :style="{ transform: !node.collapsed ? 'rotate(90deg)' : 'rotate(0deg)' }"
      />
      <span v-else class="w-3.5 flex-shrink-0" />
      <!-- Icon -->
      <component :is="iconComp" class="w-3.5 h-3.5 flex-shrink-0" :class="iconColor" />
      <!-- Name -->
      <span class="text-[13px] truncate flex-1" :title="node.isLeaf ? node.fullKey : undefined">
        {{ node.name }}
      </span>
      <!-- Type tag / TTL -->
      <span v-if="node.isLeaf" class="text-[11px] text-text-muted flex-shrink-0">{{ node.keyType }}</span>
    </div>
    <!-- Children (when expanded) -->
    <template v-if="!node.isLeaf && !node.collapsed">
      <TreeNode
        v-for="child in node.children"
        :key="child.fullKey"
        :node="child"
        :depth="depth + 1"
        :conn-id="connId"
        :selected-key="selectedKey"
      @select="(key: string, keyType: string) => $emit('select', key, keyType)"
      @contextmenu="(payload: any) => $emit('contextmenu', payload)"
      />
      <!-- Load more button -->
      <div
        v-if="node.hasMore"
        class="flex items-center gap-1 py-1 cursor-pointer text-text-muted hover:text-accent-blue transition-colors select-none"
        :style="{ paddingLeft: ((depth + 1) * 16 + 4) + 'px', paddingRight: '4px' }"
        @click.stop="loadMore"
      >
        <span class="w-3.5 flex-shrink-0" />
        <span class="w-3.5 h-3.5 flex-shrink-0" />
        <span class="text-[12px] truncate flex-1">Load more...</span>
      </div>
      <!-- Empty indicator: folder has no children and loading finished -->
      <div
        v-if="node.children.length === 0 && !node.loading && !node.hasMore"
        class="py-1 select-none"
        :style="{ paddingLeft: ((depth + 1) * 16 + 4) + 'px' }"
      >
        <span class="text-[12px] text-text-muted italic">Empty</span>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ChevronRight, Folder, Key, Type, Hash, List, Box, Sliders, Loader } from 'lucide-vue-next'
import type { TreeNode } from '@/stores/key'
import { useKeyStore } from '@/stores/key'

const props = defineProps<{
  node: TreeNode
  depth: number
  connId: string
  selectedKey: string
}>()

const emit = defineEmits<{
  select: [key: string, keyType: string]
  contextmenu: [payload: { event: MouseEvent; key: string }]
}>()

const store = useKeyStore()

const isSelected = computed(() => props.selectedKey === props.node.fullKey)

const iconMap: Record<string, any> = {
  string: Type, hash: Hash, list: List, set: Box, zset: Sliders,
}
const iconComp = computed(() => iconMap[props.node.keyType] || (props.node.isLeaf ? Key : Folder))

const iconColor = computed(() => {
  const map: Record<string, string> = {
    string: 'text-accent-green', hash: 'text-accent-orange',
    list: 'text-accent-blue', set: 'text-accent-red',
    zset: 'text-accent-blue',
  }
  return map[props.node.keyType] || (props.node.isLeaf ? 'text-text-secondary' : 'text-text-muted')
})

function onClick() {
  if (props.node.isLeaf) {
    emit('select', props.node.fullKey, props.node.keyType)
  } else {
    store.toggleNode(props.connId, props.node)
  }
}

function onCtxMenu(e: MouseEvent) {
  if (props.node.isLeaf) {
    emit('contextmenu', { event: e, key: props.node.fullKey })
  }
}

function loadMore() {
  store.loadMoreChildren(props.connId, props.node)
}
</script>
