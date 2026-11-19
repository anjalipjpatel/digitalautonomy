import { useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { Button } from './ui/shadcn/button'
import {
  NavigationMenu,
  NavigationMenuItem,
  NavigationMenuLink,
  NavigationMenuList,
} from './ui/shadcn/navigation-menu'

function isActive(currentPath: string, pathname: string): boolean {
  return currentPath === pathname
}

export default function AppNavigation() {
  const { pathname } = useLocation()
  const navigate = useNavigate()
  const [logoutError, setLogoutError] = useState('')
  const [loggingOut, setLoggingOut] = useState(false)

  async function handleLogout() {
    setLogoutError('')
    setLoggingOut(true)

    try {
      const response = await fetch('/api/logout', {
        method: 'POST',
        credentials: 'include',
      })

      if (!response.ok) {
        setLogoutError((await response.text()) || 'Could not log out.')
        return
      }

      navigate('/login', { replace: true })
    } catch (err) {
      setLogoutError(err instanceof Error ? err.message : 'Could not log out.')
    } finally {
      setLoggingOut(false)
    }
  }

  return (
    <header className="dark border-b bg-background px-6 py-3 text-foreground">
      <nav aria-label="Navigation Menu" className="mx-auto flex items-center gap-8">
        <Link to="/" className="font-semibold text-white">Digital Autonomy</Link>
        <NavigationMenu className="max-w-none justify-start">
          <NavigationMenuList className="justify-start gap-2">
            <NavigationMenuItem>
              <NavigationMenuLink render={<Link to="/" />} active={isActive('/', pathname)}>Home</NavigationMenuLink>
            </NavigationMenuItem>
            <NavigationMenuItem>
              <NavigationMenuLink render={<Link to="/login" />} active={isActive('/login', pathname)}>Login</NavigationMenuLink>
            </NavigationMenuItem>
          </NavigationMenuList>
        </NavigationMenu>
        {pathname === '/dashboard' && (
          <Button type="button" variant="outline" disabled={loggingOut} onClick={handleLogout}>
            {loggingOut ? 'Logging out...' : 'Log out'}
          </Button>
        )}
      </nav>
      {logoutError && <p role="alert" className="mx-auto mt-2 max-w-5xl text-sm text-destructive">{logoutError}</p>}
    </header>
  )
}
