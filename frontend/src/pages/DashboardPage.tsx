import {Navigate, useLocation } from 'react-router-dom'
import type { UserProfile } from '@/types/user'

export default function DashboardPage() {
    const location = useLocation()
    const user = (location.state as {user?: UserProfile} | null)?.user

    if (!user) {
        return <Navigate to="/login" replace />
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
