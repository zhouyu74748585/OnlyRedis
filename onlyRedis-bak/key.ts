import { defineStore } from 'pinia'
import { reactive } from 'vue'

export interface TreeNode {
  name: string
  fullKey: string
  isLeaf: boolean
  keyType: string    // only for leaf: string/hash/list/set/zset
  ttl: number        // only for leaf
  children: TreeNode[]
  collapsed: boolean  // true = collapsed, not expanded
  hasMore: boolean    // has more children to load (next page)
  cursor: number      // pagination cursor for loading more children
  loading: boolean    // whether children are currently loading
}

export const useKeyStore = defineStore('key', {
  state: () => ({
    rootNodes: [] as TreeNode[],
    rootCursor: 0,
    rootHasMore: false,
    rootLoading: false,
    searchPattern: '*',
    selectedKey: '' as string,
    currentDb: 0,

    // Quick access: flat key list for search mode (when pattern != '*')
    searchKeys: [] as { key: string; type: string; ttl: number }[],
    searchCursor: 0,
    searchComplete: false,
    searchLoading: false,

    openedTabs: [] as string[],
    activeTab: '' as string,
    /** Pre-cached key types: avoids async GetKeyType / WRONGTYPE errors */
    keyTypes: {} as Record<string, string>,
  }),

  actions: {
    async selectDB(connId: string, db: number) {
      await (window.go.main.App as any).SelectDB(connId, db)
      this.currentDb = db
      this.rootNodes = []
      this.searchKeys = []
      await this.loadRoot(connId)
    },

    /** Load root-level nodes (depth 0) with pagination */
    async loadRoot(connId: string, append = false) {
      if (!connId) return
      this.rootLoading = true
      const cursor = append ? this.rootCursor : 0
      const raw = await (window.go.main.App as any).BrowseKeys(connId, '', cursor, 100)
      const page = JSON.parse(raw)
      const nodes = (page.nodes || []).map(toTreeNode)
      if (append) {
        this.rootNodes.push(...nodes)
      } else {
        this.rootNodes = nodes
      }
      this.rootCursor = page.cursor || 0
      this.rootHasMore = page.cursor > 0 && (page.nodes?.length || 0) > 0
      this.rootLoading = false
    },

    /** Load children for a given parent node (lazy, with pagination) */
    async loadChildren(connId: string, node: TreeNode, append = false) {
      if (!connId || node.isLeaf) return
      node.loading = true
      const cursor = append ? node.cursor : 0
      const raw = await (window.go.main.App as any).BrowseKeys(connId, node.fullKey, cursor, 100)
      const page = JSON.parse(raw)
      const nodes = (page.nodes || []).map(toTreeNode)
      if (append) {
        node.children.push(...nodes)
      } else {
        node.children = nodes
      }
      node.cursor = page.cursor || 0
      node.hasMore = page.cursor > 0 && (page.nodes?.length || 0) > 0
      node.loading = false
    },

    /** Toggle collapse/expand of a folder node */
    async toggleNode(connId: string, node: TreeNode) {
      if (node.isLeaf) {
        this.selectKey(node.fullKey)
        return
      }
      if (node.collapsed) {
        // Expanding: load children if not yet loaded
        node.collapsed = false
        if (node.children.length === 0) {
          await this.loadChildren(connId, node)
        }
      } else {
        node.collapsed = true
      }
    },

    /** Load more children for a node (next page) */
    async loadMoreChildren(connId: string, node: TreeNode) {
      await this.loadChildren(connId, node, true)
    },

    /** Search mode: multi-round SCAN for reliable results */
    async scanKeys(connId: string, pattern = '*') {
      this.searchLoading = true
      this.searchPattern = pattern
      this.searchKeys = []

      // Multi-round SCAN to collect enough results
      const maxRounds = 12
      const batchSize = 500
      const maxKeys = 200
      let cursor: number | string = '0'
      const allKeys: { key: string; type: string; ttl: number }[] = []

      for (let round = 0; round < maxRounds && allKeys.length < maxKeys; round++) {
        const raw = await window.go.main.App.ScanKeys(connId, pattern, Number(cursor), batchSize)
        const result = JSON.parse(raw)
        if (result.keys) {
          for (const k of result.keys) {
            if (allKeys.length >= maxKeys) break
            // deduplicate
            if (!allKeys.find(x => x.key === k.key)) {
              allKeys.push(k)
            }
          }
        }
        cursor = result.cursor || 0
        if (cursor === 0) break
      }

      this.searchKeys = allKeys
      this.searchCursor = Number(cursor) || 0
      this.searchComplete = cursor === 0 || allKeys.length >= maxKeys
      this.searchLoading = false
    },

    selectKey(key: string, keyType?: string) {
      this.selectedKey = key
      if (key && !this.openedTabs.includes(key)) {
        this.openedTabs.push(key)
      }
      this.activeTab = key
      // Pre-cache key type to avoid WRONGTYPE errors
      if (keyType && key) {
        this.keyTypes[key] = keyType
      }
    },

    closeTab(key: string) {
      const idx = this.openedTabs.indexOf(key)
      if (idx > -1) {
        this.openedTabs.splice(idx, 1)
        if (this.activeTab === key) {
          this.activeTab = this.openedTabs[Math.max(0, idx - 1)] || ''
        }
      }
    },

    setActiveTab(key: string) {
      this.activeTab = key
      this.selectedKey = key
    },
  },
})

/** Map backend BrowseNode to frontend TreeNode */
function toTreeNode(raw: any): TreeNode {
  return reactive({
    name: raw.name || '',
    fullKey: raw.fullKey || '',
    isLeaf: raw.isLeaf === true,
    keyType: raw.keyType || 'string',
    ttl: Number(raw.ttl) || 0,
    children: [],
    collapsed: true,  // default collapsed, expand on click
    hasMore: false,
    cursor: 0,
    loading: false,
  })
}
