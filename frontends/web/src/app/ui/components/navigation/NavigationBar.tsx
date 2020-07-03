import React, { useEffect, useState } from 'react'
import styled from 'styled-components'
import media from 'styled-media-query'
import { connect } from 'react-redux'
import { RootState } from '../../../store'
import { User } from '@app/session/interfaces'
import { RankingRegistration } from '@app/ranking/interfaces'
import { ActiveUserNavigationBar } from './ActiveUserNavigationBar'
import { AnonymousNavigationBar } from './AnonymousNavigationBar'
import { runEffects as sessionRunEffects } from '@app/session/redux'
import { runEffects as rankingRunEffects } from '@app/ranking/redux'
import { rankingRegistrationMapper } from '@app/ranking/transform/ranking-registration'
import LinkContainer from './LinkContainer'
import Link from 'next/link'
import { ButtonLink } from '..'
import { useRouter } from 'next/router'

interface Props {
  user: User | undefined
  registration: RankingRegistration | undefined
  refreshSession: () => void
  isOpen: boolean
  closeNavigation: () => void
}

const NavigationBar = ({
  user,
  registration,
  refreshSession,
  isOpen,
  closeNavigation,
}: Props) => {
  const [hasMounted, setHasMounted] = useState(false)
  const router = useRouter()

  useEffect(() => {
    setHasMounted(true)
  }, [])

  if (!hasMounted) {
    return null
  }

  return (
    <StyledNav isOpen={isOpen}>
      <LinkContainer onClick={closeNavigation}>
        <Link href="/blog" passHref>
          <ButtonLink plain active={router.pathname == '/blog'}>
            Blog
          </ButtonLink>
        </Link>
        <Link href="/ranking" passHref>
          <ButtonLink plain active={router.pathname == '/ranking'}>
            Ranking
          </ButtonLink>
        </Link>
        <Link href="/manual" passHref>
          <ButtonLink plain active={router.pathname == '/manual'}>
            Manual
          </ButtonLink>
        </Link>
        {user ? (
          <ActiveUserNavigationBar registration={registration} user={user} />
        ) : (
          <AnonymousNavigationBar refreshSession={refreshSession} />
        )}
      </LinkContainer>
    </StyledNav>
  )
}

const mapStateToProps = (state: RootState) => ({
  user: state.session.user,
  registration: rankingRegistrationMapper.optional.fromRaw(
    state.ranking.rawRegistration,
  ),
})

const mapDispatchToProps = {
  refreshSession: sessionRunEffects,
  refreshRanking: rankingRunEffects,
}

export default connect(mapStateToProps, mapDispatchToProps)(NavigationBar)

const StyledNav = styled.nav<{ isOpen: boolean }>`
  display: flex;
  align-items: center;

  ${({ isOpen }) => media.lessThan('medium')`
      display: ${isOpen ? 'block' : 'none'};
      margin: 30px -30px -30px -30px;
      width: calc(100% + 60px);
  `}
`
