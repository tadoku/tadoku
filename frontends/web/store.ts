import { createStore } from 'redux'

const initialState = {}

export type State = typeof initialState

// Actions

export enum AppActionTypes {
  AppReset = '@app/reset',
}

export interface AppReset {
  type: typeof AppActionTypes.AppReset
}

export type AppActions = AppReset
export type Action = AppActions

export const actionTypes = {
  ...AppActionTypes,
}

// REDUCERS
export const reducer = (state = initialState, action: Action) => {
  switch (action.type) {
    case actionTypes.AppReset:
      return { ...initialState }
    default:
      return state
  }
}

export function initializeStore(state = initialState) {
  return createStore(reducer, state)
}
