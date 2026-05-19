import { defineStore } from 'pinia'

export interface RedisConnection {
  id: string
  name: string
  host: string
  port: number
  password: string
  db: number
  sshEnabled: boolean
  sshHost: string
  sshPort: number
  sshUser: string
  sshKeyFile: string
  sshPassword: string
  timeout: number
  retries: number
  status: 'disconnected' | 'connecting' | 'connected' | 'error'
}

export const useConnectionStore = defineStore('connection', {
  state: () => ({
    connections: [] as RedisConnection[],
    activeId: '' as string,
    loading: false,
    redisVersions: {} as Record<string, string>, // connId -> version string
  }),

  getters: {
    activeConnection: (state) =>
      state.connections.find((c) => c.id === state.activeId) || null,
    connectedCount: (state) =>
      state.connections.filter((c) => c.status === 'connected').length,
    activeRedisVersion: (state) =>
      state.redisVersions[state.activeId] || '',
  },

  actions: {
    async loadConnections() {
      this.loading = true
      const raw = await window.go.main.App.GetConnections()
      this.connections = JSON.parse(raw)
      this.loading = false
    },

    async addConnection(cfg: RedisConnection) {
      const id = await window.go.main.App.AddConnection(JSON.stringify(cfg))
      await this.loadConnections()
      return id
    },

    async updateConnection(id: string, cfg: RedisConnection) {
      await window.go.main.App.UpdateConnection(id, JSON.stringify(cfg))
      await this.loadConnections()
    },

    async deleteConnection(id: string) {
      await window.go.main.App.DeleteConnection(id)
      if (this.activeId === id) this.activeId = ''
      await this.loadConnections()
    },

    async testConnection(cfg: RedisConnection) {
      const result = await window.go.main.App.TestConnection(JSON.stringify(cfg))
      return JSON.parse(result)
    },

    async connect(id: string) {
      const conn = this.connections.find((c) => c.id === id)
      if (!conn) return
      conn.status = 'connecting'
      try {
        await window.go.main.App.Connect(id)
        conn.status = 'connected'
        this.activeId = id
        // Fetch Redis version after successful connection
        try {
          this.redisVersions[id] = await window.go.main.App.GetRedisVersion(id)
        } catch {
          this.redisVersions[id] = ''
        }
      } catch (e) {
        conn.status = 'disconnected'
        throw e
      }
    },

    async disconnect(id: string) {
      await window.go.main.App.Disconnect(id)
      const conn = this.connections.find((c) => c.id === id)
      if (conn) conn.status = 'disconnected'
      delete this.redisVersions[id]
    },

    setActive(id: string) {
      this.activeId = id
    },
  },
})
