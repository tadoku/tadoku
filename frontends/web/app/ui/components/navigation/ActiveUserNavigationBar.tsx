import React from 'react'
import { RankingRegistration } from '../../../ranking/interfaces'
import SubmitPagesLink from '../../../ranking/navigation/SubmitPages'
import SignOutLink from '../../../session/navigation/LogOut'
import { User } from '../../../user/interfaces'
import { UserInNavigation } from '../UserInNavigation'

interface Props {
  user: User
  registration: RankingRegistration | undefined
  refreshRanking: () => void
}

export const ActiveUserNavigationBar = ({
  user,
  registration,
  refreshRanking,
}: Props) => (
  <>
    <SubmitPagesLink
      registration={registration}
      refreshRanking={refreshRanking}
    />
    <SignOutLink />
    <UserInNavigation user={user} />
  </>
)
