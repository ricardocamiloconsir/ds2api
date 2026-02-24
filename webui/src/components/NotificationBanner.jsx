import { X, AlertTriangle, AlertCircle } from 'lucide-react'
import { useEffect, useState } from 'react'

export default function NotificationBanner({ notifications, onDismiss }) {
    const [visibleNotifications, setVisibleNotifications] = useState([])

    useEffect(() => {
        setVisibleNotifications(notifications.filter(n => !n.dismissed))
    }, [notifications])

    if (visibleNotifications.length === 0) return null

    return (
        <div className="fixed top-4 right-4 z-50 space-y-2 max-w-md">
            {visibleNotifications.map((notification) => (
                <div
                    key={notification.id}
                    className={`
                        flex items-start gap-3 p-4 rounded-lg shadow-lg border animate-slide-in
                        ${notification.type === 'warning'
                            ? 'bg-yellow-50 border-yellow-200 dark:bg-yellow-900/20 dark:border-yellow-800'
                            : notification.type === 'error'
                                ? 'bg-red-50 border-red-200 dark:bg-red-900/20 dark:border-red-800'
                                : 'bg-blue-50 border-blue-200 dark:bg-blue-900/20 dark:border-blue-800'
                        }
                    `}
                >
                    <div className="flex-shrink-0">
                        {notification.type === 'warning' ? (
                            <AlertTriangle className="w-5 h-5 text-yellow-600 dark:text-yellow-400" />
                        ) : notification.type === 'error' ? (
                            <AlertCircle className="w-5 h-5 text-red-600 dark:text-red-400" />
                        ) : null}
                    </div>
                    <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium text-gray-900 dark:text-gray-100">
                            {notification.message}
                        </p>
                        {notification.apiKey && (
                            <p className="text-xs text-gray-600 dark:text-gray-400 mt-1 font-mono">
                                Key: {notification.apiKey}
                            </p>
                        )}
                    </div>
                    <button
                        onClick={() => onDismiss(notification.id)}
                        className="flex-shrink-0 p-1 hover:bg-black/5 dark:hover:bg-white/10 rounded-md transition-colors"
                    >
                        <X className="w-4 h-4 text-gray-500 dark:text-gray-400" />
                    </button>
                </div>
            ))}
        </div>
    )
}
