import { useState } from 'react'
import { Button } from '@/components/ui/button'
import './App.css'

function App() {
  const [apiStatus, setApiStatus] = useState('Not checked')
  const [checking, setChecking] = useState(false)

  async function checkAPI() {
    setChecking(true)
    setApiStatus('Checking…')

    try {
      const response = await fetch('/health')
      if (!response.ok) {
        throw new Error(`API returned HTTP ${response.status}`)
      }
      setApiStatus(`Connected: ${await response.text()}`)
    } catch (error) {
      setApiStatus(error instanceof Error ? error.message : 'Could not reach API')
    } finally {
      setChecking(false)
    }
  }

  return (
    <main className="app-shell">
      <header>
        <p className="eyebrow">Prototype</p>
        <h1>Digital Autonomy</h1>
        <p className="intro">A starting point for the financial proxy dashboard.</p>
      </header>

      <section className="status-card" aria-labelledby="api-heading">
        <div>
          <h2 id="api-heading">API connection</h2>
          <p role="status" aria-live="polite">{apiStatus}</p>
        </div>
        <Button type="button" onClick={checkAPI} disabled={checking} className="min-h-12 px-5">
          {checking ? 'Checking…' : 'Check API'}
        </Button>
      </section>
    </main>
  )
}

export default App
