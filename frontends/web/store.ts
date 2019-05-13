import { createStore, combineReducers } from 'redux'
import { User } from './app/user/interfaces'

const sessionInitialState = {
  token: undefined as (string | undefined),
  user: undefined as (User | undefined),
}

const initialState = {
  session: sessionInitialState,
}

export type State = typeof initialState

// Actions

export enum SessionActionTypes {
  SessionSignIn = '@session/sign-in',
  SessionSignOut = '@session/sign-out',
}

export interface SessionSignIn {
  type: typeof SessionActionTypes.SessionSignIn
  payload: {
    token: string
    user: User
  }
}

export interface SessionSignOut {
  type: typeof SessionActionTypes.SessionSignOut
}

export type SessionActions = SessionSignIn | SessionSignOut
export type Action = SessionActions

export const actionTypes = {
  ...SessionActionTypes,
}

// REDUCERS

export const sessionReducer = (state = initialState, action: Action) => {
  switch (action.type) {
    case actionTypes.SessionSignIn:
      const payload = (action as SessionSignIn).payload
      return { ...state, token: payload.token, user: payload.user }
    case actionTypes.SessionSignOut:
      return { ...state, token: undefined, user: undefined }
    default:
      return state
  }
}

export const reducer = combineReducers({ session: sessionReducer })

export function initializeStore(state = initialState) {
  return createStore(reducer, state)
}
