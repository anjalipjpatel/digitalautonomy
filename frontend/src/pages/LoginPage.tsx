import { Input } from '@/components/ui/shadcn/input'
import { ButtonLink } from '../components/ui/button-link'

export default function LoginPage() {
  return (
    <>
      <h1>Login</h1>
      <ButtonLink to="/">
        Back to home
      </ButtonLink>


      <Input type='text' />
      <Input type='password' />
    </>
  )
}