import { configureStore, combineReducers } from '@reduxjs/toolkit'
import appReducer, { appInitialState } from './redux'
import contestReducer, { contestInitialState } from './contest/redux'
import rankingReducer, { rankingInitialState } from './ranking/redux'
import * as SessionStore from './session/redux'

const initialState = {
  app: appInitialState,
  contest: contestInitialState,
  ranking: rankingInitialState,
  session: SessionStore.initialState,
}

export type State = typeof initialState
export type Action = SessionStore.Action

export const actionTypes = {
  ...SessionStore.ActionTypes,
}

export const reducer = combineReducers({
  app: appReducer,
  contest: contestReducer,
  ranking: rankingReducer,
  session: SessionStore.reducer,
})

export function initializeStore(state = initialState) {
  return configureStore({
    reducer,
    preloadedState: state,
  })
}
