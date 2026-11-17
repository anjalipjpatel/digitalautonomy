import './App.css'
import LoginPage from './pages/LoginPage'

function App() {
  if (window.location.pathname === '/login') {
    return <LoginPage />
  }

  return (
    <main>
      <h1>Welcome</h1>
      <a className="button" href="/login">
        Go to login
      </a>
    </main>
  )
}

export default App
