package config

import (
	"crypto/sha256"
	"encoding/hex"
	"slices"
	"time"

	apierrors "ds2api/internal/errors"
)

type APIKeyManager struct {
	store *Store
}

func NewAPIKeyManager(store *Store) *APIKeyManager {
	return &APIKeyManager{store: store}
}

func (m *APIKeyManager) AddAPIKey(key string) error {
	if key == "" {
		return ErrInvalidAPIKey
	}

	metadata := APIKeyMetadata{
		ID:        generateAPIKeyID(key),
		Key:       key,
		CreatedAt: time.Now(),
	}

	return m.store.Update(func(c *Config) error {
		for i, existing := range c.APIKeys {
			if existing.Key == key {
				c.APIKeys[i] = metadata
				return nil
			}
		}
		c.APIKeys = append(c.APIKeys, metadata)
		return nil
	})
}

func (m *APIKeyManager) RemoveAPIKey(key string) error {
	return m.store.Update(func(c *Config) error {
		for i, metadata := range c.APIKeys {
			if metadata.Key == key {
				c.APIKeys = append(c.APIKeys[:i], c.APIKeys[i+1:]...)
				return nil
			}
		}
		return ErrAPIKeyNotFound
	})
}

type KeyFilterFunc func(APIKeyMetadata) bool

func (m *APIKeyManager) filterKeys(filter KeyFilterFunc) []APIKeyMetadata {
	cfg := m.store.Snapshot()
	result := make([]APIKeyMetadata, 0)
	for _, metadata := range cfg.APIKeys {
		if filter(metadata) {
			result = append(result, metadata)
		}
	}
	return result
}

func (m *APIKeyManager) IsAPIKeyValid(key string) bool {
	cfg := m.store.Snapshot()
	for _, metadata := range cfg.APIKeys {
		if metadata.Key == key {
			return IsAPIKeyActiveAt(metadata, time.Time{})
		}
	}

	for _, k := range cfg.Keys {
		if k == key {
			return true
		}
	}

	return false
}

func (m *APIKeyManager) GetAPIKeyMetadata(key string) (APIKeyMetadata, bool) {
	cfg := m.store.Snapshot()
	for _, metadata := range cfg.APIKeys {
		if metadata.Key == key {
			return metadata, true
		}
	}

	return APIKeyMetadata{}, false
}

func (m *APIKeyManager) GetExpiringKeys(daysBefore int) []APIKeyMetadata {
	_ = daysBefore
	return m.filterKeys(func(k APIKeyMetadata) bool {
		_ = k
		return false
	})
}

func (m *APIKeyManager) GetExpiredKeys() []APIKeyMetadata {
	return m.filterKeys(func(k APIKeyMetadata) bool {
		_ = k
		return false
	})
}

func (m *APIKeyManager) CleanExpiredKeys() (int, error) {
	return 0, nil
}

func (m *APIKeyManager) GetAllAPIKeysMetadata() []APIKeyMetadata {
	cfg := m.store.Snapshot()
	return slices.Clone(cfg.APIKeys)
}

func (m *APIKeyManager) GetValidKeys() []string {
	cfg := m.store.Snapshot()
	validKeys := make([]string, 0, len(cfg.APIKeys)+len(cfg.Keys))
	seen := make(map[string]struct{}, len(cfg.APIKeys)+len(cfg.Keys))

	for _, metadata := range cfg.APIKeys {
		if IsAPIKeyActiveAt(metadata, time.Time{}) {
			if _, exists := seen[metadata.Key]; !exists {
				seen[metadata.Key] = struct{}{}
				validKeys = append(validKeys, metadata.Key)
			}
		}
	}

	for _, k := range cfg.Keys {
		if _, exists := seen[k]; !exists {
			seen[k] = struct{}{}
			validKeys = append(validKeys, k)
		}
	}

	return validKeys
}

func (m *APIKeyManager) GetValidAPIKeysMetadata() []APIKeyMetadata {
	cfg := m.store.Snapshot()
	validMetadata := make([]APIKeyMetadata, 0, len(cfg.APIKeys))

	for _, metadata := range cfg.APIKeys {
		if IsAPIKeyActiveAt(metadata, time.Time{}) {
			validMetadata = append(validMetadata, metadata)
		}
	}

	return validMetadata
}

func generateAPIKeyID(key string) string {
	sum := sha256.Sum256([]byte(key))
	return "apikey:" + hex.EncodeToString(sum[:16])
}

var (
	ErrInvalidAPIKey  = apierrors.NewAppError("INVALID_REQUEST", "API key cannot be empty", nil)
	ErrAPIKeyNotFound = apierrors.NewAppError("API_KEY_NOT_FOUND", "API key not found", nil)
	ErrAPIKeyExpired  = apierrors.NewAppError("API_KEY_EXPIRED", "API key has expired", nil)
	ErrAPIKeyExpiring = apierrors.NewAppError("API_KEY_EXPIRING", "API key is expiring soon", nil)
)

func maskAPIKey(key string) string {
	if len(key) <= 17 {
		return "****"
	}
	return key[:11] + "****" + key[len(key)-4:]
}
