import React from 'react'
import styled from 'styled-components'
import media from 'styled-media-query'
import { RankingRegistration } from '../../../ranking/interfaces'
import SubmitPagesLink from '../../../ranking/components/navigation/SubmitPages'
import SignOutLink from '../../../session/components/navigation/LogOut'
import { User } from '../../../session/interfaces'
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

  ${media.lessThan('medium')`
    border: none;
    margin: 0;
    padding: 0;
    flex-direction: column;
  `}
`

const UserProfileContainer = styled.div`
  display: flex;
  align-items: center;

  ${media.lessThan('medium')`
    margin-top: 20px;
    padding: 4px 10px;
    border-radius: 2px;
    box-shadow: 4px 5px 15px 1px rgba(0, 0, 0, 0.08);
  `}

  * + * {
    margin-left: 20px;
  }
`
