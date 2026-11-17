import { ButtonLink } from './components/ui/button-link'
import LoginPage from './pages/LoginPage'
import { Route, Routes } from 'react-router-dom'

function App() {
  return (
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
    </Routes>
  )
}

export default App
