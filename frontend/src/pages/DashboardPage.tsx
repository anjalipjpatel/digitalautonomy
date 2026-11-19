import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import type { UserProfile } from '@/types/user'

export default function DashboardPage() {
  const navigate = useNavigate()
  const [user, setUser] = useState<UserProfile | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

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

  return (
    <main className="mx-auto max-w-2xl space-y-4 p-6">
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
    </main>
  )
}
