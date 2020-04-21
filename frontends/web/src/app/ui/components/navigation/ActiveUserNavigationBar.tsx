import React from 'react'
import styled from 'styled-components'
import media from 'styled-media-query'
import Link from 'next/link'

import { RankingRegistration } from '../../../ranking/interfaces'
import { User } from '../../../session/interfaces'
import UserMenu from '../../../session/components/navigation/UserMenu'
import { Button } from '..'
import Constants from '../../Constants'

interface Props {
  user: User
  registration: RankingRegistration | undefined
}

export const ActiveUserNavigationBar = ({ user, registration }: Props) => (
  <>
    <LinkContainer>
      <Link href="/blog" passHref>
        <a href="">
          <Button plain>Blog</Button>
        </a>
      </Link>
      <Link href="/ranking" passHref>
        <a href="">
          <Button plain>Ranking</Button>
        </a>
      </Link>
      <Link href="/manual" passHref>
        <a href="">
          <Button plain>Manual</Button>
        </a>
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

const LinkContainer = styled.div`
  display: flex;
  padding-right: 20px;

  * + * {
    margin-left: 20px;

    ${media.lessThan('medium')`
      margin-left: 0;
    `}
  }

  > * {
    ${media.lessThan('medium')`
      margin-left: 0;
      padding-left: 30px;
      border-top: 1px solid ${Constants.colors.lightGray};
      display: block;
    `}
  }

  ${media.lessThan('medium')`
    border: none;
    margin: 0;
    padding: 0;
    flex-direction: column;
  `}
`

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
