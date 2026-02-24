import { Check, ChevronDown, Copy, Plus, Trash2, Clock, AlertTriangle, AlertCircle } from 'lucide-react'
import clsx from 'clsx'
import { getKeyExpiryStatusFromMetadata } from '../../utils/apiKeyUtils'

export default function ApiKeysPanel({
    t,
    config,
    keysExpanded,
    setKeysExpanded,
    setShowAddKey,
    copiedKey,
    setCopiedKey,
    onDeleteKey,
    apiKeysMetadata = [],
}) {
    const getKeyExpiryStatus = (key) => {
        const metadata = apiKeysMetadata.find(m => m.key === key)
        return getKeyExpiryStatusFromMetadata(metadata)
    }

    return (
        <div className="bg-card border border-border rounded-xl overflow-hidden shadow-sm">
            <div
                className="p-6 flex flex-col md:flex-row md:items-center justify-between gap-4 cursor-pointer select-none hover:bg-muted/30 transition-colors"
                onClick={() => setKeysExpanded(!keysExpanded)}
            >
                <div className="flex items-center gap-3">
                    <ChevronDown className={clsx(
                        "w-5 h-5 text-muted-foreground transition-transform duration-200",
                        keysExpanded ? "rotate-0" : "-rotate-90"
                    )} />
                    <div>
                        <h2 className="text-lg font-semibold">{t('accountManager.apiKeysTitle')}</h2>
                        <p className="text-sm text-muted-foreground">{t('accountManager.apiKeysDesc')} ({config.keys?.length || 0})</p>
                    </div>
                </div>
                <button
                    onClick={(e) => { e.stopPropagation(); setShowAddKey(true) }}
                    className="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors font-medium text-sm shadow-sm"
                >
                    <Plus className="w-4 h-4" />
                    {t('accountManager.addKey')}
                </button>
            </div>

            {keysExpanded && (
                <div className="divide-y divide-border border-t border-border">
                    {config.keys?.length > 0 ? (
                        config.keys.map((key, i) => {
                            const expiryStatus = getKeyExpiryStatus(key)
                            return (
                                <div key={i} className="p-4 flex items-center justify-between hover:bg-muted/50 transition-colors group">
                                    <div className="flex items-center gap-3">
                                        <div className="font-mono text-sm bg-muted/50 px-3 py-1 rounded inline-block">
                                            {key.slice(0, 16)}****
                                        </div>
                                        {expiryStatus.status !== 'valid' && (
                                            <span className={clsx(
                                                "text-xs font-medium px-2 py-1 rounded flex items-center gap-1",
                                                expiryStatus.status === 'expired'
                                                    ? "bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300"
                                                    : "bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-300"
                                            )}>
                                                {expiryStatus.status === 'expired' ? (
                                                    <AlertCircle className="w-3 h-3" />
                                                ) : (
                                                    <AlertTriangle className="w-3 h-3" />
                                                )}
                                                {expiryStatus.status === 'expired' 
                                                    ? t('apiKey.expired')
                                                    : `${t('apiKey.expiresIn')} ${expiryStatus.daysLeft} ${t('apiKey.days')}`
                                                }
                                            </span>
                                        )}
                                        {copiedKey === key && (
                                            <span className="text-xs text-green-500 animate-pulse">{t('accountManager.copied')}</span>
                                        )}
                                    </div>
                                    <div className="flex items-center gap-1">
                                        <button
                                            onClick={() => {
                                                navigator.clipboard.writeText(key)
                                                setCopiedKey(key)
                                                setTimeout(() => setCopiedKey(null), 2000)
                                            }}
                                            className="p-2 text-muted-foreground hover:text-primary hover:bg-primary/10 rounded-md transition-colors opacity-0 group-hover:opacity-100"
                                            title={t('accountManager.copyKeyTitle')}
                                        >
                                            {copiedKey === key ? <Check className="w-4 h-4 text-green-500" /> : <Copy className="w-4 h-4" />}
                                        </button>
                                        <button
                                            onClick={() => onDeleteKey(key)}
                                            className="p-2 text-muted-foreground hover:text-destructive hover:bg-destructive/10 rounded-md transition-colors opacity-0 group-hover:opacity-100"
                                            title={t('accountManager.deleteKeyTitle')}
                                        >
                                            <Trash2 className="w-4 h-4" />
                                        </button>
                                    </div>
                                </div>
                            )
                        })
                    ) : (
                        <div className="p-8 text-center text-muted-foreground">{t('accountManager.noApiKeys')}</div>
                    )}
                </div>
            )}
        </div>
    )
}
