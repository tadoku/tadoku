import React from 'react'
import styled from 'styled-components'
import media from 'styled-media-query'
import Link from 'next/link'

import { RankingRegistration } from '../../../ranking/interfaces'
import SubmitPagesLink from '../../../ranking/components/navigation/SubmitPages'
import { User } from '../../../session/interfaces'
import UserMenu from '../../../session/components/navigation/UserMenu'
import { Button } from '..'

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
      <Link href="/about" passHref>
        <a href="">
          <Button plain>About</Button>
        </a>
      </Link>
      <SubmitPagesLink
        registration={registration}
        refreshRanking={refreshRanking}
      />
    </LinkContainer>
    <UserMenuContainer>
      <UserMenu user={user} registration={registration} />
    </UserMenuContainer>
  </>
)

const LinkContainer = styled.div`
  display: flex;
  padding-right: 20px;

  ${media.lessThan('medium')`
    border: none;
    margin: 0;
    padding: 0;
    flex-direction: column;
  `}

  * + * {
    margin-left: 20px;
  }
`

const UserMenuContainer = styled.div`
  display: flex;
  align-items: center;

  ${media.lessThan('medium')`
    margin: 5px 0 15px;
    padding: 4px 10px;
    border-radius: 2px;
    box-shadow: 4px 5px 15px 1px rgba(0, 0, 0, 0.08);
  `}
`
