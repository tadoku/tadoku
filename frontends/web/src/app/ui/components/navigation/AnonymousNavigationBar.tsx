import React from 'react'
import LogInLink from '../../../session/components/navigation/LogIn'
import RegisterLink from '../../../session/components/navigation/Register'

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
