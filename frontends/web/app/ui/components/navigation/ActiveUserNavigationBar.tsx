import React from 'react'
import styled from 'styled-components'
import { RankingRegistration } from '../../../ranking/interfaces'
import SubmitPagesLink from '../../../ranking/components/navigation/SubmitPages'
import SignOutLink from '../../../session/components/navigation/LogOut'
import { User } from '../../../user/interfaces'
import UserProfile from '../../../session/components/navigation/UserProfile'

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
    <LinkContainer>
      <SubmitPagesLink
        registration={registration}
        refreshRanking={refreshRanking}
      />
    </LinkContainer>
    <UserProfileContainer>
      <UserProfile user={user} />
      <SignOutLink />
    </UserProfileContainer>
  </>
)

const LinkContainer = styled.div`
  display: flex;
  margin-right: 20px;
  padding-right: 20px;
  border-right: 1px solid rgba(0, 0, 0, 0.1);
`

const UserProfileContainer = styled.div`
  display: flex;
  align-items: center;

  * + * {
    margin-left: 20px;
  }
`
