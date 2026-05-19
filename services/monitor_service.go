package services

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

// MonitorData holds parsed Redis INFO metrics
type MonitorData struct {
	UsedMemory       int64   `json:"usedMemory"`
	UsedMemoryHuman  string  `json:"usedMemoryHuman"`
	PeakMemoryHuman  string  `json:"peakMemoryHuman"`
	ConnectedClients int64   `json:"connectedClients"`
	BlockedClients   int64   `json:"blockedClients"`
	HitRate          float64 `json:"hitRate"`
	MissRate         float64 `json:"missRate"`
	QPS              float64 `json:"qps"`
	KeysCount        int64   `json:"keysCount"`
	ExpiresCount     int64   `json:"expiresCount"`
	AvgTTL           float64 `json:"avgTTL"`
	CPUSys           float64 `json:"cpuSys"`
	CPUUser          float64 `json:"cpuUser"`
	NetInputRate     float64 `json:"netInputRate"`
	NetOutputRate    float64 `json:"netOutputRate"`
	UptimeInSeconds  int64   `json:"uptimeInSeconds"`
	OpsPerSec        int64   `json:"opsPerSec"`
	Timestamp        int64   `json:"timestamp"`
}

// MonitorResult wraps monitor data for frontend consumption
type MonitorResult struct {
	Data *MonitorData `json:"data"`
}

// MonitorService manages periodic INFO collection
type MonitorService struct {
	mu        sync.Mutex
	tickers   map[string]*time.Ticker
	stopChans map[string]chan struct{}
	dataStore map[string]*MonitorData
	prevOps   map[string]int64
	prevTime  map[string]time.Time
}

// NewMonitorService creates a new monitor service
func NewMonitorService() *MonitorService {
	return &MonitorService{
		tickers:   make(map[string]*time.Ticker),
		stopChans: make(map[string]chan struct{}),
		dataStore: make(map[string]*MonitorData),
		prevOps:   make(map[string]int64),
		prevTime:  make(map[string]time.Time),
	}
}

// StartMonitor begins periodic INFO collection for a connection
func (ms *MonitorService) StartMonitor(connID string, redisService *RedisService, interval int) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Stop existing monitor if any
	if ch, ok := ms.stopChans[connID]; ok {
		close(ch)
	}

	stopChan := make(chan struct{})
	ms.stopChans[connID] = stopChan

	if interval <= 0 {
		interval = 2
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	ms.tickers[connID] = ticker

	log.Printf("[Monitor] Start: conn=%s, interval=%ds", connID, interval)

	go func() {
		// Collect immediately
		ms.collect(connID, redisService)

		for {
			select {
			case <-ticker.C:
				ms.collect(connID, redisService)
			case <-stopChan:
				ticker.Stop()
				log.Printf("[Monitor] Stop: conn=%s, ticker stopped", connID)
				return
			}
		}
	}()
}

// StopMonitor stops monitoring for a connection
func (ms *MonitorService) StopMonitor(connID string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ch, ok := ms.stopChans[connID]; ok {
		close(ch)
		delete(ms.stopChans, connID)
	}
	if ticker, ok := ms.tickers[connID]; ok {
		ticker.Stop()
		delete(ms.tickers, connID)
	}
}

// GetMonitorData returns the latest monitor data
func (ms *MonitorService) GetMonitorData(connID string) string {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	data, ok := ms.dataStore[connID]
	if !ok {
		return `{"data":null}`
	}

	result := MonitorResult{Data: data}
	raw, _ := json.Marshal(result)
	return string(raw)
}

// collect fetches and parses Redis INFO
func (ms *MonitorService) collect(connID string, redisService *RedisService) {
	info, err := redisService.getClient(connID)
	if err != nil {
		log.Printf("[Monitor] collect: conn=%s, getClient failed: %v", connID, err)
		return
	}

	ctx := context.Background()
	infoStr, err := info.Info(ctx).Result()
	if err != nil {
		log.Printf("[Monitor] collect: conn=%s, INFO failed: %v", connID, err)
		return
	}

	data := parseInfo(infoStr)

	// Calculate QPS based on ops counter delta
	ms.mu.Lock()
	if prevOps, ok := ms.prevOps[connID]; ok {
		prevTime := ms.prevTime[connID]
		elapsed := time.Since(prevTime).Seconds()
		if elapsed > 0 {
			data.QPS = float64(data.OpsPerSec-prevOps) / elapsed
		}
	}
	ms.prevOps[connID] = data.OpsPerSec
	ms.prevTime[connID] = time.Now()
	ms.dataStore[connID] = data
	ms.mu.Unlock()

	log.Printf("[Monitor] collect: conn=%s, mem=%s, clients=%d, qps=%.0f, hitRate=%.1f%%",
		connID, data.UsedMemoryHuman, data.ConnectedClients, data.QPS, data.HitRate)
}

// parseInfo parses Redis INFO command output
func parseInfo(infoStr string) *MonitorData {
	data := &MonitorData{Timestamp: time.Now().Unix()}

	lines := strings.Split(infoStr, "\r\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key, value := parts[0], parts[1]

		switch key {
		case "used_memory":
			data.UsedMemory, _ = strconv.ParseInt(value, 10, 64)
		case "used_memory_human":
			data.UsedMemoryHuman = value
		case "used_memory_peak_human":
			data.PeakMemoryHuman = value
		case "connected_clients":
			data.ConnectedClients, _ = strconv.ParseInt(value, 10, 64)
		case "blocked_clients":
			data.BlockedClients, _ = strconv.ParseInt(value, 10, 64)
		case "keyspace_hits":
			hits, _ := strconv.ParseFloat(value, 64)
			if missStr := findInfoValue(lines, "keyspace_misses"); missStr != "" {
				misses, _ := strconv.ParseFloat(missStr, 64)
				total := hits + misses
				if total > 0 {
					data.HitRate = hits / total * 100
					data.MissRate = misses / total * 100
				}
			}
		case "instantaneous_ops_per_sec":
			data.OpsPerSec, _ = strconv.ParseInt(value, 10, 64)
		case "uptime_in_seconds":
			data.UptimeInSeconds, _ = strconv.ParseInt(value, 10, 64)
		case "used_cpu_sys":
			data.CPUSys, _ = strconv.ParseFloat(value, 64)
		case "used_cpu_user":
			data.CPUUser, _ = strconv.ParseFloat(value, 64)
		case "instantaneous_input_kbps":
			data.NetInputRate, _ = strconv.ParseFloat(value, 64)
		case "instantaneous_output_kbps":
			data.NetOutputRate, _ = strconv.ParseFloat(value, 64)
		}

		// Parse db stats (db0:keys=100,expires=50,avg_ttl=3600)
		if strings.HasPrefix(key, "db") {
			dbFields := strings.Split(value, ",")
			for _, field := range dbFields {
				fp := strings.SplitN(field, "=", 2)
				if len(fp) == 2 {
					switch fp[0] {
					case "keys":
						k, _ := strconv.ParseInt(fp[1], 10, 64)
						data.KeysCount += k
					case "expires":
						e, _ := strconv.ParseInt(fp[1], 10, 64)
						data.ExpiresCount += e
					case "avg_ttl":
						ttl, _ := strconv.ParseFloat(fp[1], 64)
						data.AvgTTL = ttl
					}
				}
			}
		}
	}

	return data
}

// findInfoValue finds a value by key in INFO lines
func findInfoValue(lines []string, key string) string {
	for _, line := range lines {
		if strings.HasPrefix(line, key+":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return parts[1]
			}
		}
	}
	return ""
}
