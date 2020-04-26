import React from 'react'
import styled from 'styled-components'
import media from 'styled-media-query'

import { RankingRegistration } from '../../../ranking/interfaces'
import { User } from '../../../session/interfaces'
import UserMenu from '../../../session/components/navigation/UserMenu'
import Constants from '../../Constants'

interface Props {
  user: User
  registration: RankingRegistration | undefined
}

export const ActiveUserNavigationBar = ({ user, registration }: Props) => (
  <UserMenuContainer>
    <SmallContainer>
      <UserMenu.List user={user} registration={registration} />
    </SmallContainer>
    <LargeContainer>
      <UserMenu.Dropdown user={user} registration={registration} />
    </LargeContainer>
  </UserMenuContainer>
)

const LargeContainer = styled.div`
  ${media.lessThan('medium')`
    display: none;
  `}
`

const SmallContainer = styled.div`
  display: none;

  ${media.lessThan('medium')`
    display: block;
    border-top: 2px solid ${Constants.colors.lightGray};
    width: 100%;
  `}
`

const UserMenuContainer = styled.div`
  display: flex;
  align-items: center;
  ${media.lessThan('medium')`
      padding-left: 0 !important;
  `}
`
