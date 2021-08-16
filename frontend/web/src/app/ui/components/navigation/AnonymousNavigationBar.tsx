import React from 'react'

import LogInLink from '@app/session/components/navigation/LogIn'
import RegisterLink from '@app/session/components/navigation/Register'

export const AnonymousNavigationBar = ({
  refreshSession,
}: {
  refreshSession: () => void
}) => {
  return (
    <>
      <LogInLink refreshSession={refreshSession} />
      <RegisterLink refreshSession={refreshSession} />
    </>
  )
}
