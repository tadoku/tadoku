import { configureStore, combineReducers } from '@reduxjs/toolkit'
import appReducer, { appInitialState } from './redux'
import contestReducer, { contestInitialState } from './contest/redux'
import * as RankingStore from './ranking/redux'
import * as SessionStore from './session/redux'

const initialState = {
  app: appInitialState,
  contest: contestInitialState,
  ranking: RankingStore.initialState,
  session: SessionStore.initialState,
}

export type State = typeof initialState
export type Action = RankingStore.Action | SessionStore.Action

export const actionTypes = {
  ...RankingStore.ActionTypes,
  ...SessionStore.ActionTypes,
}

export const reducer = combineReducers({
  app: appReducer,
  contest: contestReducer,
  ranking: RankingStore.reducer,
  session: SessionStore.reducer,
})

export function initializeStore(state = initialState) {
  return configureStore({
    reducer,
    preloadedState: state,
  })
}
