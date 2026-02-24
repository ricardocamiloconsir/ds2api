package config

import "time"

type NotificationType string

const (
	DefaultCheckInterval   = 24 * time.Hour
	DefaultWarningDays     = 7
	DefaultMaxHistory      = 100
	NotificationBufferSize = 10

	APIKeyTTLDays = 30
	APIKeyTTL     = APIKeyTTLDays * 24 * time.Hour

	MaxPageSize     = 100
	DefaultPageSize = 10
)

const (
	NotificationTypeWarning NotificationType = "warning"
	NotificationTypeExpired NotificationType = "expired"
)

const (
	SSEContentType         = "text/event-stream"
	SSECacheControl        = "no-cache"
	SSEConnection     = "keep-alive"
	SSETimeoutDefault = 2 * time.Hour
)
