package config

import (
	"os"
	"strings"
	"time"
)

func MigrateAPIKeysToV2(cfg *Config) bool {
	if len(cfg.Keys) == 0 {
		return false
	}

	// Idempotent migration: allow reruns to fill missing APIKeys after partial migration.
	if len(cfg.APIKeys) >= len(cfg.Keys) {
		return false
	}

	Logger.Info("[migration] migrating API keys from v1 to v2 format")

	now := time.Now()
	apiKeys := make([]APIKeyMetadata, 0, len(cfg.Keys))
	existing := make(map[string]struct{}, len(cfg.APIKeys))
	for _, metadata := range cfg.APIKeys {
		existing[metadata.Key] = struct{}{}
		apiKeys = append(apiKeys, metadata)
	}

	for _, key := range cfg.Keys {
		if strings.TrimSpace(key) == "" {
			continue
		}
		if _, ok := existing[key]; ok {
			continue
		}
		metadata := APIKeyMetadata{
			ID:        generateAPIKeyID(key),
			Key:       key,
			CreatedAt: now,
			ExpiresAt: APIKeyExpiryFrom(now),
		}
		apiKeys = append(apiKeys, metadata)
		existing[key] = struct{}{}
	}

	cfg.APIKeys = apiKeys
	cfg.Keys = nil

	Logger.Info("[migration] migration completed", "migrated_keys", len(apiKeys))

	return true
}

func BackupConfig(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", err
	}

	backupPath := filePath + ".backup." + time.Now().Format("20060102-150405")
	safePerm := fileInfo.Mode().Perm() & 0o600
	if safePerm == 0 {
		safePerm = 0o600
	}
	return backupPath, os.WriteFile(backupPath, data, safePerm)
}

func RestoreConfig(backupPath, targetPath string) error {
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return err
	}

	backupInfo, err := os.Stat(backupPath)
	if err != nil {
		return err
	}

	safePerm := backupInfo.Mode().Perm() & 0o600
	if safePerm == 0 {
		safePerm = 0o600
	}

	return os.WriteFile(targetPath, data, safePerm)
}
