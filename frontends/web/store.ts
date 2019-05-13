import { createStore, combineReducers } from 'redux'
import {
  initialState as sessionInitialState,
  SessionAction,
  SessionActionTypes,
  sessionReducer,
} from './app/session/redux'

const initialState = {
  session: sessionInitialState,
}

export type State = typeof initialState
export type Action = SessionAction

export const actionTypes = {
  ...SessionActionTypes,
}

export const reducer = combineReducers({ session: sessionReducer })

export function initializeStore(state = initialState) {
  return createStore(reducer, state)
}
