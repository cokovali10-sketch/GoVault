package vault

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

type Entry struct {
	Key       string    `json:"key"`
	Value     []byte    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Tags      []string  `json:"tags,omitempty"`
}

type VaultStore struct {
	Entries   map[string]*Entry `json:"entries"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Version   int               `json:"version"`
}

type Vault struct {
	mu       sync.RWMutex
	store    VaultStore
	filePath string
	key      []byte
}

func deriveKey(password string) []byte {
	h := sha256.Sum256([]byte(password))
	return h[:]
}

func NewVault(filePath string, password string) *Vault {
	return &Vault{
		filePath: filePath,
		key:      deriveKey(password),
		store: VaultStore{
			Entries:   make(map[string]*Entry),
			CreatedAt: time.Now(),
			Version:   1,
		},
	}
}

func (v *Vault) encrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(v.key)
	if err != nil {
		return nil, err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return aesGCM.Seal(nonce, nonce, data, nil), nil
}

func (v *Vault) decrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(v.key)
	if err != nil {
		return nil, err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return aesGCM.Open(nil, nonce, ciphertext, nil)
}

func (v *Vault) Set(key string, value []byte, tags ...string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	now := time.Now()
	if existing, ok := v.store.Entries[key]; ok {
		existing.Value = value
		existing.UpdatedAt = now
		if len(tags) > 0 {
			existing.Tags = tags
		}
	} else {
		v.store.Entries[key] = &Entry{
			Key:       key,
			Value:     value,
			CreatedAt: now,
			UpdatedAt: now,
			Tags:      tags,
		}
	}
	v.store.UpdatedAt = now
	return nil
}

func (v *Vault) Get(key string) ([]byte, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	entry, ok := v.store.Entries[key]
	if !ok {
		return nil, errors.New("key not found")
	}
	return entry.Value, nil
}

func (v *Vault) Delete(key string) bool {
	v.mu.Lock()
	defer v.mu.Unlock()

	if _, ok := v.store.Entries[key]; !ok {
		return false
	}
	delete(v.store.Entries, key)
	v.store.UpdatedAt = time.Now()
	return true
}

func (v *Vault) List() []string {
	v.mu.RLock()
	defer v.mu.RUnlock()

	keys := make([]string, 0, len(v.store.Entries))
	for k := range v.store.Entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (v *Vault) Search(query string) []string {
	v.mu.RLock()
	defer v.mu.RUnlock()

	var results []string
	q := query
	for key, entry := range v.store.Entries {
		if key == q {
			results = append(results, key)
			continue
		}
		for _, tag := range entry.Tags {
			if tag == q {
				results = append(results, key)
				break
			}
		}
	}
	sort.Strings(results)
	return results
}

func (v *Vault) SearchPrefix(prefix string) []string {
	v.mu.RLock()
	defer v.mu.RUnlock()

	var results []string
	p := strings.ToLower(prefix)
	for key := range v.store.Entries {
		if strings.HasPrefix(strings.ToLower(key), p) {
			results = append(results, key)
		}
	}
	sort.Strings(results)
	return results
}

func (v *Vault) SearchRegex(pattern string) ([]string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	v.mu.RLock()
	defer v.mu.RUnlock()

	var results []string
	for key := range v.store.Entries {
		if re.MatchString(key) {
			results = append(results, key)
		}
	}
	sort.Strings(results)
	return results, nil
}

func (v *Vault) SearchByTag(tag string) []string {
	v.mu.RLock()
	defer v.mu.RUnlock()

	var results []string
	for key, entry := range v.store.Entries {
		for _, t := range entry.Tags {
			if t == tag {
				results = append(results, key)
				break
			}
		}
	}
	sort.Strings(results)
	return results
}

func (v *Vault) GetEntry(key string) (*Entry, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	entry, ok := v.store.Entries[key]
	if !ok {
		return nil, errors.New("key not found")
	}
	return entry, nil
}

func (v *Vault) Rename(oldKey, newKey string) bool {
	v.mu.Lock()
	defer v.mu.Unlock()

	entry, ok := v.store.Entries[oldKey]
	if !ok {
		return false
	}
	if _, exists := v.store.Entries[newKey]; exists {
		return false
	}

	delete(v.store.Entries, oldKey)
	entry.Key = newKey
	entry.UpdatedAt = time.Now()
	v.store.Entries[newKey] = entry
	v.store.UpdatedAt = time.Now()
	return true
}

func (v *Vault) ChangePassword(newPassword string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.key = deriveKey(newPassword)
	return nil
}

func (v *Vault) RotatePassword(oldPassword, newPassword string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if !bytes.Equal(v.key, deriveKey(oldPassword)) {
		return errors.New("old password verification failed")
	}

	v.key = deriveKey(newPassword)
	return nil
}

func (v *Vault) ExportJSON() ([]byte, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	export := make(map[string]interface{})
	for k, e := range v.store.Entries {
		export[k] = map[string]interface{}{
			"value":      string(e.Value),
			"created_at": e.CreatedAt,
			"updated_at": e.UpdatedAt,
			"tags":       e.Tags,
		}
	}
	return json.MarshalIndent(export, "", "  ")
}

func (v *Vault) ImportJSON(data []byte) error {
	var imported map[string]interface{}
	if err := json.Unmarshal(data, &imported); err != nil {
		return err
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	for k, raw := range imported {
		item, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		entry := &Entry{
			Key: k,
			Value: []byte(func() string {
				if s, ok := item["value"].(string); ok {
					return s
				}
				return ""
			}()),
			Tags: func() []string {
				if tags, ok := item["tags"].([]interface{}); ok {
					var result []string
					for _, t := range tags {
						if s, ok := t.(string); ok {
							result = append(result, s)
						}
					}
					return result
				}
				return nil
			}(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		v.store.Entries[k] = entry
	}
	v.store.UpdatedAt = time.Now()
	return nil
}

func (v *Vault) Backup(backupPath string) error {
	v.mu.RLock()
	data, err := json.Marshal(v.store)
	v.mu.RUnlock()
	if err != nil {
		return err
	}

	encrypted, err := v.encrypt(data)
	if err != nil {
		return err
	}

	return os.WriteFile(backupPath, encrypted, 0600)
}

func (v *Vault) Save() error {
	v.mu.RLock()
	data, err := json.Marshal(v.store)
	v.mu.RUnlock()
	if err != nil {
		return err
	}

	encrypted, err := v.encrypt(data)
	if err != nil {
		return err
	}

	return os.WriteFile(v.filePath, encrypted, 0600)
}

func (v *Vault) Load() error {
	encrypted, err := os.ReadFile(v.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	data, err := v.decrypt(encrypted)
	if err != nil {
		return errors.New("decryption failed (wrong password?)")
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	return json.Unmarshal(data, &v.store)
}

func (v *Vault) Stats() map[string]interface{} {
	v.mu.RLock()
	defer v.mu.RUnlock()

	totalSize := 0
	tagCounts := make(map[string]int)
	for _, e := range v.store.Entries {
		totalSize += len(e.Value)
		for _, t := range e.Tags {
			tagCounts[t]++
		}
	}

	return map[string]interface{}{
		"entries":     len(v.store.Entries),
		"total_size":  totalSize,
		"created_at":  v.store.CreatedAt,
		"updated_at":  v.store.UpdatedAt,
		"version":     v.store.Version,
		"unique_tags": len(tagCounts),
		"tag_counts":  tagCounts,
	}
}

func (v *Vault) Exists(key string) bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	_, ok := v.store.Entries[key]
	return ok
}

func (v *Vault) Count() int {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return len(v.store.Entries)
}
