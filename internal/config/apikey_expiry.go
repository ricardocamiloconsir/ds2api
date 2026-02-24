package config

import "time"

func APIKeyExpiryFrom(createdAt time.Time) time.Time {
	return time.Time{}
}

func ResolveAPIKeyExpiry(metadata APIKeyMetadata) time.Time {
	_ = metadata
	return time.Time{}
}

func IsAPIKeyActiveAt(metadata APIKeyMetadata, now time.Time) bool {
	_ = now
	return metadata.Key != ""
}
