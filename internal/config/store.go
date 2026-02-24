package config

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"slices"
	"strings"
	"sync"
)

type Store struct {
	mu         sync.RWMutex
	cfg        Config
	path       string
	fromEnv    bool
	keyMap     map[string]struct{}       // O(1) legacy API key lookup index
	keyMetaMap map[string]APIKeyMetadata // O(1) API key metadata lookup
	accMap     map[string]int            // O(1) account lookup: identifier -> slice index
}

func NewStore(cfg *Config, path string) *Store {
	storePath := path
	if strings.TrimSpace(storePath) == "" {
		storePath = ConfigPath()
	}

	var initial Config
	if cfg != nil {
		initial = cfg.Clone()
	}

	s := &Store{cfg: initial, path: storePath}
	s.rebuildIndexes()
	return s
}

func LoadStore() *Store {
	cfg, fromEnv, err := loadConfig()
	if err != nil {
		Logger.Warn("[config] load failed", "error", err)
	}
	if len(cfg.Keys) == 0 && len(cfg.Accounts) == 0 {
		Logger.Warn("[config] empty config loaded")
	}

	if len(cfg.Keys) > 0 && len(cfg.APIKeys) == 0 && !fromEnv {
		tempPath := ConfigPath() + ".tmp"

		backupPath, err := BackupConfig(ConfigPath())
		if err != nil {
			Logger.Warn("[config] backup failed", "error", err)
			Logger.Warn("[config] aborting migration due to backup failure")
		} else {
			Logger.Info("[config] config backed up", "path", backupPath)

			originalCfg := cfg.Clone()
			if MigrateAPIKeysToV2(&cfg) {
				if err := SaveConfigToPath(&cfg, tempPath); err != nil {
					Logger.Error("[config] failed to write migrated config to temp file", "error", err)
					Logger.Warn("[config] cleaning up temp file and aborting migration")
					os.Remove(tempPath)
					cfg = originalCfg
				} else {
					if err := os.Rename(tempPath, ConfigPath()); err != nil {
						Logger.Error("[config] failed to rename temp file to config", "error", err)
						Logger.Warn("[config] restoring config from backup")
						if restoreErr := RestoreConfig(backupPath, ConfigPath()); restoreErr != nil {
							Logger.Error("[config] failed to restore config from backup", "error", restoreErr)
						}
						os.Remove(tempPath)
						cfg = originalCfg
					} else {
						Logger.Info("[config] migration completed successfully")
					}
				}
			}
		}
	}

	s := &Store{cfg: cfg, path: ConfigPath(), fromEnv: fromEnv}
	s.rebuildIndexes()
	return s
}

func loadConfig() (Config, bool, error) {
	rawCfg := strings.TrimSpace(os.Getenv("DS2API_CONFIG_JSON"))
	if rawCfg == "" {
		rawCfg = strings.TrimSpace(os.Getenv("CONFIG_JSON"))
	}
	if rawCfg != "" {
		cfg, err := parseConfigString(rawCfg)
		return cfg, true, err
	}

	content, err := os.ReadFile(ConfigPath())
	if err != nil {
		if IsVercel() {
			// Vercel one-click deploy may start without a writable/present config file.
			// Keep an in-memory config so users can bootstrap via WebUI then sync env.
			return Config{}, true, nil
		}
		return Config{}, false, err
	}
	var cfg Config
	if err := json.Unmarshal(content, &cfg); err != nil {
		return Config{}, false, err
	}
	if IsVercel() {
		// Vercel filesystem is ephemeral/read-only for runtime writes; avoid save errors.
		return cfg, true, nil
	}
	return cfg, false, nil
}

func (s *Store) Snapshot() Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg.Clone()
}

func (s *Store) findAPIKeyMetadataLocked(k string) (APIKeyMetadata, bool) {
	metadata, ok := s.keyMetaMap[k]
	return metadata, ok
}

func (s *Store) HasAPIKey(k string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, found := s.findAPIKeyMetadataLocked(k); found {
		return true
	}
	_, ok := s.keyMap[k]
	return ok
}

func (s *Store) HasValidAPIKey(k string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, found := s.findAPIKeyMetadataLocked(k); found {
		return true
	}
	_, ok := s.keyMap[k]
	return ok
}

func (s *Store) IsAPIKeyExpired(k string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return false
}

func (s *Store) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return slices.Clone(s.cfg.Keys)
}

func (s *Store) Accounts() []Account {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return slices.Clone(s.cfg.Accounts)
}

func (s *Store) FindAccount(identifier string) (Account, bool) {
	identifier = strings.TrimSpace(identifier)
	s.mu.RLock()
	defer s.mu.RUnlock()
	if idx, ok := s.findAccountIndexLocked(identifier); ok {
		return s.cfg.Accounts[idx], true
	}
	return Account{}, false
}

func (s *Store) UpdateAccountToken(identifier, token string) error {
	identifier = strings.TrimSpace(identifier)
	s.mu.Lock()
	defer s.mu.Unlock()
	idx, ok := s.findAccountIndexLocked(identifier)
	if !ok {
		return errors.New("account not found")
	}
	oldID := s.cfg.Accounts[idx].Identifier()
	s.cfg.Accounts[idx].Token = token
	newID := s.cfg.Accounts[idx].Identifier()
	// Keep historical aliases usable for long-lived queues while also adding
	// the latest identifier after token refresh.
	if identifier != "" {
		s.accMap[identifier] = idx
	}
	if oldID != "" {
		s.accMap[oldID] = idx
	}
	if newID != "" {
		s.accMap[newID] = idx
	}
	return s.saveLocked()
}

func (s *Store) Replace(cfg Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg = cfg.Clone()
	s.rebuildIndexes()
	return s.saveLocked()
}

func (s *Store) Update(mutator func(*Config) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cfg := s.cfg.Clone()
	if err := mutator(&cfg); err != nil {
		return err
	}
	s.cfg = cfg
	s.rebuildIndexes()
	return s.saveLocked()
}

func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.fromEnv {
		Logger.Info("[save_config] source from env, skip write")
		return nil
	}
	b, err := json.MarshalIndent(s.cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, b, 0o600)
}

func (s *Store) saveLocked() error {
	if s.fromEnv {
		Logger.Info("[save_config] source from env, skip write")
		return nil
	}
	b, err := json.MarshalIndent(s.cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, b, 0o600)
}

func SaveConfigToPath(cfg *Config, path string) error {
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}

func (s *Store) IsEnvBacked() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.fromEnv
}

func (s *Store) SetVercelSync(hash string, ts int64) error {
	return s.Update(func(c *Config) error {
		c.VercelSyncHash = hash
		c.VercelSyncTime = ts
		return nil
	})
}

func (s *Store) ExportJSONAndBase64() (string, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, err := json.Marshal(s.cfg)
	if err != nil {
		return "", "", err
	}
	return string(b), base64.StdEncoding.EncodeToString(b), nil
}
