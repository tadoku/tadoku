import React, { useEffect, useState } from 'react'
import styled from 'styled-components'
import media from 'styled-media-query'
import { connect } from 'react-redux'
import { RootState } from '../../../store'
import { User } from '../../../session/interfaces'
import { RankingRegistration } from '../../../ranking/interfaces'
import { ActiveUserNavigationBar } from './ActiveUserNavigationBar'
import { AnonymousNavigationBar } from './AnonymousNavigationBar'
import { runEffects as sessionRunEffects } from '../../../session/redux'
import { runEffects as rankingRunEffects } from '../../../ranking/redux'
import { RankingRegistrationMapper } from '../../../ranking/transform/ranking-registration'

interface Props {
  user: User | undefined
  registration: RankingRegistration | undefined
  refreshSession: () => void
  isOpen: boolean
}

const NavigationBar = ({
  user,
  registration,
  refreshSession,
  isOpen,
}: Props) => {
  const [hasMounted, setHasMounted] = useState(false)

  useEffect(() => {
    setHasMounted(true)
  }, [])

  if (!hasMounted) {
    return null
  }

  return (
    <StyledNav isOpen={isOpen}>
      {user ? (
        <ActiveUserNavigationBar registration={registration} user={user} />
      ) : (
        <AnonymousNavigationBar refreshSession={refreshSession} />
      )}
    </StyledNav>
  )
}

const mapStateToProps = (state: RootState) => ({
  user: state.session.user,
  registration: RankingRegistrationMapper.optional.fromRaw(
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
      margin-top: 40px;
  `}
`
