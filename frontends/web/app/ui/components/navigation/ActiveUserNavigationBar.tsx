import React from 'react'
import { RankingRegistration } from '../../../ranking/interfaces'
import SubmitPagesLink from '../../../ranking/navigation/SubmitPages'
import SignOutLink from '../../../session/navigation/LogOut'

export const ActiveUserNavigationBar = ({
  registration,
}: {
  registration: RankingRegistration | undefined
}) => (
  <>
    <SubmitPagesLink registration={registration} />
    <SignOutLink />
  </>
)
