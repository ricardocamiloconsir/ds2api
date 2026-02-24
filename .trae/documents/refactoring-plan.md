# Refactoring Plan - ds2api Codebase

## Overview
This document outlines a comprehensive refactoring plan to improve code quality, maintainability, and performance following Clean Code principles, DRY, Object Calisthenics, and best practices for Go and React/JavaScript.

---

## Phase 1: Go Backend Refactoring

### 1.1 Config Package - Store Duplication Elimination

**File:** `internal/config/store.go`

**Issues:**
- `Save()` and `saveLocked()` methods are nearly identical (violation of DRY)
- `HasAPIKey()` and `HasValidAPIKey()` have duplicated iteration logic

**Refactoring:**
1. Extract common save logic into a single private method
2. Consolidate API key validation logic into a single reusable method
3. Extract time.Now() calls before lock acquisition where appropriate

**Changes:**
```go
// Replace Save() and saveLocked() with:
func (s *Store) saveLocked() error {
    if s.fromEnv {
        Logger.Info("[save_config] source from env, skip write")
        return nil
    }
    b, err := json.MarshalIndent(s.cfg, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(s.path, b, 0o644)
}

func (s *Store) Save() error {
    s.mu.Lock()
    defer s.mu.Unlock()
    return s.saveLocked()
}

// Consolidate API key lookup:
func (s *Store) findAPIKeyMetadata(key string) (APIKeyMetadata, bool) {
    for _, metadata := range s.cfg.APIKeys {
        if metadata.Key == key {
            return metadata, true
        }
    }
    return APIKeyMetadata{}, false
}
```

---

### 1.2 API Key Manager - Filter Pattern Extraction

**File:** `internal/config/apikey_manager.go`

**Issues:**
- Multiple methods (`GetExpiringKeys`, `GetExpiredKeys`, `CleanExpiredKeys`, `GetValidKeys`) repeat similar filtering logic
- Duplicate snapshot acquisition and iteration patterns

**Refactoring:**
1. Create a generic filter function for API keys based on predicates
2. Extract common snapshot acquisition pattern
3. Consolidate validation logic

**Changes:**
```go
type KeyFilterFunc func(APIKeyMetadata) bool

func (m *APIKeyManager) filterKeys(filter KeyFilterFunc) []APIKeyMetadata {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    cfg := m.store.Snapshot()
    result := make([]APIKeyMetadata, 0)
    for _, metadata := range cfg.APIKeys {
        if filter(metadata) {
            result = append(result, metadata)
        }
    }
    return result
}

// Then simplify methods:
func (m *APIKeyManager) GetExpiringKeys(daysBefore int) []APIKeyMetadata {
    now := time.Now()
    threshold := now.Add(time.Duration(daysBefore) * 24 * time.Hour)
    return m.filterKeys(func(k APIKeyMetadata) bool {
        return k.ExpiresAt.After(now) && k.ExpiresAt.Before(threshold)
    })
}
```

---

### 1.3 Admin Handlers - Error Handling Consolidation

**File:** `internal/admin/handler.go`

**Issues:**
- Repetitive JSON encoding error handling in multiple handlers
- Duplicate service availability checks
- Repeated SSE write error handling

**Refactoring:**
1. Create helper functions for common response patterns
2. Extract service availability validation
3. Create SSE writer helper with error handling

**Changes:**
```go
func (h *Handler) writeJSONResponse(w http.ResponseWriter, payload any) {
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(payload); err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
    }
}

func (h *Handler) requireService(service, serviceName string, w http.ResponseWriter) bool {
    if service == "" || service == nil {
        http.Error(w, serviceName+" not available", http.StatusServiceUnavailable)
        return false
    }
    return true
}

func (h *Handler) writeSSEData(w http.ResponseWriter, flusher http.Flusher, data []byte) bool {
    if _, err := w.Write([]byte("data: ")); err != nil {
        config.Logger.Error("[sse] failed to write prefix", "error", err)
        return false
    }
    if _, err := w.Write(data); err != nil {
        config.Logger.Error("[sse] failed to write data", "error", err)
        return false
    }
    if _, err := w.Write([]byte("\n\n")); err != nil {
        config.Logger.Error("[sse] failed to write newline", "error", err)
        return false
    }
    flusher.Flush()
    return true
}
```

---

### 1.4 Admin Handlers - Account CRUD Simplification

**Files:** `internal/admin/handler_accounts_crud.go`, `internal/admin/helpers.go`

**Issues:**
- Repetitive account validation logic
- Duplicate account lookup patterns
- Verbose debug logging throughout
- Chinese error messages mixed with English code

**Refactoring:**
1. Extract account lookup into a reusable function
2. Create validation helper functions
3. Centralize error message constants
4. Remove verbose debug logging or make it optional

**Changes:**
```go
// Add to helpers.go:
const (
    ErrAccountNotFound = "账号不存在"
    ErrEmailExists    = "邮箱已存在"
    ErrMobileExists   = "手机号已存在"
    ErrInvalidRequest = "需要 email 或 mobile"
)

func findAccountIndex(accounts []config.Account, identifier string) (int, bool) {
    for i, acc := range accounts {
        if accountMatchesIdentifier(acc, identifier) {
            return i, true
        }
    }
    return -1, false
}

func validateAccountUpdate(newAcc, existing config.Account) error {
    if newAcc.Email != "" && existing.Email != "" && newAcc.Email != existing.Email {
        return fmt.Errorf(ErrEmailExists)
    }
    if newAcc.Mobile != "" && existing.Mobile != "" && newAcc.Mobile != existing.Mobile {
        return fmt.Errorf(ErrMobileExists)
    }
    return nil
}
```

---

### 1.5 Monitor Package - Notification Broadcasting

**Files:** `internal/monitor/notifier.go`, `internal/monitor/monitor.go`

**Issues:**
- `notifyExpiring` and `notifyExpired` have identical structure (violation of DRY)
- Magic numbers (100 history limit, 10 channel buffer) should be constants

**Refactoring:**
1. Create generic notification creation method
2. Extract notification constants
3. Consolidate broadcasting logic

**Changes:**
```go
const (
    DefaultMaxHistory    = 100
    NotificationBufferSize = 10
)

func (n *Notifier) createNotification(key config.APIKeyMetadata, notificationType NotificationType, message string) Notification {
    return Notification{
        Type:      notificationType,
        APIKey:    maskAPIKey(key.Key),
        Message:   message,
        ExpiresAt: key.ExpiresAt,
        Timestamp: time.Now(),
    }
}

func (n *Notifier) notifyKeys(keys []config.APIKeyMetadata, notificationType NotificationType, message string) {
    n.mu.Lock()
    defer n.mu.Unlock()
    
    for _, key := range keys {
        notification := n.createNotification(key, notificationType, message)
        n.addToHistory(notification)
        n.broadcast(notification)
    }
}
```

---

### 1.6 Config Package - Accessor Pattern Consolidation

**File:** `internal/config/store_accessors.go`

**Issues:**
- Repeated pattern of acquiring lock, reading value, returning default
- Duplicate environment variable lookup logic

**Refactoring:**
1. Create generic accessor with default value provider
2. Extract environment variable parsing logic
3. Use consistent default value patterns

**Changes:**
```go
func (s *Store) getEnvInt(keys []string, defaultValue int) int {
    for _, key := range keys {
        raw := strings.TrimSpace(os.Getenv(key))
        if raw == "" {
            continue
        }
        if n, err := strconv.Atoi(raw); err == nil && n > 0 {
            return n
        }
    }
    return defaultValue
}

func (s *Store) getEnvIntWithFallback(keys []string, defaultValue int, allowZero bool) int {
    for _, key := range keys {
        raw := strings.TrimSpace(os.Getenv(key))
        if raw == "" {
            continue
        }
        if n, err := strconv.Atoi(raw); err == nil {
            if n > 0 || (allowZero && n >= 0) {
                return n
            }
        }
    }
    return defaultValue
}
```

---

## Phase 2: Frontend Refactoring

### 2.1 API Key Expiry Logic Consolidation

**Files:** `webui/src/features/account/ApiKeyExpiryPanel.jsx`, `webui/src/features/account/ApiKeysPanel.jsx`

**Issues:**
- Duplicate key expiry status calculation logic
- Repeated date calculation patterns
- Similar key display components

**Refactoring:**
1. Create shared utility functions for expiry calculations
2. Extract common key display component
3. Create constants for expiry thresholds

**Changes:**
```javascript
// webui/src/utils/apiKeyUtils.js
export const EXPIRY_THRESHOLDS = {
    CRITICAL: 0,
    WARNING: 7,
}

export const calculateDaysUntilExpiry = (expiresAt) => {
    const now = new Date()
    const expires = new Date(expiresAt)
    return Math.ceil((expires - now) / (1000 * 60 * 60 * 24))
}

export const getKeyExpiryStatus = (expiresAt) => {
    const daysLeft = calculateDaysUntilExpiry(expiresAt)
    if (daysLeft <= EXPIRY_THRESHOLDS.CRITICAL) {
        return { status: 'expired', daysLeft }
    }
    if (daysLeft <= EXPIRY_THRESHOLDS.WARNING) {
        return { status: 'expiring', daysLeft }
    }
    return { status: 'valid', daysLeft }
}

export const maskApiKey = (key) => {
    if (!key || key.length <= 16) return '****'
    return key.slice(0, 16) + '****'
}
```

---

### 2.2 Account Identifier Resolution

**Files:** 
- `webui/src/features/account/useAccountsData.js`
- `webui/src/features/account/useAccountActions.js`
- `webui/src/features/apiTester/ApiTesterContainer.jsx`

**Issues:**
- Duplicate `resolveAccountIdentifier` function in multiple files
- Inconsistent handling of null/undefined accounts

**Refactoring:**
1. Create shared utility module for account helpers
2. Centralize identifier resolution logic

**Changes:**
```javascript
// webui/src/utils/accountUtils.js
export const resolveAccountIdentifier = (acc) => {
    if (!acc || typeof acc !== 'object') return ''
    return String(acc.identifier || acc.email || acc.mobile || '').trim()
}

export const formatAccountDisplay = (acc) => {
    const identifier = resolveAccountIdentifier(acc)
    if (acc.email) return acc.email
    if (acc.mobile) return acc.mobile
    return identifier
}
```

---

### 2.3 API Fetch Pattern Consolidation

**Files:** 
- `webui/src/features/account/useApiKeyExpiry.js`
- `webui/src/features/account/useAccountsData.js`
- `webui/src/features/settings/settingsApi.js`

**Issues:**
- Repetitive fetch patterns with try-catch
- Duplicate error handling
- Similar state management for loading

**Refactoring:**
1. Create reusable API client wrapper
2. Extract common error handling
3. Create custom hooks for common patterns

**Changes:**
```javascript
// webui/src/utils/apiClient.js
export const createApiClient = (baseFetch) => {
    return {
        async get(endpoint) {
            const res = await baseFetch(endpoint)
            if (!res.ok) throw new Error(`GET ${endpoint} failed`)
            return res.json()
        },
        async post(endpoint, body) {
            const res = await baseFetch(endpoint, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(body),
            })
            if (!res.ok) throw new Error(`POST ${endpoint} failed`)
            return res.json()
        },
        async put(endpoint, body) {
            const res = await baseFetch(endpoint, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(body),
            })
            if (!res.ok) throw new Error(`PUT ${endpoint} failed`)
            return res.json()
        },
        async delete(endpoint) {
            const res = await baseFetch(endpoint, { method: 'DELETE' })
            if (!res.ok) throw new Error(`DELETE ${endpoint} failed`)
            return res.json()
        },
    }
}

// webui/src/hooks/useApiData.js
export const useApiData = (apiClient, endpoint, options = {}) => {
    const [data, setData] = useState(null)
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState(null)

    const fetchData = useCallback(async () => {
        setLoading(true)
        setError(null)
        try {
            const result = await apiClient.get(endpoint)
            setData(result)
        } catch (err) {
            setError(err)
            console.error(`Failed to fetch ${endpoint}:`, err)
        } finally {
            setLoading(false)
        }
    }, [apiClient, endpoint])

    useEffect(() => {
        if (options.autoFetch !== false) {
            fetchData()
        }
    }, [fetchData, options.autoFetch])

    return { data, loading, error, refetch: fetchData }
}
```

---

### 2.4 Component Prop Drilling Reduction

**File:** `webui/src/features/settings/SettingsContainer.jsx`

**Issues:**
- Too many props passed to child components
- Settings form state could be better encapsulated

**Refactoring:**
1. Create context for settings management
2. Consolidate related state into custom hooks
3. Reduce prop passing through context

**Changes:**
```javascript
// webui/src/contexts/SettingsContext.jsx
const SettingsContext = createContext()

export const SettingsProvider = ({ children, apiFetch, onRefresh, onMessage }) => {
    const settingsForm = useSettingsForm({
        apiFetch,
        t: useI18n().t,
        onMessage,
        onRefresh,
    })

    return (
        <SettingsContext.Provider value={settingsForm}>
            {children}
        </SettingsContext.Provider>
    )
}

export const useSettings = () => {
    const context = useContext(SettingsContext)
    if (!context) throw new Error('useSettings must be used within SettingsProvider')
    return context
}
```

---

### 2.5 Notification State Management

**File:** `webui/src/features/account/useApiKeyExpiry.js`

**Issues:**
- SSE connection logic could be extracted
- Retry logic is embedded in the hook
- Notification filtering is simple but could be more robust

**Refactoring:**
1. Create reusable SSE hook
2. Extract notification management logic
3. Improve retry strategy

**Changes:**
```javascript
// webui/src/hooks/useSSE.js
export const useSSE = (endpoint, options = {}) => {
    const { maxRetries = 5, retryDelay = 5000, onMessage, onError } = options
    const [connected, setConnected] = useState(false)
    const eventSourceRef = useRef(null)
    const retryCountRef = useRef(0)

    const connect = useCallback(() => {
        eventSourceRef.current = new EventSource(endpoint)
        eventSourceRef.current.onopen = () => {
            setConnected(true)
            retryCountRef.current = 0
        }
        eventSourceRef.current.onmessage = (event) => {
            if (onMessage) onMessage(event)
        }
        eventSourceRef.current.onerror = () => {
            setConnected(false)
            if (onError) onError()
            eventSourceRef.current?.close()
            retryCountRef.current++
            if (retryCountRef.current <= maxRetries) {
                setTimeout(connect, retryDelay * retryCountRef.current)
            }
        }
    }, [endpoint, maxRetries, retryDelay, onMessage, onError])

    useEffect(() => {
        connect()
        return () => {
            eventSourceRef.current?.close()
        }
    }, [connect])

    return { connected }
}
```

---

## Phase 3: Performance Optimizations

### 3.1 Go Backend - Reduce Lock Contention

**Files:** Various in `internal/config`, `internal/monitor`

**Issues:**
- Multiple lock acquisitions for related operations
- Snapshot taken repeatedly for multiple queries

**Refactoring:**
1. Batch related operations under single lock
2. Use read locks more aggressively where possible
3. Consider using sync.RWMutex more efficiently

**Changes:**
```go
// In APIKeyManager, batch operations:
func (m *APIKeyManager) GetStatus(daysBefore int) KeyStatus {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    cfg := m.store.Snapshot()
    now := time.Now()
    threshold := now.Add(time.Duration(daysBefore) * 24 * time.Hour)
    
    status := KeyStatus{
        Total:    len(cfg.APIKeys),
        Valid:    0,
        Expiring: 0,
        Expired:  0,
    }
    
    for _, metadata := range cfg.APIKeys {
        if metadata.ExpiresAt.Before(now) {
            status.Expired++
        } else if metadata.ExpiresAt.Before(threshold) {
            status.Expiring++
        } else {
            status.Valid++
        }
    }
    
    return status
}
```

---

### 3.2 Frontend - Reduce Re-renders

**Files:** Various React components

**Issues:**
- Inline function definitions in JSX
- Unnecessary object/array recreations
- Missing memoization for expensive operations

**Refactoring:**
1. Use `useCallback` and `useMemo` appropriately
2. Extract inline functions
3. Use React.memo for pure components

**Changes:**
```javascript
// Memoize expensive calculations
const validKeys = useMemo(() => 
    apiKeysMetadata.filter(k => getKeyExpiryStatus(k.expires_at).status === 'valid'),
    [apiKeysMetadata]
)

// Memoize callbacks
const handleRefresh = useCallback(async () => {
    setLoading(true)
    try {
        await onRefresh()
    } finally {
        setLoading(false)
    }
}, [onRefresh])

// Memoize pure components
export const KeyStatusBadge = React.memo(({ daysLeft, t }) => {
    const status = getKeyExpiryStatus(daysLeft)
    return (
        <span className={`badge ${status.color}`}>
            {status.text}
        </span>
    )
})
```

---

### 3.3 Go Backend - String Allocations

**Files:** Various handlers and utilities

**Issues:**
- Repeated string trimming operations
- Unnecessary string allocations

**Refactoring:**
1. Trim strings once and reuse
2. Use string builders for concatenation
3. Avoid unnecessary string conversions

**Changes:**
```go
// In helpers.go, create trim helper:
func trimString(s string) string {
    return strings.TrimSpace(s)
}

// Use consistently:
func (s *Store) AdminPasswordHash() string {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return trimString(s.cfg.Admin.PasswordHash)
}
```

---

## Phase 4: Naming and Clarity Improvements

### 4.1 Variable Naming

**Issues:**
- Short, unclear names like `acc`, `cfg`, `req`
- Inconsistent naming conventions

**Refactoring:**
1. Use descriptive names that explain purpose
2. Be consistent across codebase
3. Avoid abbreviations unless widely understood

**Changes:**
```go
// Before:
func (h *Handler) updateAccount(w http.ResponseWriter, r *http.Request) {
    identifier := chi.URLParam(r, "identifier")
    var req map[string]any
    _ = json.NewDecoder(r.Body).Decode(&req)
    updatedAcc := toAccount(req)
    // ...
}

// After:
func (h *Handler) updateAccount(w http.ResponseWriter, request *http.Request) {
    accountIdentifier := chi.URLParam(request, "identifier")
    var requestBody map[string]any
    _ = json.NewDecoder(request.Body).Decode(&requestBody)
    updatedAccount := toAccount(requestBody)
    // ...
}
```

---

### 4.2 Function Naming

**Issues:**
- Generic names like `handle`, `process`, `execute`
- Unclear return values
- Inconsistent verb/noun patterns

**Refactoring:**
1. Use descriptive action verbs
2. Make return types obvious from name
3. Be consistent with patterns

**Changes:**
```go
// Before:
func (h *Handler) getAPIKeysMetadata(w http.ResponseWriter, r *http.Request)

// After:
func (h *Handler) handleGetAPIKeysMetadata(w http.ResponseWriter, request *http.Request)
```

---

### 4.3 Constant Extraction

**Issues:**
- Magic numbers scattered throughout code
- Repeated string literals
- Hard-coded values

**Refactoring:**
1. Extract constants to package level
2. Use descriptive constant names
3. Group related constants

**Changes:**
```go
// internal/config/constants.go
package config

const (
    DefaultCheckInterval = 24 * time.Hour
    DefaultWarningDays   = 7
    DefaultMaxHistory    = 100
    NotificationBufferSize = 10
    
    APIKeyTTLDays = 30
    APIKeyTTL     = APIKeyTTLDays * 24 * time.Hour
    
    MaxPageSize = 100
    DefaultPageSize = 10
)

const (
    NotificationTypeWarning NotificationType = "warning"
    NotificationTypeError  NotificationType = "expired"
)

const (
    SSEContentType        = "text/event-stream"
    SSECacheControl      = "no-cache"
    SSEConnection        = "keep-alive"
    SSEAccessControlOrigin = "*"
)
```

---

## Phase 5: Type Safety Improvements

### 5.1 Go - Stronger Types

**Issues:**
- Using `map[string]any` extensively
- Missing type aliases for common types
- Weak typing in some areas

**Refactoring:**
1. Create strong types for domain concepts
2. Use type aliases for common patterns
3. Reduce use of `interface{}`

**Changes:**
```go
type APIKey string
type AccountIdentifier string
type ConfigurationJSON string
type ResponsePayload map[string]any

type APIKeyStatus struct {
    Key       string
    Status    ExpiryStatus
    DaysLeft  int
    ExpiresAt time.Time
}

type ExpiryStatus string

const (
    ExpiryStatusValid   ExpiryStatus = "valid"
    ExpiryStatusExpiring ExpiryStatus = "expiring"
    ExpiryStatusExpired  ExpiryStatus = "expired"
)
```

---

### 5.2 Frontend - PropTypes or TypeScript Migration

**Issues:**
- No type checking on props
- Unclear component interfaces
- Runtime type errors

**Refactoring:**
1. Add PropTypes (short term)
2. Consider TypeScript migration (long term)
3. Document component contracts

**Changes:**
```javascript
// webui/src/components/NotificationBanner.jsx
import PropTypes from 'prop-types'

NotificationBanner.propTypes = {
    notifications: PropTypes.arrayOf(
        PropTypes.shape({
            id: PropTypes.string.isRequired,
            type: PropTypes.oneOf(['warning', 'error', 'info']).isRequired,
            message: PropTypes.string.isRequired,
            apiKey: PropTypes.string,
            dismissed: PropTypes.bool,
        })
    ).isRequired,
    onDismiss: PropTypes.func.isRequired,
}

export default function NotificationBanner({ notifications, onDismiss }) {
    // ...
}
```

---

## Phase 6: Error Handling Standardization

### 6.1 Go Backend - Structured Error Handling

**Issues:**
- Inconsistent error creation
- Missing error wrapping
- Unclear error types

**Refactoring:**
1. Create custom error types
2. Use error wrapping consistently
3. Standardize error responses

**Changes:**
```go
// internal/errors/errors.go
package errors

type AppError struct {
    Code    string
    Message string
    Err     error
}

func (e *AppError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Err)
    }
    return e.Message
}

func (e *AppError) Unwrap() error {
    return e.Err
}

func NewAppError(code, message string, err error) *AppError {
    return &AppError{
        Code:    code,
        Message: message,
        Err:     err,
    }
}

// Predefined errors
var (
    ErrUnauthorized      = NewAppError("UNAUTHORIZED", "unauthorized: missing auth token", nil)
    ErrAccountNotFound  = NewAppError("ACCOUNT_NOT_FOUND", "账号不存在", nil)
    ErrAPIKeyNotFound   = NewAppError("API_KEY_NOT_FOUND", "API key not found", nil)
    ErrInvalidRequest   = NewAppError("INVALID_REQUEST", "invalid request", nil)
)

// HTTP response helper
func WriteErrorResponse(w http.ResponseWriter, err error) {
    var appErr *AppError
    if errors.As(err, &appErr) {
        http.Error(w, appErr.Message, statusCodeForError(appErr.Code))
        return
    }
    http.Error(w, "Internal server error", http.StatusInternalServerError)
}

func statusCodeForError(code string) int {
    switch code {
    case "UNAUTHORIZED":
        return http.StatusUnauthorized
    case "ACCOUNT_NOT_FOUND", "API_KEY_NOT_FOUND":
        return http.StatusNotFound
    case "INVALID_REQUEST":
        return http.StatusBadRequest
    default:
        return http.StatusInternalServerError
    }
}
```

---

### 6.2 Frontend - Error Boundaries

**Issues:**
- No error boundaries
- Unhandled promise rejections
- Poor error UX

**Refactoring:**
1. Add React error boundaries
2. Handle promise rejections globally
3. Improve error display

**Changes:**
```javascript
// webui/src/components/ErrorBoundary.jsx
class ErrorBoundary extends React.Component {
    constructor(props) {
        super(props)
        this.state = { hasError: false, error: null }
    }

    static getDerivedStateFromError(error) {
        return { hasError: true, error }
    }

    componentDidCatch(error, errorInfo) {
        console.error('ErrorBoundary caught:', error, errorInfo)
    }

    render() {
        if (this.state.hasError) {
            return (
                <div className="error-fallback">
                    <h2>Something went wrong</h2>
                    <details>
                        <summary>Error details</summary>
                        <pre>{this.state.error?.toString()}</pre>
                    </details>
                    <button onClick={() => this.setState({ hasError: false, error: null })}>
                        Try again
                    </button>
                </div>
            )
        }
        return this.props.children
    }
}

// Add to main.jsx:
root.render(
    <ErrorBoundary>
        <App />
    </ErrorBoundary>
)
```

---

## Phase 7: Testing Improvements

### 7.1 Go Backend - Test Organization

**Issues:**
- Tests scattered across files
- Missing edge case coverage
- No integration tests

**Refactoring:**
1. Organize tests by feature
2. Add table-driven tests
3. Increase coverage

**Changes:**
```go
// internal/config/apikey_manager_test.go
func TestAPIKeyManager_IsAPIKeyValid(t *testing.T) {
    tests := []struct {
        name     string
        key      string
        expected bool
    }{
        {"valid key", "sk-1234567890", true},
        {"expired key", "sk-expired", false},
        {"non-existent key", "sk-unknown", false},
        {"empty key", "", false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            manager := setupTestManager(t)
            result := manager.IsAPIKeyValid(tt.key)
            if result != tt.expected {
                t.Errorf("IsAPIKeyValid(%q) = %v, want %v", tt.key, result, tt.expected)
            }
        })
    }
}
```

---

### 7.2 Frontend - Test Coverage

**Issues:**
- Minimal or no tests
- No component testing
- No hook testing

**Refactoring:**
1. Add component tests with React Testing Library
2. Add hook tests
3. Add integration tests

**Changes:**
```javascript
// webui/src/features/account/__tests__/ApiKeyExpiryPanel.test.jsx
import { render, screen } from '@testing-library/react'
import ApiKeyExpiryPanel from '../ApiKeyExpiryPanel'
import { useI18n } from '../../../i18n'

jest.mock('../../../i18n')

describe('ApiKeyExpiryPanel', () => {
    beforeEach(() => {
        useI18n.mockReturnValue({ t: (key) => key })
    })

    it('displays valid keys', () => {
        const metadata = [
            { id: '1', key: 'sk-123', expires_at: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString() }
        ]
        render(<ApiKeyExpiryPanel apiKeysMetadata={metadata} />)
        expect(screen.getByText('apiKey.validKeys')).toBeInTheDocument()
    })

    it('displays expiring keys warning', () => {
        const metadata = [
            { id: '1', key: 'sk-123', expires_at: new Date(Date.now() + 5 * 24 * 60 * 60 * 1000).toISOString() }
        ]
        render(<ApiKeyExpiryPanel apiKeysMetadata={metadata} />)
        expect(screen.getByText('apiKey.expiringKeys')).toBeInTheDocument()
    })
})
```

---

## Phase 8: Documentation Improvements

### 8.1 Go - Godoc Comments

**Issues:**
- Missing function documentation
- Unclear package documentation
- No examples

**Refactoring:**
1. Add comprehensive Godoc comments
2. Add usage examples
3. Document exported types

**Changes:**
```go
// Package config provides configuration management for the ds2api application.
// It handles loading, saving, and accessing configuration data with thread-safe
// operations and automatic migration between configuration versions.
package config

// Store provides thread-safe access to application configuration.
// It supports both file-based and environment-based configuration sources
// and maintains indexes for efficient lookups.
type Store struct {
    // mu protects all configuration access
    mu      sync.RWMutex
    cfg     Config
    path    string
    fromEnv bool
    keyMap  map[string]struct{} // O(1) API key lookup index
    accMap  map[string]int      // O(1) account lookup: identifier -> slice index
}

// HasValidAPIKey checks if the provided key exists and is not expired.
// It returns true for both legacy keys (from Keys array) and new API keys
// (from APIKeys array) that are within their validity period.
//
// Parameters:
//   key - The API key to validate
//
// Returns:
//   bool - True if the key exists and is valid, false otherwise
func (s *Store) HasValidAPIKey(key string) bool {
    // implementation...
}
```

---

### 8.2 Frontend - Component Documentation

**Issues:**
- No component documentation
- Unclear prop requirements
- No usage examples

**Refactoring:**
1. Add JSDoc comments
2. Document component interfaces
3. Add usage examples

**Changes:**
```javascript
/**
 * NotificationBanner displays a list of notifications in a fixed position on the screen.
 * Each notification can be dismissed individually and supports different severity levels.
 *
 * @param {Object} props
 * @param {Array<Notification>} props.notifications - Array of notification objects to display
 * @param {Function} props.onDismiss - Callback function called when a notification is dismissed
 * @returns {JSX.Element}
 *
 * @example
 * <NotificationBanner
 *   notifications={[
 *     { id: '1', type: 'warning', message: 'API key expiring soon', dismissed: false }
 *   ]}
 *   onDismiss={(id) => console.log('Dismissed:', id)}
 * />
 */
export default function NotificationBanner({ notifications, onDismiss }) {
    // implementation...
}
```

---

## Implementation Priority

### High Priority (Immediate Impact)
1. Phase 1.1: Store duplication elimination
2. Phase 1.2: API Key Manager filter pattern
3. Phase 2.1: Expiry logic consolidation
4. Phase 4.3: Constant extraction
5. Phase 6.1: Error handling standardization

### Medium Priority (Quality Improvements)
6. Phase 1.3: Admin handler error handling
7. Phase 1.4: Account CRUD simplification
8. Phase 2.2: Account identifier resolution
9. Phase 3.2: Frontend re-render optimization
10. Phase 4.1: Variable naming improvements

### Low Priority (Long-term Benefits)
11. Phase 2.3: API fetch consolidation
12. Phase 2.4: Context-based state management
13. Phase 3.1: Lock contention reduction
14. Phase 5: Type safety improvements
15. Phase 7: Testing improvements
16. Phase 8: Documentation improvements

---

## Success Criteria

- **Code Quality**: Reduced code duplication by at least 30%
- **Maintainability**: Improved function cohesion and reduced cyclomatic complexity
- **Performance**: Reduced lock contention and unnecessary re-renders
- **Type Safety**: Stronger typing throughout the codebase
- **Testing**: Increased test coverage to at least 70%
- **Documentation**: All exported functions and components documented

---

## Notes

- All changes should maintain backward compatibility
- Run full test suite after each phase
- Monitor performance metrics during refactoring
- Keep changelog of breaking changes
- Consider gradual migration for large refactoring
