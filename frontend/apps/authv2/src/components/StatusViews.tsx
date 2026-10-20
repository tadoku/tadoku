import { useLocation } from '@tanstack/react-router'
import { buttonClassName } from 'paper-ui'
import { CardPage } from './CardPage'

export function Unavailable() {
  const location = useLocation({ select: current => current.href })
  return (
    <CardPage
      title="We can’t reach your account right now"
      lead="Tadoku’s account service didn’t answer. Your account is safe; try again in a moment."
    >
      <a className={buttonClassName({ fullWidth: true })} href={location}>Try again</a>
    </CardPage>
  )
}

export function NotFound() {
  return (
    <CardPage title="Page not found" lead="This account page doesn’t exist. It may have moved.">
      <a className={buttonClassName({ variant: 'outline', fullWidth: true })} href="/">Go to your account</a>
    </CardPage>
  )
}
