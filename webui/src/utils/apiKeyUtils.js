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

export const getKeyExpiryStatusFromMetadata = (metadata) => {
    if (!metadata) return { status: 'valid', daysLeft: null }
    return getKeyExpiryStatus(metadata.expires_at)
}

export const maskApiKey = (key) => {
    if (!key || key.length <= 16) return '****'
    return key.slice(0, 16) + '****'
}
