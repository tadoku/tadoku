import React from 'react'
import styled from 'styled-components'
import media from 'styled-media-query'
import Link from 'next/link'

import { RankingRegistration } from '../../../ranking/interfaces'
import { User } from '../../../session/interfaces'
import UserMenu from '../../../session/components/navigation/UserMenu'
import { Button } from '..'

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
      <UserMenu user={user} registration={registration} />
    </UserMenuContainer>
  </>
)

const LinkContainer = styled.div`
  display: flex;
  padding-right: 20px;

  * + * {
    ${media.greaterThan('medium')`
      margin-left: 20px;
    `}
  }

  ${media.lessThan('medium')`
    border: none;
    margin: 0;
    padding: 0;
    flex-direction: column;
  `}
`

const UserMenuContainer = styled.div`
  display: flex;
  align-items: center;

  ${media.lessThan('medium')`
    margin: 0 5px;
  `}
`
