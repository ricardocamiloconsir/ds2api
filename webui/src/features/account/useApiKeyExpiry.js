import { useState, useEffect, useCallback } from 'react'

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
                setNotifications(data)
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
                    fetchMonitorStatus()
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
            n.id === id ? { ...n, dismissed: true } : n
        ))
    }, [])

    useEffect(() => {
        fetchApiKeysMetadata()
        fetchNotifications()
        fetchMonitorStatus()
    }, [fetchApiKeysMetadata, fetchNotifications, fetchMonitorStatus])

    useEffect(() => {
        let eventSource
        let retryTimeout
        let retryCount = 0
        const maxRetries = 5

        const connectStream = () => {
            eventSource = new EventSource('/admin/notifications/stream')

            eventSource.onmessage = (event) => {
                try {
                    const notification = JSON.parse(event.data)
                    setNotifications(prev => [
                        { ...notification, id: Date.now() + Math.random(), dismissed: false },
                        ...prev
                    ])
                } catch (e) {
                    console.error('Failed to parse notification:', e)
                }
            }

            eventSource.onerror = () => {
                console.warn('SSE connection error, attempting to reconnect...')
                eventSource.close()
                retryCount++
                if (retryCount <= maxRetries) {
                    retryTimeout = setTimeout(connectStream, 5000 * retryCount)
                }
            }

            eventSource.onopen = () => {
                retryCount = 0
                console.log('SSE connection established')
            }
        }

        connectStream()

        return () => {
            if (eventSource) {
                eventSource.close()
            }
            if (retryTimeout) {
                clearTimeout(retryTimeout)
            }
        }
    }, [])

    return {
        apiKeysMetadata,
        notifications,
        monitorStatus,
        loading,
        refresh: async () => {
            setLoading(true)
            await Promise.all([
                fetchApiKeysMetadata(),
                fetchNotifications(),
                fetchMonitorStatus()
            ])
            setLoading(false)
        },
        checkNow,
        updateMonitorSettings,
        dismissNotification,
    }
}
