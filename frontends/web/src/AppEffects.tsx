import React from 'react'
import SessionEffects from './session/components/Effects'
import RankingEffects from './ranking/components/Effects'
import ContestEffects from './contest/components/Effects'

const AppEffects = () => (
  <>
    <SessionEffects />
    <RankingEffects />
    <ContestEffects />
  </>
)

export default AppEffects
