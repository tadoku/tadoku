import React from 'react'
import styled from 'styled-components'
import media from 'styled-media-query'
import Link from 'next/link'

import { RankingRegistration } from '../../../ranking/interfaces'
import { User } from '../../../session/interfaces'
import UserMenu from '../../../session/components/navigation/UserMenu'
import { ButtonLink } from '..'
import Constants from '../../Constants'
import LinkContainer from './LinkContainer'

interface Props {
  user: User
  registration: RankingRegistration | undefined
}

export const ActiveUserNavigationBar = ({ user, registration }: Props) => (
  <>
    <LinkContainer>
      <Link href="/blog" passHref>
        <ButtonLink plain>Blog</ButtonLink>
      </Link>
      <Link href="/ranking" passHref>
        <ButtonLink plain>Ranking</ButtonLink>
      </Link>
      <Link href="/manual" passHref>
        <ButtonLink plain>Manual</ButtonLink>
      </Link>
    </LinkContainer>
    <UserMenuContainer>
      <SmallContainer>
        <UserMenu.List user={user} registration={registration} />
      </SmallContainer>
      <LargeContainer>
        <UserMenu.Dropdown user={user} registration={registration} />
      </LargeContainer>
    </UserMenuContainer>
  </>
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
`
