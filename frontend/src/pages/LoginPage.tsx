import { useEffect, useState, type SubmitEvent } from 'react'
import { Button } from '@/components/ui/shadcn/button'
import { Input } from '@/components/ui/shadcn/input'
import { ButtonLink } from '@/components/ui/button-link'
import { useLocation } from 'react-router-dom'
import { useNavigate } from 'react-router-dom'

export default function LoginPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const signupSuccess = (
    location.state as { signupSuccess?: string } | null
  )?.signupSuccess


  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [checkingSession, setCheckingSession] = useState(true)

  useEffect(() => {
    const controller = new AbortController()

    async function checkSession() {
      try {
        const response = await fetch('/api/me', {
          credentials: 'include',
          signal: controller.signal,
        })

        if (response.ok) {
          navigate('/dashboard', { replace: true })
        } else if (response.status !== 401) {
          setError(`Could not check your session (HTTP ${response.status}).`)
        }
      } catch (err) {
        if (!controller.signal.aborted) {
          setError(err instanceof Error ? err.message : 'Could not check your session.')
        }
      } finally {
        if (!controller.signal.aborted) {
          setCheckingSession(false)
        }
      }
    }

    void checkSession()
    return () => controller.abort()
  }, [navigate])

  async function handleSubmit(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault()
    setError('')
    setSubmitting(true)

    try {
      const response = await fetch('/api/login', {
        method: "POST",
        headers: { 'Content-Type' : 'application/json' },
        credentials: 'include',
        body: JSON.stringify({email, password})
      })

      if (!response.ok) {
        setError((await response.text()) || 'Login failed.')
      } else {
        navigate('/dashboard', { replace: true })
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not connect to the server')
    } finally {
      setSubmitting(false)
    }
  }
  
  if (checkingSession) {
    return (
      <main className="flex min-h-[calc(100vh-3.5rem)] items-center justify-center p-6">
        <p role="status">Checking your session...</p>
      </main>
    )
  }

  return (
    <main className='flex min-h-[calc(100vh-3.5rem)] items-center justify-center p-6'>
      <section className='w-full max-w-sm space-y-6'>
      <h1 className='text-center text-2xl font-semibold'>Login</h1>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div className='space-y-2'>
          <label htmlFor='email'>Email</label>
          <Input id='email' name='email' type='email' autoComplete='email' required value={email} onChange={(event) => setEmail(event.target.value)} />
        </div>
        <div className='space-y-2'>
          <label htmlFor='password'>Password</label>
          <Input id='password' name='password' type='password' autoComplete='current-password' required value={password} onChange={(event) => setPassword(event.target.value)} />
        </div>
        
        <Button className='w-full' type="submit" disabled={submitting}>{submitting ? 'Logging in...' : 'Log in'}</Button>
        <ButtonLink to={'/signup'} variant='outline' className='w-full'>Sign Up</ButtonLink>
      
        {/* Alerting */}
        {signupSuccess && <p role="status" className="rounded-md border border-green-500/30 bg-green-500/10 p-3 text-center text-sm text-green-700 dark:text-green-400">{signupSuccess}</p>}
        {error && <p role='alert'>{error}</p>}

      </form>
      </section>
    </main>
  )
}