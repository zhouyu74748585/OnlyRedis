import { defineStore } from 'pinia'

export interface MonitorData {
  usedMemory: string
  usedMemoryHuman: string
  peakMemoryHuman: string
  connectedClients: number
  blockedClients: number
  hitRate: number
  missRate: number
  qps: number
  keysCount: number
  expiresCount: number
  avgTTL: number
  cpuSys: number
  cpuUser: number
  netInputRate: number
  netOutputRate: number
  uptimeInSeconds: number
  opsPerSec: number
  timestamp: number
}

export interface QpsPoint {
  time: string
  qps: number
}

export const useMonitorStore = defineStore('monitor', {
  state: () => ({
    data: null as MonitorData | null,
    qpsHistory: [] as QpsPoint[],
    isMonitoring: false,
    interval: 2,
  }),

  actions: {
    async startMonitoring(connId: string) {
      this.isMonitoring = true
      await window.go.main.App.StartMonitor(connId, this.interval)
    },

    async stopMonitoring(connId: string) {
      this.isMonitoring = false
      await window.go.main.App.StopMonitor(connId)
    },

    async fetchData(connId: string) {
      const raw = await window.go.main.App.GetMonitorData(connId)
      const parsed = JSON.parse(raw)
      if (parsed.data) {
        this.data = parsed.data
        if (parsed.data.qps !== undefined) {
          const now = new Date()
          const time = `${now.getHours().toString().padStart(2, '0')}:${now.getMinutes().toString().padStart(2, '0')}:${now.getSeconds().toString().padStart(2, '0')}`
          this.qpsHistory.push({ time, qps: parsed.data.qps })
          if (this.qpsHistory.length > 60) {
            this.qpsHistory.shift()
          }
        }
      }
    },
  },
})
