package config

import "time"

func APIKeyExpiryFrom(createdAt time.Time) time.Time {
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	return createdAt.AddDate(0, 1, 0)
}

func ResolveAPIKeyExpiry(metadata APIKeyMetadata) time.Time {
	if !metadata.ExpiresAt.IsZero() {
		return metadata.ExpiresAt
	}
	if !metadata.CreatedAt.IsZero() {
		return APIKeyExpiryFrom(metadata.CreatedAt)
	}
	return time.Time{}
}

func IsAPIKeyActiveAt(metadata APIKeyMetadata, now time.Time) bool {
	expiry := ResolveAPIKeyExpiry(metadata)
	if expiry.IsZero() {
		return false
	}
	return now.Before(expiry)
}
