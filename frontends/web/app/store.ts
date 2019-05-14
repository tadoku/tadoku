import { createStore, combineReducers } from 'redux'
import * as SessionStore from './session/redux'
import * as RankingStore from './ranking/redux'

const initialState = {
  session: SessionStore.initialState,
  ranking: RankingStore.initialState,
}

export type State = typeof initialState
export type Action = SessionStore.Action | RankingStore.Action

export const actionTypes = {
  ...SessionStore.ActionTypes,
  ...RankingStore.ActionTypes,
}

export const reducer = combineReducers({
  session: SessionStore.reducer,
  ranking: RankingStore.reducer,
})

export function initializeStore(state = initialState) {
  return createStore(reducer, state)
}
