import { useEffect, useRef, useState } from 'react'
import { API_BASE_URL } from './config'

function App() {
  const activeRequest = useRef<AbortController | null>(null)
  const [result, setResult] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  useEffect(() => () => activeRequest.current?.abort(), [])

  async function ping() {
    activeRequest.current?.abort()
    const controller = new AbortController()
    activeRequest.current = controller
    setLoading(true)
    setResult('')
    setError('')

    try {
      const response = await fetch(`${API_BASE_URL.replace(/\/$/, '')}/api/v1/ping`, {
        signal: controller.signal,
      })
      if (!response.ok) throw new Error(`Ping failed (HTTP ${response.status})`)
      const data = await response.json()
      if (data.message !== 'pong' || data.service !== 'ping-2') {
        throw new Error('Invalid pong response')
      }
      if (!controller.signal.aborted) setResult(`${data.message} (${data.service})`)
    } catch (err) {
      if (!controller.signal.aborted) {
        setError(err instanceof Error ? err.message : 'Ping failed')
      }
    } finally {
      if (!controller.signal.aborted) setLoading(false)
    }
  }

  return (
    <main>
      <h1>ping-web</h1>
      <p>ping-web → ping-api → ping-2</p>
      <button type="button" disabled={loading} onClick={() => void ping()}>
        {loading ? 'Pinging…' : 'Ping'}
      </button>
      {result && <p role="status">{result}</p>}
      {error && <p role="alert">{error}</p>}
    </main>
  )
}

export default App
