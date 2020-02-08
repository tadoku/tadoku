import React, { useEffect, useState } from 'react'
import styled from 'styled-components'
import media from 'styled-media-query'
import { connect } from 'react-redux'
import { State } from '../../../store'
import { User } from '../../../session/interfaces'
import { RankingRegistration } from '../../../ranking/interfaces'
import { ActiveUserNavigationBar } from './ActiveUserNavigationBar'
import { AnonymousNavigationBar } from './AnonymousNavigationBar'
import { runEffects as sessionRunEffects } from '../../../session/redux'
import { runEffects as rankingRunEffects } from '../../../ranking/redux'

interface Props {
  user: User | undefined
  registration: RankingRegistration | undefined
  refreshSession: () => void
  refreshRanking: () => void
}

const NavigationBar = ({
  user,
  registration,
  refreshSession,
  refreshRanking,
}: Props) => {
  const [hasMounted, setHasMounted] = useState(false)

  useEffect(() => {
    setHasMounted(true)
  }, [])

  if (!hasMounted) {
    return null
  }

  return (
    <StyledNav prefersVertical={!!user}>
      {user ? (
        <ActiveUserNavigationBar
          registration={registration}
          refreshRanking={refreshRanking}
          user={user}
        />
      ) : (
        <AnonymousNavigationBar refreshSession={refreshSession} />
      )}
    </StyledNav>
  )
}

const mapStateToProps = (state: State) => ({
  user: state.session.user,
  registration: state.ranking.registration,
})

const mapDispatchToProps = {
  refreshSession: sessionRunEffects,
  refreshRanking: rankingRunEffects,
}

export default connect(mapStateToProps, mapDispatchToProps)(NavigationBar)

const StyledNav = styled.nav`
  display: flex;
  align-items: center;

  ${({ prefersVertical }: { prefersVertical?: boolean }) =>
    prefersVertical &&
    media.lessThan('medium')`
      flex-direction: column-reverse;
    `}
`
