import { useEffect, useState, type SubmitEvent } from 'react'
import { Button } from '@/components/ui/shadcn/button'
import { Input } from '@/components/ui/shadcn/input'
import { useNavigate } from 'react-router-dom'
import { USER_ROLES, type UserProfile } from '@/types/user'

interface CreatedDelegationLink {
  id: string
  status: string
}

export default function DashboardPage() {
  const navigate = useNavigate()
  const [user, setUser] = useState<UserProfile | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [delegateEmail, setDelegateEmail] = useState('')
  const [permissionsInput, setPermissionsInput] = useState('')
  const [validUntil, setValidUntil] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [delegationError, setDelegationError] = useState('')
  const [createdLink, setCreatedLink] = useState<CreatedDelegationLink | null>(null)

  useEffect(() => {
    const controller = new AbortController()

    async function loadProfile() {
      try {
        const response = await fetch('/api/me', {
          credentials: 'include',
          signal: controller.signal,
        })

        if (response.status === 401) {
          navigate('/login', { replace: true })
          return
        }
        if (!response.ok) {
          setError(`Could not load your profile (HTTP ${response.status}).`)
          return
        }

        const profile: UserProfile = await response.json()
        setUser(profile)
      } catch (err) {
        if (!controller.signal.aborted) {
          setError(err instanceof Error ? err.message : 'Could not load your profile.')
        }
      } finally {
        if (!controller.signal.aborted) {
          setLoading(false)
        }
      }
    }

    void loadProfile()
    return () => controller.abort()
  }, [navigate])

  if (loading) {
    return <main className="mx-auto max-w-2xl p-6" role="status">Loading your profile...</main>
  }

  if (error) {
    return <main className="mx-auto max-w-2xl p-6" role="alert">{error}</main>
  }

  if (!user) {
    return null
  }

  async function createDelegationLink(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault()
    setDelegationError('')
    setCreatedLink(null)
    setSubmitting(true)

    const permissions = permissionsInput
      .split(',')
      .map((permission) => permission.trim())
      .filter(Boolean)

    try {
      const response = await fetch('/api/delegation-links', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({
          delegate_email: delegateEmail.trim(),
          permissions,
          ...(validUntil && { valid_until: new Date(validUntil).toISOString() }),
        }),
      })

      if (!response.ok) {
        const message = (await response.text()).trim()
        setDelegationError(message || `Could not create delegation link (HTTP ${response.status}).`)
        return
      }

      const link: CreatedDelegationLink = await response.json()
      setCreatedLink(link)
      setDelegateEmail('')
      setPermissionsInput('')
      setValidUntil('')
    } catch (err) {
      setDelegationError(err instanceof Error ? err.message : 'Could not connect to the server.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <main className="mx-auto max-w-2xl space-y-8 p-6">
      <h1 className="text-2xl font-semibold">Your dashboard</h1>
      <dl className="space-y-2">
        <div><dt className="font-medium">Name</dt><dd>{user.Name}</dd></div>
        <div><dt className="font-medium">Email</dt><dd>{user.Email}</dd></div>
        <div><dt className="font-medium">Role</dt><dd>{user.Role}</dd></div>
        <div>
          <dt className="font-medium">Member since</dt>
          <dd>{new Date(user.CreatedAt).toLocaleDateString()}</dd>
        </div>
      </dl>

      {user.Role === USER_ROLES.OWNER && (
        <section className="space-y-4 rounded-lg border p-6">
          <div className="space-y-1">
            <h2 className="text-xl font-semibold">Create a delegation link</h2>
            <p className="text-sm text-muted-foreground">
              Invite a registered delegate by email. The link will be created as pending.
            </p>
          </div>

          <form onSubmit={createDelegationLink} className="space-y-4">
            <div className="space-y-2">
              <label htmlFor="delegate-email" className="text-sm font-medium">Delegate email</label>
              <Input
                id="delegate-email"
                name="delegate_email"
                type="email"
                autoComplete="email"
                required
                value={delegateEmail}
                onChange={(event) => setDelegateEmail(event.target.value)}
              />
            </div>

            <div className="space-y-2">
              <label htmlFor="permissions" className="text-sm font-medium">Permissions</label>
              <Input
                id="permissions"
                name="permissions"
                type="text"
                placeholder="view_balance, make_payment"
                required
                value={permissionsInput}
                onChange={(event) => setPermissionsInput(event.target.value)}
              />
              <p className="text-sm text-muted-foreground">
                Enter one or more permission names separated by commas.
              </p>
            </div>

            <div className="space-y-2">
              <label htmlFor="valid-until" className="text-sm font-medium">
                Expiration date <span className="font-normal text-muted-foreground">(optional)</span>
              </label>
              <Input
                id="valid-until"
                name="valid_until"
                type="datetime-local"
                value={validUntil}
                onChange={(event) => setValidUntil(event.target.value)}
              />
            </div>

            <Button type="submit" disabled={submitting}>
              {submitting ? 'Creating link...' : 'Create delegation link'}
            </Button>

            {delegationError && <p role="alert" className="text-sm text-destructive">{delegationError}</p>}
            {createdLink && (
              <p role="status" className="text-sm text-green-700 dark:text-green-400">
                Delegation link created with status “{createdLink.status}” (ID: {createdLink.id}).
              </p>
            )}
          </form>
        </section>
      )}
    </main>
  )
}
