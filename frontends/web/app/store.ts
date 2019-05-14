import { createStore, combineReducers } from 'redux'
import {
  initialState as sessionInitialState,
  SessionAction,
  SessionActionTypes,
  sessionReducer,
} from './session/redux'
import {
  initialState as rankingInitialState,
  rankingReducer,
  RankingActionTypes,
  RankingAction,
} from './ranking/redux'

const initialState = {
  session: sessionInitialState,
  ranking: rankingInitialState,
}

export type State = typeof initialState
export type Action = SessionAction | RankingAction

export const actionTypes = {
  ...SessionActionTypes,
  ...RankingActionTypes,
}

export const reducer = combineReducers({
  session: sessionReducer,
  ranking: rankingReducer,
})

export function initializeStore(state = initialState) {
  return createStore(reducer, state)
}
