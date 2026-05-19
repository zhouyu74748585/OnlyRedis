/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

interface Window {
  go: {
    main: {
      App: {
        Greet(name: string): Promise<string>
        // Redis Connection
        AddConnection(cfg: string): Promise<string>
        UpdateConnection(id: string, cfg: string): Promise<void>
        DeleteConnection(id: string): Promise<void>
        GetConnections(): Promise<string>
        TestConnection(cfg: string): Promise<string>
        Connect(id: string): Promise<void>
        Disconnect(id: string): Promise<void>
        // Window
        SetWindowTitle(title: string): Promise<void>
        // Key Operations
        ScanKeys(connId: string, pattern: string, cursor: number, count: number): Promise<string>
        GetKeyType(connId: string, key: string): Promise<string>
        GetRedisVersion(connId: string): Promise<string>
        GetTTL(connId: string, key: string): Promise<number>
        SetTTL(connId: string, key: string, ttl: number): Promise<void>
        DeleteKey(connId: string, key: string): Promise<number>
        RenameKey(connId: string, oldKey: string, newKey: string): Promise<void>
        // String Operations
        GetStringValue(connId: string, key: string): Promise<string>
        SetStringValue(connId: string, key: string, value: string, ttl: number): Promise<void>
        // Hash Operations
        HashScan(connId: string, key: string, cursor: number, count: number, withTTL: boolean): Promise<string>
        HashSet(connId: string, key: string, field: string, value: string): Promise<void>
        HashDel(connId: string, key: string, field: string): Promise<void>
        // List Operations
        ListRange(connId: string, key: string, start: number, stop: number): Promise<string>
        ListPush(connId: string, key: string, value: string, left: boolean): Promise<number>
        ListRemove(connId: string, key: string, value: string, count: number): Promise<number>
        // Set Operations
        SetMembers(connId: string, key: string, cursor: number, count: number): Promise<string>
        SetAdd(connId: string, key: string, member: string): Promise<void>
        SetRemove(connId: string, key: string, member: string): Promise<void>
        // ZSet Operations
        ZSetScan(connId: string, key: string, cursor: number, count: number): Promise<string>
        ZSetAdd(connId: string, key: string, member: string, score: number): Promise<void>
        ZSetRemove(connId: string, key: string, member: string): Promise<void>
        // Monitor
        StartMonitor(connId: string, interval: number): Promise<void>
        StopMonitor(connId: string): Promise<void>
        GetMonitorData(connId: string): Promise<string>
      }
    }
  }
}
