import React from 'react'
import { RankingRegistration } from '../../../ranking/interfaces'
import SubmitPagesLink from '../../../ranking/navigation/SubmitPages'
import SignOutLink from '../../../session/navigation/LogOut'

interface Props {
  registration: RankingRegistration | undefined
  refreshRanking: () => void
}

export const ActiveUserNavigationBar = ({
  registration,
  refreshRanking,
}: Props) => (
  <>
    <SubmitPagesLink
      registration={registration}
      refreshRanking={refreshRanking}
    />
    <SignOutLink />
  </>
)
