import { useState, type SubmitEvent } from 'react'
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

  async function handleSubmit(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault()
    setError('')
    setSubmitting(true)

    try {
      const response = await fetch('/api/login', {
        method: "POST",
        headers: { 'Content-Type' : 'application/json' },
        body: JSON.stringify({email, password})
      })

      if (!response.ok) {
        setError((await response.text()) || 'Login failed.')
      } else {
        // Handle successful login - redirect to user dashboards.
        navigate('/dashboard')
      } 
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not connect to the server')
    } finally {
      setSubmitting(false)
    }
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