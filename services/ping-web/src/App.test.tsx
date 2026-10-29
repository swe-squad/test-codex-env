import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import App from './App'

afterEach(() => {
  cleanup()
  vi.unstubAllGlobals()
})

describe('App', () => {
  it('calls ping-api and displays the pong from ping-2', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ message: 'pong', service: 'ping-2' }),
    })
    vi.stubGlobal('fetch', fetchMock)
    render(<App />)
    fireEvent.click(screen.getByRole('button', { name: 'Ping' }))
    expect(screen.getByRole('button', { name: 'Pinging…' })).toBeDisabled()
    expect(await screen.findByRole('status')).toHaveTextContent('pong (ping-2)')
    expect(fetchMock).toHaveBeenCalledWith('/api/v1/ping', {
      signal: expect.any(AbortSignal),
    })
    expect(screen.getByRole('button', { name: 'Ping' })).toBeEnabled()
  })

  it('shows upstream failure and allows retry', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce({ ok: false, status: 502 })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({ message: 'pong', service: 'ping-2' }),
      })
    vi.stubGlobal('fetch', fetchMock)
    render(<App />)
    fireEvent.click(screen.getByRole('button', { name: 'Ping' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('Ping failed (HTTP 502)')
    fireEvent.click(screen.getByRole('button', { name: 'Ping' }))
    expect(await screen.findByRole('status')).toHaveTextContent('pong (ping-2)')
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('shows network errors', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('Network error')))
    render(<App />)
    fireEvent.click(screen.getByRole('button', { name: 'Ping' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('Network error')
  })

  it('rejects a pong from the wrong service', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ message: 'pong', service: 'ping-api' }),
    }))
    render(<App />)
    fireEvent.click(screen.getByRole('button', { name: 'Ping' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('Invalid pong response')
  })

  it('cancels an in-flight request on unmount', async () => {
    const fetchMock = vi.fn().mockImplementation(() => new Promise(() => {}))
    vi.stubGlobal('fetch', fetchMock)
    const { unmount } = render(<App />)
    fireEvent.click(screen.getByRole('button', { name: 'Ping' }))
    await waitFor(() => expect(fetchMock).toHaveBeenCalledOnce())
    const signal = fetchMock.mock.calls[0][1].signal as AbortSignal
    unmount()
    expect(signal.aborted).toBe(true)
  })
})
