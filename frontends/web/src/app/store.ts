import { createStore, combineReducers } from 'redux'
import * as AppStore from './redux'
import * as ContestStore from './contest/redux'
import * as RankingStore from './ranking/redux'
import * as SessionStore from './session/redux'

const initialState = {
  app: AppStore.initialState,
  contest: ContestStore.initialState,
  ranking: RankingStore.initialState,
  session: SessionStore.initialState,
}

export type State = typeof initialState
export type Action =
  | AppStore.Action
  | ContestStore.Action
  | RankingStore.Action
  | SessionStore.Action

export const actionTypes = {
  ...AppStore.ActionTypes,
  ...ContestStore.ActionTypes,
  ...RankingStore.ActionTypes,
  ...SessionStore.ActionTypes,
}

export const reducer = combineReducers({
  app: AppStore.reducer,
  contest: ContestStore.reducer,
  ranking: RankingStore.reducer,
  session: SessionStore.reducer,
})

export function initializeStore(state = initialState) {
  return createStore(reducer, state)
}
