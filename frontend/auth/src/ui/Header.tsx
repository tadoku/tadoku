import Link from 'next/link'
import { useLogoutHandler, useSession } from '../session'
import { Session } from '@ory/client'

const Header = () => {
  const [session] = useSession()

  if (session) {
    return <ActiveSessionHeader session={session} />
  }

  return (
    <>
      <Link href="/">
        <a>Home</a>
      </Link>
      <br />
      <Link href="/login">
        <a>Log in</a>
      </Link>
      <br />
      <Link href="/register">
        <a>Register</a>
      </Link>
      <br />
      <Link href="/account-recovery">
        <a>Forgot password?</a>
      </Link>
    </>
  )
}

const ActiveSessionHeader = ({ session }: { session: Session }) => {
  const onLogout = useLogoutHandler([session])

  return (
    <>
      <span>
        Hello, <strong>{session.identity.traits['display_name']}</strong>
      </span>
      <br />
      <Link href="/">
        <a>Settings</a>
      </Link>
      <br />
      <a href="#" onClick={onLogout}>
        Log out
      </a>
    </>
  )
}

export default Header
