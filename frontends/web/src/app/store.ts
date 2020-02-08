import { configureStore, combineReducers } from '@reduxjs/toolkit'
import appReducer, { appInitialState } from './redux'
import contestReducer, { contestInitialState } from './contest/redux'
import rankingReducer, { rankingInitialState } from './ranking/redux'
import sessionReducer, { sessionInitialState } from './session/redux'

const initialState = {
  app: appInitialState,
  contest: contestInitialState,
  ranking: rankingInitialState,
  session: sessionInitialState,
}

export type State = typeof initialState

export const reducer = combineReducers({
  app: appReducer,
  contest: contestReducer,
  ranking: rankingReducer,
  session: sessionReducer,
})

export function initializeStore(state = initialState) {
  return configureStore({
    reducer,
    preloadedState: state,
  })
}
