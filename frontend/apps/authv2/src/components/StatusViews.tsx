import * as stylex from '@stylexjs/stylex'
import { buttonStyles } from 'paper-ui'
import { useLocation, useRouteContext } from '@tanstack/react-router'
import { CardPage } from './CardPage'
import { cardStyles } from './styles'

export function Unavailable() {
  const location = useLocation({ select: current => current.href })
  return (
    <CardPage
      title="We can’t reach your account right now"
      lead="Tadoku’s account service didn’t answer. Your account is safe; try again in a moment."
    >
      <a {...stylex.props(...buttonStyles({ fullWidth: true }))} href={location}>Try again</a>
    </CardPage>
  )
}

export function NotFound() {
  return (
    <CardPage title="Page not found" lead="This account page doesn’t exist. It may have moved.">
      <a {...stylex.props(...buttonStyles({ variant: 'outline', fullWidth: true }))} href="/">Go to your account</a>
    </CardPage>
  )
}

/** Shown after a deletion request, whoever is signed in now and whether or not the session API answers. */
export function Deleted() {
  const webUrl = useRouteContext({ from: '__root__', select: context => context.identity.webUrl })
  return (
    <CardPage title="Your account is deleted" lead="You’ve been logged out, and you can’t log in to this account again.">
      <div {...stylex.props(cardStyles.confirmation, cardStyles.spaced)}>
        <p {...stylex.props(cardStyles.confirmationText)}>We’re removing its data now; this finishes shortly.</p>
        <p {...stylex.props(cardStyles.confirmationText)}>We’ve sent a notice to your old email address.</p>
      </div>
      <a {...stylex.props(...buttonStyles({ variant: 'outline', fullWidth: true }))} href={webUrl}>Go to Tadoku</a>
    </CardPage>
  )
}
