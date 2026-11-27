import { useEffect, useState, type SubmitEvent } from 'react'
import { Button } from '@/components/ui/shadcn/button'
import { Input } from '@/components/ui/shadcn/input'
import { useNavigate } from 'react-router-dom'
import { USER_ROLES, type UserProfile } from '@/types/user'

interface DelegationLink {
  id: string
  primary_user_id: string
  delegate_user_id: string
  valid_until?: string
  permissions: string[]
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
  const [createdLink, setCreatedLink] = useState<DelegationLink | null>(null)
  const [delegationLinks, setDelegationLinks] = useState<DelegationLink[]>([])
  const [delegationLinksError, setDelegationLinksError] = useState('')
  const [acceptingLinkID, setAcceptingLinkID] = useState('')
  const [acceptedMessage, setAcceptedMessage] = useState('')

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
        if (profile.Role === USER_ROLES.DELEGATE || profile.Role === USER_ROLES.OWNER) {
          try {
            const linksResponse = await fetch('/api/delegation-links', {
              credentials: 'include',
              signal: controller.signal,
            })
            if (!linksResponse.ok) {
              const message = (await linksResponse.text()).trim()
              setDelegationLinksError(message || `Could not load delegation links (HTTP ${linksResponse.status}).`)
            } else {
              setDelegationLinks(await linksResponse.json())
            }
          } catch (err) {
            if (!controller.signal.aborted) {
              setDelegationLinksError(err instanceof Error ? err.message : 'Could not load delegation links.')
            }
          }
        }
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

      const link: DelegationLink = await response.json()
      setCreatedLink(link)
      setDelegationLinks((links) => [link, ...links])
      setDelegateEmail('')
      setPermissionsInput('')
      setValidUntil('')
    } catch (err) {
      setDelegationError(err instanceof Error ? err.message : 'Could not connect to the server.')
    } finally {
      setSubmitting(false)
    }
  }

  async function acceptDelegationLink(linkID: string) {
    setDelegationLinksError('')
    setAcceptedMessage('')
    setAcceptingLinkID(linkID)

    try {
      const response = await fetch(`/api/delegation-links/${encodeURIComponent(linkID)}/accept`, {
        method: 'POST',
        credentials: 'include',
      })
      if (!response.ok) {
        const message = (await response.text()).trim()
        setDelegationLinksError(message || `Could not accept invitation (HTTP ${response.status}).`)
        return
      }

      const acceptedLink: DelegationLink = await response.json()
      setDelegationLinks((links) => links.map((link) => link.id === linkID ? acceptedLink : link))
      setAcceptedMessage('Delegation accepted.')
    } catch (err) {
      setDelegationLinksError(err instanceof Error ? err.message : 'Could not connect to the server.')
    } finally {
      setAcceptingLinkID('')
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

      {(user.Role === USER_ROLES.DELEGATE || user.Role === USER_ROLES.OWNER) && (
        <section className="space-y-4 rounded-lg border p-6">
          <div className="space-y-1">
            <h2 className="text-xl font-semibold">Delegation links</h2>
            <p className="text-sm text-muted-foreground">
              {user.Role === USER_ROLES.DELEGATE
                ? 'Links where you are the delegate. Accept pending invitations to activate them.'
                : 'Links where you are the primary user.'}
            </p>
          </div>

          {delegationLinksError && <p role="alert" className="text-sm text-destructive">{delegationLinksError}</p>}
          {acceptedMessage && <p role="status" className="text-sm text-green-700 dark:text-green-400">{acceptedMessage}</p>}
          {delegationLinks.length === 0 ? (
            <p className="text-sm text-muted-foreground">No delegation links found.</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full min-w-175 border-collapse text-left text-sm">
                <thead>
                  <tr className="border-b">
                    <th scope="col" className="px-3 py-2 font-medium">Link ID</th>
                    <th scope="col" className="px-3 py-2 font-medium">
                      {user.Role === USER_ROLES.OWNER ? 'Delegate user ID' : 'Primary user ID'}
                    </th>
                    <th scope="col" className="px-3 py-2 font-medium">Permissions</th>
                    <th scope="col" className="px-3 py-2 font-medium">Expires</th>
                    <th scope="col" className="px-3 py-2 font-medium">Status</th>
                    {user.Role === USER_ROLES.DELEGATE && <th scope="col" className="px-3 py-2 font-medium">Action</th>}
                  </tr>
                </thead>
                <tbody>
                  {delegationLinks.map((link) => (
                    <tr key={link.id} className="border-b last:border-0">
                      <td className="px-3 py-3 font-mono text-xs">{link.id}</td>
                      <td className="px-3 py-3 font-mono text-xs">
                        {user.Role === USER_ROLES.OWNER ? link.delegate_user_id : link.primary_user_id}
                      </td>
                      <td className="px-3 py-3">{link.permissions.join(', ')}</td>
                      <td className="px-3 py-3">
                        {link.valid_until ? new Date(link.valid_until).toLocaleString() : 'No expiry'}
                      </td>
                      <td className="px-3 py-3 capitalize">{link.status}</td>
                      {user.Role === USER_ROLES.DELEGATE && (
                        <td className="px-3 py-3">
                          {link.status === 'pending' ? (
                            <Button
                              type="button"
                              disabled={acceptingLinkID !== ''}
                              onClick={() => void acceptDelegationLink(link.id)}
                            >
                              {acceptingLinkID === link.id ? 'Accepting...' : 'Accept'}
                            </Button>
                          ) : '—'}
                        </td>
                      )}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>
      )}
    </main>
  )
}
