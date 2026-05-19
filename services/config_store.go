package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

// ConnectionConfig represents a saved Redis connection configuration
type ConnectionConfig struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Password    string `json:"password"`
	DB          int    `json:"db"`
	SSHEnabled  bool   `json:"sshEnabled"`
	SSHHost     string `json:"sshHost"`
	SSHPort     int    `json:"sshPort"`
	SSHUser     string `json:"sshUser"`
	SSHKeyFile  string `json:"sshKeyFile"`
	SSHPassword string `json:"sshPassword"`
	Timeout     int    `json:"timeout"`
	Retries     int    `json:"retries"`
}

// ConfigStore manages encrypted connection configuration persistence
type ConfigStore struct {
	configPath string
	encryptKey []byte
}

// NewConfigStore creates a new config store with machine-fingerprint-derived encryption
func NewConfigStore() *ConfigStore {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".onlyRedis")
	os.MkdirAll(configDir, 0700)

	// Derive encryption key from machine fingerprint
	hostname, _ := os.Hostname()
	fingerprint := hostname + runtime.GOOS + os.Getenv("USER")
	keyHash := sha256.Sum256([]byte(fingerprint))

	return &ConfigStore{
		configPath: filepath.Join(configDir, "connections.json"),
		encryptKey: keyHash[:],
	}
}

// LoadConnections reads and decrypts saved connections
func (cs *ConfigStore) LoadConnections() ([]ConnectionConfig, error) {
	data, err := os.ReadFile(cs.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("[Config] LoadConnections: config file not found at %s, returning empty", cs.configPath)
			return []ConnectionConfig{}, nil
		}
		log.Printf("[Config] LoadConnections: read file failed: %v", err)
		return nil, err
	}

	decrypted, err := cs.decrypt(string(data))
	if err != nil {
		// Fall back to reading unencrypted data for migration
		var configs []ConnectionConfig
		if err := json.Unmarshal(data, &configs); err != nil {
			log.Printf("[Config] LoadConnections: parse failed: %v", err)
			return []ConnectionConfig{}, nil
		}
		log.Printf("[Config] LoadConnections: loaded %d connections (unencrypted legacy)", len(configs))
		return configs, nil
	}

	var configs []ConnectionConfig
	if err := json.Unmarshal([]byte(decrypted), &configs); err != nil {
		log.Printf("[Config] LoadConnections: unmarshal failed: %v", err)
		return []ConnectionConfig{}, nil
	}
	log.Printf("[Config] LoadConnections: loaded %d connections from %s", len(configs), cs.configPath)
	return configs, nil
}

// SaveConnections encrypts and persists connection configurations
func (cs *ConfigStore) SaveConnections(configs []ConnectionConfig) error {
	data, err := json.Marshal(configs)
	if err != nil {
		log.Printf("[Config] SaveConnections: marshal failed: %v", err)
		return err
	}

	encrypted, err := cs.encrypt(string(data))
	if err != nil {
		log.Printf("[Config] SaveConnections: encrypt failed: %v", err)
		return err
	}

	err = os.WriteFile(cs.configPath, []byte(encrypted), 0600)
	if err != nil {
		log.Printf("[Config] SaveConnections: write file failed: %v", err)
		return err
	}
	log.Printf("[Config] SaveConnections: saved %d connections to %s", len(configs), cs.configPath)
	return nil
}

// encrypt encrypts plaintext using AES-256-GCM
func (cs *ConfigStore) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(cs.encryptKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decrypts AES-256-GCM encrypted ciphertext
func (cs *ConfigStore) decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(cs.encryptKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", err
	}

	nonce, encryptedData := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// MaskPassword replaces password with masked string for frontend display
func MaskPassword(password string) string {
	if len(password) > 0 {
		return "****"
	}
	return ""
}
