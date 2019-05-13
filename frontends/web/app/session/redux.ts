import { User } from './../user/interfaces'

export const initialState = {
  token: undefined as (string | undefined),
  user: undefined as (User | undefined),
}

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

export type SessionAction = SessionSignIn | SessionSignOut

// REDUCER

export const sessionReducer = (state = initialState, action: SessionAction) => {
  switch (action.type) {
    case SessionActionTypes.SessionSignIn:
      const payload = (action as SessionSignIn).payload
      return { ...state, token: payload.token, user: payload.user }
    case SessionActionTypes.SessionSignOut:
      return { ...state, token: undefined, user: undefined }
    default:
      return state
  }
}
