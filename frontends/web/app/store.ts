import { createStore, combineReducers } from 'redux'
import {
  initialState as sessionInitialState,
  SessionAction,
  SessionActionTypes,
  sessionReducer,
} from './session/redux'
import * as RankingStore from './ranking/redux'

const initialState = {
  session: sessionInitialState,
  ranking: RankingStore.initialState,
}

export type State = typeof initialState
export type Action = SessionAction | RankingStore.Action

export const actionTypes = {
  ...SessionActionTypes,
  ...RankingStore.ActionTypes,
}

export const reducer = combineReducers({
  session: sessionReducer,
  ranking: RankingStore.reducer,
})

export function initializeStore(state = initialState) {
  return createStore(reducer, state)
}
