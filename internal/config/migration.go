package config

import (
	"os"
	"time"
)

func MigrateAPIKeysToV2(cfg *Config) bool {
	if len(cfg.APIKeys) > 0 {
		return false
	}

	if len(cfg.Keys) == 0 {
		return false
	}

	Logger.Info("[migration] migrating API keys from v1 to v2 format")

	now := time.Now()
	apiKeys := make([]APIKeyMetadata, 0, len(cfg.Keys))

	for _, key := range cfg.Keys {
		metadata := APIKeyMetadata{
			ID:        generateAPIKeyID(key),
			Key:       key,
			CreatedAt: now,
			ExpiresAt: now.Add(APIKeyTTL),
		}
		apiKeys = append(apiKeys, metadata)
	}

	cfg.APIKeys = apiKeys

	Logger.Info("[migration] migration completed", "migrated_keys", len(apiKeys))

	return true
}

func BackupConfig(filePath string) error {
	backupPath := filePath + ".backup." + time.Now().Format("20060102-150405")
	
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	return os.WriteFile(backupPath, data, 0644)
}

func RestoreConfig(backupPath, targetPath string) error {
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return err
	}

	return os.WriteFile(targetPath, data, 0644)
}
