import { User } from './../user/interfaces'

export const initialState = {
  token: undefined as (string | undefined),
  user: undefined as (User | undefined),
}

// Actions

export enum ActionTypes {
  SessionSignIn = '@session/sign-in',
  SessionSignOut = '@session/sign-out',
}

export interface SessionSignIn {
  type: typeof ActionTypes.SessionSignIn
  payload: {
    token: string
    user: User
  }
}

export interface SessionSignOut {
  type: typeof ActionTypes.SessionSignOut
}

export type Action = SessionSignIn | SessionSignOut

// REDUCER

export const reducer = (state = initialState, action: Action) => {
  switch (action.type) {
    case ActionTypes.SessionSignIn:
      const payload = (action as SessionSignIn).payload
      return { ...state, token: payload.token, user: payload.user }
    case ActionTypes.SessionSignOut:
      return { ...state, token: undefined, user: undefined }
    default:
      return state
  }
}
