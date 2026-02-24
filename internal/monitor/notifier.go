package monitor

import (
	"context"
	"ds2api/internal/config"
	"sync"
	"time"
)

type Notification struct {
	ID        string                  `json:"id"`
	Type      config.NotificationType `json:"type"`
	APIKey    string                  `json:"apiKey"`
	Message   string                  `json:"message"`
	ExpiresAt time.Time               `json:"expiresAt"`
	Timestamp time.Time               `json:"timestamp"`
	Data      map[string]any          `json:"data,omitempty"`
}

type Notifier struct {
	mu          sync.RWMutex
	subscribers map[chan Notification]struct{}
	history     []Notification
	maxHistory  int
}

func NewNotifier(maxHistory ...int) *Notifier {
	mh := config.DefaultMaxHistory
	if len(maxHistory) > 0 && maxHistory[0] > 0 {
		mh = maxHistory[0]
	}
	return &Notifier{
		subscribers: make(map[chan Notification]struct{}),
		history:     make([]Notification, 0),
		maxHistory:  mh,
	}
}

func (n *Notifier) Subscribe(ctx context.Context) <-chan Notification {
	n.mu.Lock()
	defer n.mu.Unlock()

	ch := make(chan Notification, config.NotificationBufferSize)
	n.subscribers[ch] = struct{}{}

	go func() {
		<-ctx.Done()
		n.unsubscribe(ch)
		close(ch)
	}()

	return ch
}

func (n *Notifier) unsubscribe(ch chan Notification) {
	n.mu.Lock()
	defer n.mu.Unlock()
	delete(n.subscribers, ch)
}

func (n *Notifier) notifyExpiring(keys []config.APIKeyMetadata) {
	n.mu.Lock()
	defer n.mu.Unlock()

	for _, key := range keys {
		notification := Notification{
			ID:        key.ID + ":" + key.ExpiresAt.Format(time.RFC3339Nano) + ":warning",
			Type:      config.NotificationTypeWarning,
			APIKey:    maskAPIKey(key.Key),
			Message:   "API key expiring soon",
			ExpiresAt: key.ExpiresAt,
			Timestamp: time.Now(),
		}
		n.addToHistory(notification)
		n.broadcast(notification)
	}
}

func (n *Notifier) notifyExpired(keys []config.APIKeyMetadata) {
	n.mu.Lock()
	defer n.mu.Unlock()

	for _, key := range keys {
		notification := Notification{
			ID:        key.ID + ":" + key.ExpiresAt.Format(time.RFC3339Nano) + ":expired",
			Type:      config.NotificationTypeExpired,
			APIKey:    maskAPIKey(key.Key),
			Message:   "API key has expired",
			ExpiresAt: key.ExpiresAt,
			Timestamp: time.Now(),
		}
		n.addToHistory(notification)
		n.broadcast(notification)
	}
}

func (n *Notifier) GetHistory() []Notification {
	n.mu.RLock()
	defer n.mu.RUnlock()
	result := make([]Notification, len(n.history))
	copy(result, n.history)
	return result
}

func (n *Notifier) addToHistory(notification Notification) {
	if len(n.history) < n.maxHistory {
		n.history = append(n.history, notification)
		return
	}

	trimmed := make([]Notification, n.maxHistory)
	copy(trimmed, n.history[1:])
	trimmed[n.maxHistory-1] = notification
	n.history = trimmed
}

func (n *Notifier) broadcast(notification Notification) {
	for sub := range n.subscribers {
		select {
		case sub <- notification:
		default:
			// Drop notifications for slow subscribers to avoid blocking notifier operations.
		}
	}
}

func maskAPIKey(key string) string {
	if len(key) <= 17 {
		return "****"
	}
	return key[:11] + "****" + key[len(key)-4:]
}
