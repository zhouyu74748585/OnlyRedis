package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/ssh"
)

// KeyInfo represents a scanned Redis key
type KeyInfo struct {
	Key  string `json:"key"`
	Type string `json:"type"`
	TTL  int64  `json:"ttl"`
}

// ScanResult represents the result of a SCAN operation
type ScanResult struct {
	Keys   []KeyInfo `json:"keys"`
	Cursor uint64    `json:"cursor"`
}

// HashField represents a hash field-value pair
type HashField struct {
	Field string `json:"field"`
	Value string `json:"value"`
	TTL   int64  `json:"ttl"` // field-level TTL (-2 = no TTL set, -1 = no expiry, >=0 = remaining seconds)
}

// HashScanResult represents paginated hash scan result
type HashScanResult struct {
	Fields []HashField `json:"fields"`
	Total  int         `json:"total"`
	Cursor uint64      `json:"cursor"`
}

// SetScanResult represents paginated set scan result
type SetScanResult struct {
	Members []string `json:"members"`
	Total   int      `json:"total"`
	Cursor  uint64   `json:"cursor"`
}

// ZSetEntry represents a sorted set entry
type ZSetEntry struct {
	Member string  `json:"member"`
	Score  float64 `json:"score"`
}

// ZSetScanResult represents paginated zset scan result
type ZSetScanResult struct {
	Entries []ZSetEntry `json:"entries"`
	Total   int         `json:"total"`
	Cursor  uint64      `json:"cursor"`
}

// RedisService manages all Redis client interactions
type RedisService struct {
	mu          sync.RWMutex
	clients     map[string]*redis.Client
	sshClients  map[string]*ssh.Client
	configStore *ConfigStore
	connections []ConnectionConfig
}

// NewRedisService creates a new RedisService instance
func NewRedisService() *RedisService {
	return &RedisService{
		clients:     make(map[string]*redis.Client),
		sshClients:  make(map[string]*ssh.Client),
		configStore: NewConfigStore(),
		connections: nil,
	}
}

// LoadConnections loads saved configurations
func (rs *RedisService) LoadConnections() ([]ConnectionConfig, error) {
	configs, err := rs.configStore.LoadConnections()
	if err != nil {
		return nil, err
	}
	rs.connections = configs
	// Mask passwords before returning
	masked := make([]ConnectionConfig, len(configs))
	for i, c := range configs {
		masked[i] = c
		masked[i].Password = MaskPassword(c.Password)
		masked[i].SSHPassword = MaskPassword(c.SSHPassword)
	}
	return masked, nil
}

// GetConnections returns all connections as JSON
func (rs *RedisService) GetConnections() string {
	configs, _ := rs.LoadConnections()
	data, _ := json.Marshal(configs)
	return string(data)
}

// AddConnection adds a new connection configuration
func (rs *RedisService) AddConnection(cfgJSON string) (string, error) {
	var cfg ConnectionConfig
	if err := json.Unmarshal([]byte(cfgJSON), &cfg); err != nil {
		log.Printf("[Redis] AddConnection: invalid config: %v", err)
		return "", fmt.Errorf("invalid config: %v", err)
	}
	cfg.ID = fmt.Sprintf("conn_%d", time.Now().UnixNano())

	rs.mu.Lock()
	rs.connections = append(rs.connections, cfg)
	rs.mu.Unlock()

	rs.configStore.SaveConnections(rs.connections)
	log.Printf("[Redis] AddConnection: added \"%s\" (%s:%d) id=%s", cfg.Name, cfg.Host, cfg.Port, cfg.ID)
	return cfg.ID, nil
}

// UpdateConnection updates an existing connection configuration
func (rs *RedisService) UpdateConnection(id string, cfgJSON string) error {
	var cfg ConnectionConfig
	if err := json.Unmarshal([]byte(cfgJSON), &cfg); err != nil {
		return err
	}

	rs.mu.Lock()
	defer rs.mu.Unlock()

	for i, c := range rs.connections {
		if c.ID == id {
			// Preserve password if not provided
			if cfg.Password == "****" {
				cfg.Password = c.Password
			}
			if cfg.SSHPassword == "****" {
				cfg.SSHPassword = c.SSHPassword
			}
			cfg.ID = id
			rs.connections[i] = cfg
			break
		}
	}
	return rs.configStore.SaveConnections(rs.connections)
}

// DeleteConnection removes a connection
func (rs *RedisService) DeleteConnection(id string) error {
	rs.Disconnect(id)

	rs.mu.Lock()
	defer rs.mu.Unlock()

	var filtered []ConnectionConfig
	for _, c := range rs.connections {
		if c.ID != id {
			filtered = append(filtered, c)
		}
	}
	rs.connections = filtered
	return rs.configStore.SaveConnections(rs.connections)
}

// TestConnection tests a Redis connection without saving
func (rs *RedisService) TestConnection(cfgJSON string) string {
	var cfg ConnectionConfig
	if err := json.Unmarshal([]byte(cfgJSON), &cfg); err != nil {
		return `{"success":false,"error":"invalid config"}`
	}

	client, err := rs.createClient(&cfg)
	if err != nil {
		return `{"success":false,"error":"` + err.Error() + `"}`
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return `{"success":false,"error":"` + err.Error() + `"}`
	}

	return `{"success":true}`
}

// Connect establishes a Redis connection
func (rs *RedisService) Connect(id string) error {
	rs.mu.RLock()
	var cfg *ConnectionConfig
	for _, c := range rs.connections {
		if c.ID == id {
			cfg = &c
			break
		}
	}
	rs.mu.RUnlock()

	if cfg == nil {
		log.Printf("[Redis] Connect: connection %s not found", id)
		return fmt.Errorf("connection %s not found", id)
	}

	log.Printf("[Redis] Connect: connecting to \"%s\" (%s:%d) db=%d ssh=%v", cfg.Name, cfg.Host, cfg.Port, cfg.DB, cfg.SSHEnabled)

	client, err := rs.createClient(cfg)
	if err != nil {
		log.Printf("[Redis] Connect: create client failed: %v", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		log.Printf("[Redis] Connect: ping failed: %v", err)
		return err
	}

	rs.mu.Lock()
	// Close existing client if any
	if existing, ok := rs.clients[id]; ok {
		existing.Close()
	}
	rs.clients[id] = client
	rs.mu.Unlock()

	log.Printf("[Redis] Connect: connected to \"%s\" (%s:%d) successfully", cfg.Name, cfg.Host, cfg.Port)
	return nil
}

// Disconnect closes a Redis connection
func (rs *RedisService) Disconnect(id string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if client, ok := rs.clients[id]; ok {
		client.Close()
		delete(rs.clients, id)
		log.Printf("[Redis] Disconnect: closed redis client for id=%s", id)
	}
	if sshClient, ok := rs.sshClients[id]; ok {
		sshClient.Close()
		delete(rs.sshClients, id)
		log.Printf("[Redis] Disconnect: closed ssh tunnel for id=%s", id)
	}
}

// createClient creates a Redis client, optionally through SSH tunnel
func (rs *RedisService) createClient(cfg *ConnectionConfig) (*redis.Client, error) {
	opts := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	}

	if cfg.SSHEnabled {
		dialer, err := rs.createSSHTunnel(cfg)
		if err != nil {
			return nil, fmt.Errorf("ssh tunnel failed: %v", err)
		}
		opts.Dialer = dialer
	}

	return redis.NewClient(opts), nil
}

// createSSHTunnel creates an SSH tunnel dialer
func (rs *RedisService) createSSHTunnel(cfg *ConnectionConfig) (func(ctx context.Context, network, addr string) (net.Conn, error), error) {
	log.Printf("[Redis] createSSHTunnel: establishing SSH tunnel via %s@%s:%d", cfg.SSHUser, cfg.SSHHost, cfg.SSHPort)

	sshConfig := &ssh.ClientConfig{
		User:            cfg.SSHUser,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	if cfg.SSHPassword != "" {
		sshConfig.Auth = []ssh.AuthMethod{ssh.Password(cfg.SSHPassword)}
	} else if cfg.SSHKeyFile != "" {
		key, err := os.ReadFile(cfg.SSHKeyFile)
		if err != nil {
			log.Printf("[Redis] createSSHTunnel: read key file failed: %v", err)
			return nil, fmt.Errorf("read ssh key: %v", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			log.Printf("[Redis] createSSHTunnel: parse private key failed: %v", err)
			return nil, fmt.Errorf("parse ssh key: %v", err)
		}
		sshConfig.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	} else {
		return nil, fmt.Errorf("no ssh credentials provided")
	}

	sshAddr := fmt.Sprintf("%s:%d", cfg.SSHHost, cfg.SSHPort)
	sshClient, err := ssh.Dial("tcp", sshAddr, sshConfig)
	if err != nil {
		log.Printf("[Redis] createSSHTunnel: ssh dial %s failed: %v", sshAddr, err)
		return nil, err
	}

	// Store ssh client for cleanup
	rs.mu.Lock()
	rs.sshClients[cfg.ID] = sshClient
	rs.mu.Unlock()

	log.Printf("[Redis] createSSHTunnel: SSH tunnel established to %s", sshAddr)
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return sshClient.Dial(network, addr)
	}, nil
}

// getClient retrieves a connected Redis client
func (rs *RedisService) getClient(connID string) (*redis.Client, error) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	client, ok := rs.clients[connID]
	if !ok {
		return nil, fmt.Errorf("not connected: %s", connID)
	}
	return client, nil
}

// BrowseNode represents a single node in the key tree browser
type BrowseNode struct {
	Name     string `json:"name"`     // display name (single segment)
	FullKey  string `json:"fullKey"`  // full key path to this node
	IsLeaf   bool   `json:"isLeaf"`   // true = actual key, false = folder/namespace
	KeyType  string `json:"keyType"`  // only for leaf: string/hash/list/set/zset
	TTL      int64  `json:"ttl"`      // only for leaf
	HasMore  bool   `json:"hasMore"`  // true if this folder may have more children (for lazy load)
}

// BrowsePage represents a page of direct child nodes
type BrowsePage struct {
	Nodes  []BrowseNode `json:"nodes"`
	Cursor uint64        `json:"cursor"`  // next cursor, 0 means complete
	Total  int           `json:"total"`   // number of nodes in this page
}

// BrowseKeys browses keys hierarchically - returns only direct children of parentPath
// parentPath: e.g. "" for root, "user" for user:* level
// cursor: pagination cursor, 0 to start
// count: page size (default 100)
func (rs *RedisService) BrowseKeys(connID, parentPath string, cursor uint64, count int64) string {
	client, err := rs.getClient(connID)
	if err != nil {
		log.Printf("[Redis] BrowseKeys: conn=%s, err=%v", connID, err)
		return fmt.Sprintf(`{"error":"%s"}`, err.Error())
	}

	if count <= 0 {
		count = 100
	}

	// Build match pattern: parentPath:* or just * for root
	pattern := buildBrowsePattern(parentPath)

	ctx := context.Background()
	// Collect enough keys to fill a page of distinct direct children
	// We SCAN in larger batches and deduplicate
	nodeMap := make(map[string]*BrowseNode)
	var nodeOrder []string
	scanCursor := cursor
	maxScanRounds := 12
	scanBatch := max(count*4, 500)
	finished := false

	for round := 0; round < maxScanRounds && len(nodeMap) < int(count) && !finished; round++ {
		keys, nextCursor, err := client.Scan(ctx, scanCursor, pattern, scanBatch).Result()
		if err != nil {
			log.Printf("[Redis] BrowseKeys: scan failed, conn=%s, parentPath=%s, err=%v", connID, parentPath, err)
			return fmt.Sprintf(`{"error":"%s"}`, err.Error())
		}

		for _, key := range keys {
			if len(nodeMap) >= int(count) {
				break
			}
			childName := extractDirectChild(parentPath, key)
			if childName == "" {
				continue
			}
			fullPath := childName
			if parentPath != "" {
				fullPath = parentPath + ":" + childName
			}
			// Determine if it's a leaf or folder
			relative := key
			if parentPath != "" && len(key) > len(parentPath)+1 {
				relative = key[len(parentPath)+1:]
			}
			isLeaf := !strings.Contains(relative, ":")
			nodeKey := fullPath

			if _, exists := nodeMap[nodeKey]; !exists {
				nodeMap[nodeKey] = &BrowseNode{
					Name:    childName,
					FullKey: fullPath,
					IsLeaf:  isLeaf,
				}
				nodeOrder = append(nodeOrder, nodeKey)
			}
		}

		scanCursor = nextCursor
		finished = nextCursor == 0
	}

	// Batch fetch type & ttl for leaf nodes using pipeline
	var leafNodes []*BrowseNode
	for _, key := range nodeOrder {
		node := nodeMap[key]
		if node.IsLeaf {
			leafNodes = append(leafNodes, node)
		}
	}

	if len(leafNodes) > 0 {
		pipe := client.Pipeline()
		typeCmds := make([]*redis.StatusCmd, len(leafNodes))
		ttlCmds := make([]*redis.DurationCmd, len(leafNodes))
		for i, node := range leafNodes {
			typeCmds[i] = pipe.Type(ctx, node.FullKey)
			ttlCmds[i] = pipe.TTL(ctx, node.FullKey)
		}
		_, pipeErr := pipe.Exec(ctx)
		if pipeErr != nil {
			log.Printf("[Redis] BrowseKeys: pipeline exec failed, conn=%s, err=%v", connID, pipeErr)
		}
		for i, node := range leafNodes {
			if pipeErr == nil {
				node.KeyType, _ = typeCmds[i].Result()
				ttl, _ := ttlCmds[i].Result()
				node.TTL = int64(ttl.Seconds())
			}
		}
	}

	// Build result preserving order
	nodes := make([]BrowseNode, 0, len(nodeOrder))
	for _, key := range nodeOrder {
		nodes = append(nodes, *nodeMap[key])
	}

	nextCursor := scanCursor
	if finished {
		nextCursor = 0
	}

	page := BrowsePage{Nodes: nodes, Cursor: nextCursor, Total: len(nodes)}
	data, _ := json.Marshal(page)
	log.Printf("[Redis] BrowseKeys: conn=%s, parentPath=%s, nodes=%d, cursor=%d, finished=%v",
		connID, parentPath, len(nodes), nextCursor, finished)
	return string(data)
}

// buildBrowsePattern creates SCAN match pattern for browsing
// "" → "*", "user" → "user:*"
func buildBrowsePattern(parentPath string) string {
	if parentPath == "" {
		return "*"
	}
	return escapeGlob(parentPath) + ":*"
}

// extractDirectChild returns the first segment of key after parentPath
// parentPath="" key="user:1:name" → "user"
// parentPath="user" key="user:1:name" → "1"
// parentPath="user:1" key="user:1:name" → "name"
func extractDirectChild(parentPath, key string) string {
	if parentPath == "" {
		if idx := strings.Index(key, ":"); idx > 0 {
			return key[:idx]
		}
		return key
	}
	prefix := parentPath + ":"
	if !strings.HasPrefix(key, prefix) || len(key) <= len(prefix) {
		return ""
	}
	relative := key[len(prefix):]
	if idx := strings.Index(relative, ":"); idx > 0 {
		return relative[:idx]
	}
	return relative
}

// escapeGlob escapes Redis glob special characters
func escapeGlob(s string) string {
	var b strings.Builder
	for _, ch := range s {
		switch ch {
		case '*', '?', '[', ']', '\\':
			b.WriteRune('\\')
		}
		b.WriteRune(ch)
	}
	return b.String()
}

// ScanKeys performs SCAN operation for key browsing
func (rs *RedisService) ScanKeys(connID, pattern string, cursor uint64, count int64) string {
	client, err := rs.getClient(connID)
	if err != nil {
		log.Printf("[Redis] ScanKeys: conn=%s, err=%v", connID, err)
		return fmt.Sprintf(`{"error":"%s"}`, err.Error())
	}

	if count <= 0 {
		count = 200
	}

	ctx := context.Background()
	keys, nextCursor, err := client.Scan(ctx, cursor, pattern, count).Result()
	if err != nil {
		log.Printf("[Redis] ScanKeys: scan failed, conn=%s, pattern=%s, err=%v", connID, pattern, err)
		return fmt.Sprintf(`{"error":"%s"}`, err.Error())
	}

	log.Printf("[Redis] ScanKeys: conn=%s, pattern=%s, count=%d, found=%d keys", connID, pattern, count, len(keys))

	var keyInfos []KeyInfo
	for _, key := range keys {
		keyType, _ := client.Type(ctx, key).Result()
		ttl, _ := client.TTL(ctx, key).Result()
		keyInfos = append(keyInfos, KeyInfo{
			Key:  key,
			Type: keyType,
			TTL:  int64(ttl.Seconds()),
		})
	}

	result := ScanResult{Keys: keyInfos, Cursor: nextCursor}
	data, _ := json.Marshal(result)
	return string(data)
}

// GetRedisVersion returns the Redis server version string (e.g. "7.4.0")
func (rs *RedisService) GetRedisVersion(connID string) string {
	client, err := rs.getClient(connID)
	if err != nil {
		log.Printf("[Redis] GetRedisVersion: conn=%s, err=%v", connID, err)
		return ""
	}
	info, err := client.Info(context.Background(), "server").Result()
	if err != nil {
		log.Printf("[Redis] GetRedisVersion: info failed, conn=%s, err=%v", connID, err)
		return ""
	}
	// Parse redis_version from INFO SERVER output
	for _, line := range strings.Split(info, "\r\n") {
		if strings.HasPrefix(line, "redis_version:") {
			version := strings.TrimPrefix(line, "redis_version:")
			log.Printf("[Redis] GetRedisVersion: conn=%s, version=%s", connID, version)
			return version
		}
	}
	log.Printf("[Redis] GetRedisVersion: version not found in info, conn=%s", connID)
	return ""
}

// GetKeyType returns the type of a given key
func (rs *RedisService) GetKeyType(connID, key string) string {
	client, err := rs.getClient(connID)
	if err != nil {
		return ""
	}
	keyType, err := client.Type(context.Background(), key).Result()
	if err != nil {
		return ""
	}
	return keyType
}

// GetTTL returns the TTL of a key
func (rs *RedisService) GetTTL(connID, key string) int64 {
	client, err := rs.getClient(connID)
	if err != nil {
		return -2
	}
	ttl, err := client.TTL(context.Background(), key).Result()
	if err != nil {
		return -2
	}
	return int64(ttl.Seconds())
}

// SetTTL sets the TTL of a key
func (rs *RedisService) SetTTL(connID, key string, ttl int64) error {
	client, err := rs.getClient(connID)
	if err != nil {
		return err
	}
	if ttl < 0 {
		return client.Persist(context.Background(), key).Err()
	}
	return client.Expire(context.Background(), key, time.Duration(ttl)*time.Second).Err()
}

// DeleteKey deletes one or more keys
func (rs *RedisService) DeleteKey(connID, key string) int64 {
	client, err := rs.getClient(connID)
	if err != nil {
		log.Printf("[Redis] DeleteKey: conn=%s, key=%s, err=%v", connID, key, err)
		return 0
	}
	count, err := client.Del(context.Background(), key).Result()
	if err != nil {
		log.Printf("[Redis] DeleteKey: del failed, conn=%s, key=%s, err=%v", connID, key, err)
		return 0
	}
	log.Printf("[Redis] DeleteKey: conn=%s, key=%s, deleted=%d", connID, key, count)
	return count
}

// SelectDB switches the current database for a connection
func (rs *RedisService) SelectDB(connID string, db int) error {
	client, err := rs.getClient(connID)
	if err != nil {
		log.Printf("[Redis] SelectDB: conn=%s, err=%v", connID, err)
		return err
	}
	if db < 0 || db > 15 {
		return fmt.Errorf("invalid db index: %d (must be 0-15)", db)
	}
	// Redis SELECT command via Do
	_, err = client.Do(context.Background(), "SELECT", db).Result()
	if err != nil {
		log.Printf("[Redis] SelectDB: SELECT %d failed, conn=%s, err=%v", db, connID, err)
		return err
	}
	log.Printf("[Redis] SelectDB: conn=%s, switched to db=%d", connID, db)
	return nil
}

// RenameKey renames a key
func (rs *RedisService) RenameKey(connID, oldKey, newKey string) error {
	client, err := rs.getClient(connID)
	if err != nil {
		return err
	}
	return client.Rename(context.Background(), oldKey, newKey).Err()
}

// --- String Operations ---

// GetStringValue retrieves a string value
func (rs *RedisService) GetStringValue(connID, key string) string {
	client, err := rs.getClient(connID)
	if err != nil {
		log.Printf("[Redis] GetString: conn=%s, key=%s, err=%v", connID, key, err)
		return ""
	}
	val, err := client.Get(context.Background(), key).Result()
	if err != nil {
		log.Printf("[Redis] GetString: get failed, conn=%s, key=%s, err=%v", connID, key, err)
		return ""
	}
	log.Printf("[Redis] GetString: conn=%s, key=%s, len=%d", connID, key, len(val))
	return val
}

// SetStringValue sets a string value with optional TTL
func (rs *RedisService) SetStringValue(connID, key, value string, ttl int64) error {
	client, err := rs.getClient(connID)
	if err != nil {
		log.Printf("[Redis] SetString: conn=%s, key=%s, err=%v", connID, key, err)
		return err
	}
	if ttl > 0 {
		err = client.SetEx(context.Background(), key, value, time.Duration(ttl)*time.Second).Err()
	} else {
		err = client.Set(context.Background(), key, value, 0).Err()
	}
	if err != nil {
		log.Printf("[Redis] SetString: set failed, conn=%s, key=%s, ttl=%d, err=%v", connID, key, ttl, err)
	} else {
		log.Printf("[Redis] SetString: conn=%s, key=%s, len=%d, ttl=%d", connID, key, len(value), ttl)
	}
	return err
}

// --- Hash Operations ---

// HashScan scans hash fields with pagination.
// withTTL: if true, fetches per-field TTL using HTTL (requires Redis ≥ 7.4.0)
func (rs *RedisService) HashScan(connID, key string, cursor uint64, count int64, withTTL bool) string {
	client, err := rs.getClient(connID)
	if err != nil {
		log.Printf("[Redis] HashScan: conn=%s, key=%s, err=%v", connID, key, err)
		return fmt.Sprintf(`{"error":"%s"}`, err.Error())
	}
	if count <= 0 {
		count = 100
	}

	ctx := context.Background()
	total, _ := client.HLen(ctx, key).Result()
	fields, nextCursor, err := client.HScan(ctx, key, cursor, "*", count).Result()
	if err != nil {
		log.Printf("[Redis] HashScan: hscan failed, conn=%s, key=%s, err=%v", connID, key, err)
		return fmt.Sprintf(`{"error":"%s"}`, err.Error())
	}

	log.Printf("[Redis] HashScan: conn=%s, key=%s, total=%d, fetched=%d, withTTL=%v", connID, key, total, len(fields)/2, withTTL)

	var fieldList []HashField
	for i := 0; i < len(fields); i += 2 {
		fieldList = append(fieldList, HashField{Field: fields[i], Value: fields[i+1], TTL: -2})
	}

	// Fetch per-field TTLs using HTTL command (Redis ≥ 7.4.0)
	if withTTL && len(fieldList) > 0 {
		pipe := client.Pipeline()
		ttlCmds := make([]*redis.Cmd, len(fieldList))
		for i, f := range fieldList {
			ttlCmds[i] = pipe.Do(ctx, "HTTL", key, f.Field)
		}
		_, pipeErr := pipe.Exec(ctx)
		if pipeErr != nil {
			log.Printf("[Redis] HashScan: HTTL pipeline failed, conn=%s, key=%s, err=%v", connID, key, pipeErr)
		} else {
			for i, cmd := range ttlCmds {
				ttl, _ := cmd.Int()
				fieldList[i].TTL = int64(ttl)
			}
		}
	}

	result := HashScanResult{Fields: fieldList, Total: int(total), Cursor: nextCursor}
	data, _ := json.Marshal(result)
	return string(data)
}

// HashSet sets a hash field value
func (rs *RedisService) HashSet(connID, key, field, value string) error {
	client, err := rs.getClient(connID)
	if err != nil {
		return err
	}
	return client.HSet(context.Background(), key, field, value).Err()
}

// HashDel deletes a hash field
func (rs *RedisService) HashDel(connID, key, field string) error {
	client, err := rs.getClient(connID)
	if err != nil {
		return err
	}
	return client.HDel(context.Background(), key, field).Err()
}

// --- List Operations ---

// ListRange retrieves a range of list elements
func (rs *RedisService) ListRange(connID, key string, start, stop int64) string {
	client, err := rs.getClient(connID)
	if err != nil {
		log.Printf("[Redis] ListRange: conn=%s, key=%s, err=%v", connID, key, err)
		return "[]"
	}
	items, err := client.LRange(context.Background(), key, start, stop).Result()
	if err != nil {
		log.Printf("[Redis] ListRange: lrange failed, conn=%s, key=%s, err=%v", connID, key, err)
		return "[]"
	}
	if items == nil {
		items = []string{}
	}
	log.Printf("[Redis] ListRange: conn=%s, key=%s, range=[%d,%d], got=%d items", connID, key, start, stop, len(items))
	data, _ := json.Marshal(items)
	return string(data)
}

// ListPush pushes elements to the list
func (rs *RedisService) ListPush(connID, key, value string, left bool) int64 {
	client, err := rs.getClient(connID)
	if err != nil {
		return 0
	}
	ctx := context.Background()
	if left {
		count, err := client.LPush(ctx, key, value).Result()
		if err != nil {
			return 0
		}
		return count
	}
	count, err := client.RPush(ctx, key, value).Result()
	if err != nil {
		return 0
	}
	return count
}

// ListRemove removes elements from the list
func (rs *RedisService) ListRemove(connID, key, value string, count int64) int64 {
	client, err := rs.getClient(connID)
	if err != nil {
		return 0
	}
	n, err := client.LRem(context.Background(), key, count, value).Result()
	if err != nil {
		return 0
	}
	return n
}

// --- Set Operations ---

// SetMembers retrieves set members with pagination
func (rs *RedisService) SetMembers(connID, key string, cursor uint64, count int64) string {
	client, err := rs.getClient(connID)
	if err != nil {
		log.Printf("[Redis] SetMembers: conn=%s, key=%s, err=%v", connID, key, err)
		return fmt.Sprintf(`{"error":"%s"}`, err.Error())
	}
	if count <= 0 {
		count = 500
	}

	ctx := context.Background()
	total, _ := client.SCard(ctx, key).Result()
	members, nextCursor, err := client.SScan(ctx, key, cursor, "*", count).Result()
	if err != nil {
		log.Printf("[Redis] SetMembers: sscan failed, conn=%s, key=%s, err=%v", connID, key, err)
		return fmt.Sprintf(`{"error":"%s"}`, err.Error())
	}
	if members == nil {
		members = []string{}
	}

	log.Printf("[Redis] SetMembers: conn=%s, key=%s, total=%d, fetched=%d", connID, key, total, len(members))
	result := SetScanResult{Members: members, Total: int(total), Cursor: nextCursor}
	data, _ := json.Marshal(result)
	return string(data)
}

// SetAdd adds a member to a set
func (rs *RedisService) SetAdd(connID, key, member string) error {
	client, err := rs.getClient(connID)
	if err != nil {
		return err
	}
	return client.SAdd(context.Background(), key, member).Err()
}

// SetRemove removes a member from a set
func (rs *RedisService) SetRemove(connID, key, member string) error {
	client, err := rs.getClient(connID)
	if err != nil {
		return err
	}
	return client.SRem(context.Background(), key, member).Err()
}

// --- ZSet Operations ---

// ZSetScan scans sorted set entries with pagination
func (rs *RedisService) ZSetScan(connID, key string, cursor uint64, count int64) string {
	client, err := rs.getClient(connID)
	if err != nil {
		log.Printf("[Redis] ZSetScan: conn=%s, key=%s, err=%v", connID, key, err)
		return fmt.Sprintf(`{"error":"%s"}`, err.Error())
	}
	if count <= 0 {
		count = 200
	}

	ctx := context.Background()
	total, _ := client.ZCard(ctx, key).Result()
	entries, nextCursor, err := client.ZScan(ctx, key, cursor, "*", count).Result()
	if err != nil {
		log.Printf("[Redis] ZSetScan: zscan failed, conn=%s, key=%s, err=%v", connID, key, err)
		return fmt.Sprintf(`{"error":"%s"}`, err.Error())
	}

	log.Printf("[Redis] ZSetScan: conn=%s, key=%s, total=%d, fetched=%d", connID, key, total, len(entries)/2)

	var zsetEntries []ZSetEntry
	for i := 0; i < len(entries); i += 2 {
		score, _ := strconv.ParseFloat(entries[i+1], 64)
		zsetEntries = append(zsetEntries, ZSetEntry{Member: entries[i], Score: score})
	}

	result := ZSetScanResult{Entries: zsetEntries, Total: int(total), Cursor: nextCursor}
	data, _ := json.Marshal(result)
	return string(data)
}

// ZSetAdd adds an entry to a sorted set
func (rs *RedisService) ZSetAdd(connID, key, member string, score float64) error {
	client, err := rs.getClient(connID)
	if err != nil {
		return err
	}
	return client.ZAdd(context.Background(), key, redis.Z{Score: score, Member: member}).Err()
}

// ZSetRemove removes a member from a sorted set
func (rs *RedisService) ZSetRemove(connID, key, member string) error {
	client, err := rs.getClient(connID)
	if err != nil {
		return err
	}
	return client.ZRem(context.Background(), key, member).Err()
}
