import { createStore } from 'redux'
import { User } from './domain/User'

const initialState = {
  token: undefined as (string | undefined),
  user: undefined as (User | undefined),
}

export type State = typeof initialState

// Actions

export enum AppActionTypes {
  AppReset = '@app/reset',
}

export enum SessionActionTypes {
  SessionSignIn = '@session/sign-in',
}

export interface AppReset {
  type: typeof AppActionTypes.AppReset
}

export interface SessionSignIn {
  type: typeof SessionActionTypes.SessionSignIn
  payload: {
    token: string
    user: User
  }
}

export type SessionActions = SessionSignIn
export type AppActions = AppReset
export type Action = AppActions | SessionActions

export const actionTypes = {
  ...AppActionTypes,
  ...SessionActionTypes,
}

// REDUCERS
export const reducer = (state = initialState, action: Action) => {
  switch (action.type) {
    case actionTypes.AppReset:
      return { ...initialState }
    case actionTypes.SessionSignIn:
      const payload = (action as SessionSignIn).payload
      return { ...state, token: payload.token, user: payload.user }
    default:
      return state
  }
}

export function initializeStore(state = initialState) {
  return createStore(reducer, state)
}
