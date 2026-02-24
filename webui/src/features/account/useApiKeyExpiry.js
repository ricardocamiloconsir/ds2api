import { useState, useEffect, useCallback } from 'react'

const normalizeNotification = (notification) => {
    const apiKey = notification.api_key ?? notification.apiKey ?? ''
    const timestamp = notification.timestamp ?? ''

    return {
        ...notification,
        id: notification.id ?? `${timestamp}_${apiKey}`,
        dismissed: notification.dismissed ?? false,
    }
}

export function useApiKeyExpiry({ apiFetch, t }) {
    const [apiKeysMetadata, setApiKeysMetadata] = useState([])
    const [notifications, setNotifications] = useState([])
    const [monitorStatus, setMonitorStatus] = useState(null)
    const [loading, setLoading] = useState(false)

    const fetchApiKeysMetadata = useCallback(async () => {
        try {
            const res = await apiFetch('/admin/keys/metadata')
            if (res.ok) {
                const data = await res.json()
                setApiKeysMetadata(data)
            }
        } catch (e) {
            console.error('Failed to fetch API keys metadata:', e)
        }
    }, [apiFetch])

    const fetchNotifications = useCallback(async () => {
        try {
            const res = await apiFetch('/admin/notifications')
            if (res.ok) {
                const data = await res.json()
                setNotifications(data.map(normalizeNotification))
            }
        } catch (e) {
            console.error('Failed to fetch notifications:', e)
        }
    }, [apiFetch])

    const fetchMonitorStatus = useCallback(async () => {
        try {
            const res = await apiFetch('/admin/monitor/status')
            if (res.ok) {
                const data = await res.json()
                setMonitorStatus(data)
            }
        } catch (e) {
            console.error('Failed to fetch monitor status:', e)
        }
    }, [apiFetch])

    const checkNow = useCallback(async () => {
        try {
            const res = await apiFetch('/admin/monitor/check', { method: 'POST' })
            if (res.ok) {
                await Promise.all([
                    fetchApiKeysMetadata(),
                    fetchNotifications(),
                    fetchMonitorStatus(),
                ])
            }
        } catch (e) {
            console.error('Failed to check monitor:', e)
        }
    }, [apiFetch, fetchApiKeysMetadata, fetchNotifications, fetchMonitorStatus])

    const updateMonitorSettings = useCallback(async (settings) => {
        try {
            const res = await apiFetch('/admin/monitor/settings', {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(settings),
            })
            if (res.ok) {
                await fetchMonitorStatus()
            }
        } catch (e) {
            console.error('Failed to update monitor settings:', e)
        }
    }, [apiFetch, fetchMonitorStatus])

    const dismissNotification = useCallback((id) => {
        setNotifications(prev => prev.map(n =>
            n.id === id ? { ...n, dismissed: true } : n,
        ))
    }, [])

    useEffect(() => {
        fetchApiKeysMetadata()
        fetchNotifications()
        fetchMonitorStatus()
    }, [fetchApiKeysMetadata, fetchNotifications, fetchMonitorStatus])

    useEffect(() => {
        let retryTimeout
        let retryCount = 0
        let cancelled = false
        let abortController = null
        const maxRetries = 5

        const processSSEBuffer = (buffer, onMessage) => {
            const events = buffer.split('\n\n')
            const incomplete = events.pop() ?? ''

            for (const eventChunk of events) {
                const lines = eventChunk.split('\n')
                const payload = lines
                    .filter(line => line.startsWith('data:'))
                    .map(line => line.slice(5).trimStart())
                    .join('\n')

                if (payload) {
                    onMessage(payload)
                }
            }

            return incomplete
        }

        const connectStream = async () => {
            if (cancelled) return

            abortController = new AbortController()

            try {
                // Use authenticated fetch for SSE because EventSource cannot send Authorization headers.
                const res = await apiFetch('/admin/notifications/stream', {
                    signal: abortController.signal,
                    headers: { Accept: 'text/event-stream' },
                })

                if (!res.ok || !res.body) {
                    throw new Error(`SSE stream failed with status ${res.status}`)
                }

                retryCount = 0
                console.log('SSE connection established')

                const reader = res.body.getReader()
                const decoder = new TextDecoder()
                let buffer = ''

                while (!cancelled) {
                    const { value, done } = await reader.read()
                    if (done) {
                        break
                    }

                    buffer += decoder.decode(value, { stream: true })
                    buffer = processSSEBuffer(buffer, (payload) => {
                        try {
                            const notification = normalizeNotification(JSON.parse(payload))
                            setNotifications(prev => [notification, ...prev])
                        } catch (e) {
                            console.error('Failed to parse notification:', e)
                        }
                    })
                }
            } catch (e) {
                if (cancelled || e.name === 'AbortError') return

                console.warn('SSE connection error, attempting to reconnect...', e)
                retryCount++
                if (retryCount <= maxRetries) {
                    retryTimeout = setTimeout(() => {
                        if (!cancelled) {
                            connectStream()
                        }
                    }, 5000 * retryCount)
                }
            }
        }

        connectStream()

        return () => {
            cancelled = true
            if (abortController) {
                abortController.abort()
            }
            if (retryTimeout) {
                clearTimeout(retryTimeout)
            }
        }
    }, [apiFetch])

    const refresh = useCallback(async () => {
        setLoading(true)
        try {
            await Promise.all([
                fetchApiKeysMetadata(),
                fetchNotifications(),
                fetchMonitorStatus(),
            ])
        } finally {
            setLoading(false)
        }
    }, [fetchApiKeysMetadata, fetchNotifications, fetchMonitorStatus])

    return {
        apiKeysMetadata,
        notifications,
        monitorStatus,
        loading,
        refresh,
        checkNow,
        updateMonitorSettings,
        dismissNotification,
    }
}
