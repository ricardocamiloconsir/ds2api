import { Clock, AlertTriangle, AlertCircle, RefreshCw } from 'lucide-react'
import { useEffect, useState } from 'react'
import clsx from 'clsx'
import { calculateDaysUntilExpiry, getKeyExpiryStatusFromMetadata, EXPIRY_THRESHOLDS } from '../../utils/apiKeyUtils'

export default function ApiKeyExpiryPanel({ t, apiKeysMetadata, onRefresh, onCheckNow }) {
    const [loading, setLoading] = useState(false)

    const validKeys = apiKeysMetadata.filter(k => getKeyExpiryStatusFromMetadata(k).status === 'valid')
    const expiringKeys = apiKeysMetadata.filter(k => {
        const status = getKeyExpiryStatusFromMetadata(k)
        return status.status === 'expiring'
    })
    const expiredKeys = apiKeysMetadata.filter(k => {
        const status = getKeyExpiryStatusFromMetadata(k)
        return status.status === 'expired'
    })

    const getStatusBadge = (daysLeft) => {
        if (daysLeft <= EXPIRY_THRESHOLDS.CRITICAL) {
            return { color: 'red', text: t('apiKey.expired'), icon: AlertCircle }
        } else if (daysLeft <= EXPIRY_THRESHOLDS.WARNING) {
            return { color: 'yellow', text: t('apiKey.expiringSoon'), icon: AlertTriangle }
        }
        return { color: 'green', text: t('apiKey.valid'), icon: Clock }
    }

    const handleRefresh = async () => {
        setLoading(true)
        await onRefresh()
        setLoading(false)
    }

    const handleCheckNow = async () => {
        setLoading(true)
        await onCheckNow()
        setLoading(false)
    }

    return (
        <div className="space-y-4">
            <div className="bg-card border border-border rounded-xl overflow-hidden shadow-sm">
                <div className="p-6 border-b border-border">
                    <div className="flex items-center justify-between">
                        <div>
                            <h2 className="text-lg font-semibold">{t('apiKey.expiryTitle')}</h2>
                            <p className="text-sm text-muted-foreground mt-1">{t('apiKey.expiryDesc')}</p>
                        </div>
                        <div className="flex gap-2">
                            <button
                                onClick={handleRefresh}
                                disabled={loading}
                                className="p-2 hover:bg-muted rounded-md transition-colors disabled:opacity-50"
                                title={t('messages.refresh')}
                            >
                                <RefreshCw className={clsx("w-5 h-5", loading && "animate-spin")} />
                            </button>
                            <button
                                onClick={handleCheckNow}
                                disabled={loading}
                                className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors disabled:opacity-50 font-medium text-sm"
                            >
                                {t('apiKey.checkNow')}
                            </button>
                        </div>
                    </div>
                </div>

                <div className="divide-y divide-border">
                    {validKeys.length > 0 && (
                        <div className="p-4">
                            <h3 className="text-sm font-medium text-green-600 dark:text-green-400 mb-3 flex items-center gap-2">
                                <Clock className="w-4 h-4" />
                                {t('apiKey.validKeys')} ({validKeys.length})
                            </h3>
                            <div className="space-y-2">
                                {validKeys.map((key) => {
                                    const daysLeft = calculateDaysUntilExpiry(key.expires_at)
                                    const status = getStatusBadge(daysLeft)
                                    return (
                                        <div
                                            key={key.id}
                                            className="flex items-center justify-between p-3 bg-muted/30 rounded-lg hover:bg-muted/50 transition-colors"
                                        >
                                            <div className="flex items-center gap-3">
                                                <status.icon className="w-4 h-4 text-green-600 dark:text-green-400" />
                                                <div>
                                                    <p className="font-mono text-sm">{key.key.slice(0, 16)}****</p>
                                                    <p className="text-xs text-muted-foreground">
                                                        {t('apiKey.expiresAt')}: {new Date(key.expires_at).toLocaleString()}
                                                    </p>
                                                </div>
                                            </div>
                                            <span className="text-xs font-medium px-2 py-1 bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-300 rounded">
                                                {daysLeft} {t('apiKey.days')}
                                            </span>
                                        </div>
                                    )
                                })}
                            </div>
                        </div>
                    )}

                    {expiringKeys.length > 0 && (
                        <div className="p-4 bg-yellow-50/50 dark:bg-yellow-900/10">
                            <h3 className="text-sm font-medium text-yellow-600 dark:text-yellow-400 mb-3 flex items-center gap-2">
                                <AlertTriangle className="w-4 h-4" />
                                {t('apiKey.expiringKeys')} ({expiringKeys.length})
                            </h3>
                            <div className="space-y-2">
                                {expiringKeys.map((key) => {
                                    const daysLeft = calculateDaysUntilExpiry(key.expires_at)
                                    const status = getStatusBadge(daysLeft)
                                    return (
                                        <div
                                            key={key.id}
                                            className="flex items-center justify-between p-3 bg-yellow-100/50 dark:bg-yellow-900/20 rounded-lg"
                                        >
                                            <div className="flex items-center gap-3">
                                                <status.icon className="w-4 h-4 text-yellow-600 dark:text-yellow-400" />
                                                <div>
                                                    <p className="font-mono text-sm">{key.key.slice(0, 16)}****</p>
                                                    <p className="text-xs text-muted-foreground">
                                                        {t('apiKey.expiresAt')}: {new Date(key.expires_at).toLocaleString()}
                                                    </p>
                                                </div>
                                            </div>
                                            <span className="text-xs font-medium px-2 py-1 bg-yellow-200 dark:bg-yellow-800 text-yellow-800 dark:text-yellow-200 rounded">
                                                {daysLeft} {t('apiKey.days')}
                                            </span>
                                        </div>
                                    )
                                })}
                            </div>
                        </div>
                    )}

                    {expiredKeys.length > 0 && (
                        <div className="p-4 bg-red-50/50 dark:bg-red-900/10">
                            <h3 className="text-sm font-medium text-red-600 dark:text-red-400 mb-3 flex items-center gap-2">
                                <AlertCircle className="w-4 h-4" />
                                {t('apiKey.expiredKeys')} ({expiredKeys.length})
                            </h3>
                            <div className="space-y-2">
                                {expiredKeys.map((key) => {
                                    const daysLeft = calculateDaysUntilExpiry(key.expires_at)
                                    const status = getStatusBadge(daysLeft)
                                    return (
                                        <div
                                            key={key.id}
                                            className="flex items-center justify-between p-3 bg-red-100/50 dark:bg-red-900/20 rounded-lg"
                                        >
                                            <div className="flex items-center gap-3">
                                                <status.icon className="w-4 h-4 text-red-600 dark:text-red-400" />
                                                <div>
                                                    <p className="font-mono text-sm">{key.key.slice(0, 16)}****</p>
                                                    <p className="text-xs text-muted-foreground">
                                                        {t('apiKey.expiredOn')}: {new Date(key.expires_at).toLocaleString()}
                                                    </p>
                                                </div>
                                            </div>
                                            <span className="text-xs font-medium px-2 py-1 bg-red-200 dark:bg-red-800 text-red-800 dark:text-red-200 rounded">
                                                {t('apiKey.expired')}
                                            </span>
                                        </div>
                                    )
                                })}
                            </div>
                            <div className="mt-4 p-3 bg-red-100 dark:bg-red-900/30 rounded-lg">
                                <p className="text-sm text-red-700 dark:text-red-300">
                                    {t('apiKey.expiredAdvice')}
                                </p>
                            </div>
                        </div>
                    )}

                    {apiKeysMetadata.length === 0 && (
                        <div className="p-8 text-center text-muted-foreground">
                            {t('apiKey.noApiKeys')}
                        </div>
                    )}
                </div>
            </div>
        </div>
    )
}
