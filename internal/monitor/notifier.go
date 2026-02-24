package monitor

import (
	"context"
	"time"

	"ds2api/internal/config"
)

type NotificationType string

const (
	NotificationTypeWarning NotificationType = "warning"
	NotificationTypeError  NotificationType = "expired"
)

type Notification struct {
	Type      NotificationType `json:"type"`
	APIKey    string          `json:"api_key"`
	Message   string          `json:"message"`
	ExpiresAt time.Time       `json:"expires_at"`
	Timestamp time.Time       `json:"timestamp"`
}

type Notifier struct {
	subscribers []chan<- Notification
	history     []Notification
	maxHistory  int
	mu          sync.Mutex
}

func NewNotifier() *Notifier {
	return &Notifier{
		history:    make([]Notification, 0, 100),
		maxHistory: 100,
		subscribers: make([]chan<- Notification, 0),
	}
}

func (n *Notifier) Subscribe(ctx context.Context) <-chan Notification {
	n.mu.Lock()
	defer n.mu.Unlock()

	ch := make(chan Notification, 10)
	n.subscribers = append(n.subscribers, ch)

	go func() {
		<-ctx.Done()
		n.unsubscribe(ch)
		close(ch)
	}()

	return ch
}

func (n *Notifier) unsubscribe(ch chan<- Notification) {
	n.mu.Lock()
	defer n.mu.Unlock()
	for i, sub := range n.subscribers {
		if sub == ch {
			n.subscribers = append(n.subscribers[:i], n.subscribers[i+1:]...)
			break
		}
	}
}

func (n *Notifier) notifyExpiring(keys []config.APIKeyMetadata) {
	n.mu.Lock()
	defer n.mu.Unlock()

	for _, key := range keys {
		notification := Notification{
			Type:      NotificationTypeWarning,
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
			Type:      NotificationTypeError,
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
	n.mu.Lock()
	defer n.mu.Unlock()
	result := make([]Notification, len(n.history))
	copy(result, n.history)
	return result
}

func (n *Notifier) addToHistory(notification Notification) {
	n.history = append(n.history, notification)
	if len(n.history) > n.maxHistory {
		n.history = n.history[1:]
	}
}

func (n *Notifier) broadcast(notification Notification) {
	for _, sub := range n.subscribers {
		select {
		case sub <- notification:
		default:
		}
	}
}

func maskAPIKey(key string) string {
	if len(key) <= 16 {
		return "****"
	}
	return key[:8] + "****" + key[len(key)-4:]
}
