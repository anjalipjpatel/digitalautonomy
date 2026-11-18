import { useState, type SubmitEvent } from 'react'
import { Button } from '@/components/ui/shadcn/button'
import { Input } from '@/components/ui/shadcn/input'
import { ButtonLink } from '@/components/ui/button-link'
import { USER_ROLES } from '@/types/user'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/shadcn/select'
import { useNavigate } from 'react-router-dom'

export default function SignUpPage() {
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const [role, setRole] = useState(USER_ROLES.OWNER)
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault()
    setError('')
    setSubmitting(true)

    try {
      // Client-side validation - passwords match
      if (!validatePasswordsSame(password, confirmPassword)) {
        setError('Passwords must match.')
        return
      }

      // Send sign-up request
      const response = await fetch('/api/users', {
        method: "POST",
        headers: { 'Content-Type' : 'application/json' },
        body: JSON.stringify({email, name, password, role})
      })

      if (!response.ok) {
        setError((await response.text()) || 'Login failed.')
      } else {
        // Handle successful sign up - redirect to login
        navigate('/login', {
        state: { signupSuccess: 'Account created. Please log in.' },
        })
        
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not connect to the server')
    } finally {
      setSubmitting(false)
    }
  }

    return(
        <main className='flex min-h-[calc(100vh-3.5rem)] items-center justify-center p-6'>
        <section className='w-full max-w-sm space-y-6'>
        <h1 className='text-center text-2xl font-semibold'>Sign Up</h1>

        <form onSubmit={handleSubmit} className="space-y-4">
            
            <div className='space-y-2'>            
            <label htmlFor='name'>Name</label>
            <Input id='name' name='name' type='text' autoComplete='text' required value={name} onChange={(event) => setName(event.target.value)} />
            </div>

            <div className='space-y-2'>
            <label htmlFor="role">Role</label>
            <Select value={role} onValueChange={(value) => {
                if(value === USER_ROLES.OWNER || value === USER_ROLES.DELEGATE) {
                    setRole(value)
                }
            }}
            >
              <SelectTrigger id="role" className="w-full">
              <SelectValue>
                {(value) => {
                    if (value === USER_ROLES.OWNER) return 'Owner'
                    if (value === USER_ROLES.DELEGATE) return 'Delegate'
                    return 'Choose a role'
                }}
              </SelectValue>
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={USER_ROLES.OWNER}>Owner</SelectItem>
                <SelectItem value={USER_ROLES.DELEGATE}>Delegate</SelectItem>
              </SelectContent>
            </Select>
            </div>

            <div className='space-y-2'>
            <label htmlFor='email'>Email</label>
            <Input id='email' name='email' type='email' autoComplete='email' required value={email} onChange={(event) => setEmail(event.target.value)} />
            </div>
            <div className='space-y-2'>
            <label htmlFor='password'>Password</label>
            <Input id='password' name='password' type='password' autoComplete='password' required value={password} onChange={(event) => setPassword(event.target.value)} />
            </div>
            <div className='space-y-2'>
            <label htmlFor='password'>Confirm Password</label>
            <Input id='confirm-password' name='confirm-password' type='password' autoComplete='confirm-password' required value={confirmPassword} onChange={(event) => setConfirmPassword(event.target.value)} />
            </div>

            {/* Role:  RoleVulnerableOwner, RoleDelegate*/}

            <Button className='w-full' type="submit" disabled={submitting}>{submitting ? 'Signing up...' : 'Sign up'}</Button>
            <ButtonLink to={'/login'} variant='outline' className='w-full'>Log in</ButtonLink>

            {error && <p role='alert'>{error}</p>}
        </form>
        </section>
        </main>
    )
}

function validatePasswordsSame(password1: string, password2: string): boolean {
    return password1 === password2;
}