import { configureStore, combineReducers } from '@reduxjs/toolkit'
import AppReducer, { appInitialState } from './redux'
import * as ContestStore from './contest/redux'
import * as RankingStore from './ranking/redux'
import * as SessionStore from './session/redux'

const initialState = {
  app: appInitialState,
  contest: ContestStore.initialState,
  ranking: RankingStore.initialState,
  session: SessionStore.initialState,
}

export type State = typeof initialState
export type Action =
  | ContestStore.Action
  | RankingStore.Action
  | SessionStore.Action

export const actionTypes = {
  ...ContestStore.ActionTypes,
  ...RankingStore.ActionTypes,
  ...SessionStore.ActionTypes,
}

export const reducer = combineReducers({
  app: AppReducer,
  contest: ContestStore.reducer,
  ranking: RankingStore.reducer,
  session: SessionStore.reducer,
})

export function initializeStore(state = initialState) {
  return configureStore({
    reducer,
    preloadedState: state,
  })
}
