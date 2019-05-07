import { createStore } from 'redux'

const initialState = {
  open: false,
}

export type State = typeof initialState

// Actions

export enum AppActionTypes {
  AppReset = '@app/reset',
  AppOpen = '@app/open',
}

export interface AppReset {
  type: typeof AppActionTypes.AppReset
}

export interface AppOpen {
  type: typeof AppActionTypes.AppOpen
}

export type AppActions = AppReset | AppOpen
export type Action = AppActions

export const actionTypes = {
  ...AppActionTypes,
}

// REDUCERS
export const reducer = (state = initialState, action: Action) => {
  switch (action.type) {
    case actionTypes.AppReset:
      return { ...initialState }
    case actionTypes.AppOpen:
      return { ...state, open: true }
    default:
      return state
  }
}

export function initializeStore(state = initialState) {
  return createStore(reducer, state)
}
