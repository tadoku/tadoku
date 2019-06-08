import React, { useEffect, useState } from 'react'
import styled from 'styled-components'
import { connect } from 'react-redux'
import { State } from '../../../store'
import { User } from '../../../user/interfaces'
import { RankingRegistration } from '../../../ranking/interfaces'
import { ActiveUserNavigationBar } from './ActiveUserNavigationBar'
import { AnonymousNavigationBar } from './AnonymousNavigationBar'
import { Dispatch } from 'redux'
import * as SessionStore from '../../../session/redux'

const StyledNav = styled.nav`
  display: flex;
  align-items: center;
`

interface Props {
  user: User | undefined
  registration: RankingRegistration | undefined
  refreshSession: () => void
}

const NavigationBar = ({ user, registration, refreshSession }: Props) => {
  const [hasMounted, setHasMounted] = useState(false)

  useEffect(() => {
    setHasMounted(true)
  }, [])

  if (!hasMounted) {
    return null
  }

  return (
    <StyledNav>
      {user ? (
        <ActiveUserNavigationBar
          registration={registration}
          refreshSession={refreshSession}
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

const mapDispatchToProps = (dispatch: Dispatch<SessionStore.Action>) => ({
  refreshSession: () => {
    dispatch({
      type: SessionStore.ActionTypes.SessionRunEffects,
    })
  },
})

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(NavigationBar)
