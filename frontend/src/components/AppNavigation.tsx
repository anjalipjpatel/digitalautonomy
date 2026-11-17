import { Link, useLocation } from 'react-router-dom'
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
      </nav>
    </header>
  )
}
