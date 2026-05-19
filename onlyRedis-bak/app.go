package main

import (
	"context"
	"log"
	"onlyRedis/services"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the main application struct bound to Wails frontend
type App struct {
	ctx      context.Context
	redis    *services.RedisService
	monitor  *services.MonitorService
}

// NewApp creates a new App instance
func NewApp() *App {
	redisService := services.NewRedisService()
	return &App{
		redis:   redisService,
		monitor: services.NewMonitorService(),
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	log.Printf("[App] onlyRedis backend started")
}

// SetWindowTitle updates the native window title
func (a *App) SetWindowTitle(title string) {
	runtime.WindowSetTitle(a.ctx, title)
}

// ===== Connection Management =====

// GetConnections returns all saved connections as JSON
func (a *App) GetConnections() string {
	return a.redis.GetConnections()
}

// AddConnection adds a new connection configuration
func (a *App) AddConnection(cfg string) string {
	id, err := a.redis.AddConnection(cfg)
	if err != nil {
		return ""
	}
	return id
}

// UpdateConnection updates an existing connection
func (a *App) UpdateConnection(id string, cfg string) error {
	return a.redis.UpdateConnection(id, cfg)
}

// DeleteConnection removes a connection
func (a *App) DeleteConnection(id string) error {
	return a.redis.DeleteConnection(id)
}

// TestConnection tests a Redis connection
func (a *App) TestConnection(cfg string) string {
	return a.redis.TestConnection(cfg)
}

// Connect establishes a Redis connection
func (a *App) Connect(id string) error {
	return a.redis.Connect(id)
}

// Disconnect closes a Redis connection
func (a *App) Disconnect(id string) {
	a.monitor.StopMonitor(id)
	a.redis.Disconnect(id)
}

// ===== Key Operations =====

// ScanKeys scans Redis keys with pattern matching
func (a *App) ScanKeys(connID string, pattern string, cursor uint64, count int64) string {
	return a.redis.ScanKeys(connID, pattern, cursor, count)
}

// SelectDB switches the current database for a connection
func (a *App) SelectDB(connID string, db int) error {
	return a.redis.SelectDB(connID, db)
}

// BrowseKeys browses keys hierarchically - returns direct children of parentPath
func (a *App) BrowseKeys(connID, parentPath string, cursor uint64, count int64) string {
	return a.redis.BrowseKeys(connID, parentPath, cursor, count)
}

// ===== Legacy (kept for compatibility) =====

// GetRedisVersion returns the Redis server version for a connection
func (a *App) GetRedisVersion(connID string) string {
	return a.redis.GetRedisVersion(connID)
}

// GetKeyType returns the type of a key
func (a *App) GetKeyType(connID string, key string) string {
	return a.redis.GetKeyType(connID, key)
}

// GetTTL returns the TTL of a key
func (a *App) GetTTL(connID string, key string) int64 {
	return a.redis.GetTTL(connID, key)
}

// SetTTL sets the TTL of a key
func (a *App) SetTTL(connID string, key string, ttl int64) error {
	return a.redis.SetTTL(connID, key, ttl)
}

// DeleteKey deletes a key
func (a *App) DeleteKey(connID string, key string) int64 {
	return a.redis.DeleteKey(connID, key)
}

// RenameKey renames a key
func (a *App) RenameKey(connID string, oldKey string, newKey string) error {
	return a.redis.RenameKey(connID, oldKey, newKey)
}

// ===== String Operations =====

// GetStringValue retrieves a string value
func (a *App) GetStringValue(connID string, key string) string {
	return a.redis.GetStringValue(connID, key)
}

// SetStringValue sets a string value
func (a *App) SetStringValue(connID string, key string, value string, ttl int64) error {
	return a.redis.SetStringValue(connID, key, value, ttl)
}

// ===== Hash Operations =====

// HashScan scans hash fields
func (a *App) HashScan(connID string, key string, cursor uint64, count int64, withTTL bool) string {
	return a.redis.HashScan(connID, key, cursor, count, withTTL)
}

// HashSet sets a hash field
func (a *App) HashSet(connID string, key string, field string, value string) error {
	return a.redis.HashSet(connID, key, field, value)
}

// HashDel deletes a hash field
func (a *App) HashDel(connID string, key string, field string) error {
	return a.redis.HashDel(connID, key, field)
}

// ===== List Operations =====

// ListRange retrieves list elements
func (a *App) ListRange(connID string, key string, start int64, stop int64) string {
	return a.redis.ListRange(connID, key, start, stop)
}

// ListPush pushes elements to a list
func (a *App) ListPush(connID string, key string, value string, left bool) int64 {
	return a.redis.ListPush(connID, key, value, left)
}

// ListRemove removes elements from a list
func (a *App) ListRemove(connID string, key string, value string, count int64) int64 {
	return a.redis.ListRemove(connID, key, value, count)
}

// ===== Set Operations =====

// SetMembers retrieves set members
func (a *App) SetMembers(connID string, key string, cursor uint64, count int64) string {
	return a.redis.SetMembers(connID, key, cursor, count)
}

// SetAdd adds a member to a set
func (a *App) SetAdd(connID string, key string, member string) error {
	return a.redis.SetAdd(connID, key, member)
}

// SetRemove removes a member from a set
func (a *App) SetRemove(connID string, key string, member string) error {
	return a.redis.SetRemove(connID, key, member)
}

// ===== ZSet Operations =====

// ZSetScan scans sorted set entries
func (a *App) ZSetScan(connID string, key string, cursor uint64, count int64) string {
	return a.redis.ZSetScan(connID, key, cursor, count)
}

// ZSetAdd adds an entry to a sorted set
func (a *App) ZSetAdd(connID string, key string, member string, score float64) error {
	return a.redis.ZSetAdd(connID, key, member, score)
}

// ZSetRemove removes an entry from a sorted set
func (a *App) ZSetRemove(connID string, key string, member string) error {
	return a.redis.ZSetRemove(connID, key, member)
}

// ===== Monitor Operations =====

// StartMonitor starts periodic INFO collection
func (a *App) StartMonitor(connID string, interval int) error {
	a.monitor.StartMonitor(connID, a.redis, interval)
	return nil
}

// StopMonitor stops monitoring for a connection
func (a *App) StopMonitor(connID string) {
	a.monitor.StopMonitor(connID)
}

// GetMonitorData returns the latest monitor data
func (a *App) GetMonitorData(connID string) string {
	return a.monitor.GetMonitorData(connID)
}
