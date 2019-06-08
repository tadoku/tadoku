import React from 'react'
import LogInLink from '../../../session/navigation/LogIn'
import RegisterLink from '../../../session/navigation/Register'

export const AnonymousNavigationBar = ({
  refreshSession,
}: {
  refreshSession: () => void
}) => (
  <>
    <LogInLink refreshSession={refreshSession} />
    <RegisterLink refreshSession={refreshSession} />
  </>
)
