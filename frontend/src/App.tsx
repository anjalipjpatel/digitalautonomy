import AppNavigation from './components/AppNavigation'
import { ButtonLink } from './components/ui/button-link'
import DashboardPage from './pages/DashboardPage'
import LoginPage from './pages/LoginPage'
import SignUpPage from './pages/SignUpPage'
import { Route, Routes } from 'react-router-dom'

function App() {
  return (
    <>
      <AppNavigation />
      <Routes>
        <Route
          path="/"
          element={
            <main>
              <h1>Digital Autonomy</h1>
              <ButtonLink to="/login">Go to login</ButtonLink>
            </main>
          }
        />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/signup" element={<SignUpPage />} />
        <Route path="/dashboard" element={<DashboardPage />} />
      </Routes>
    </>
  )
}

export default App
