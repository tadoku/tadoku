import { Contest } from './interfaces'

export const initialState = {
  currentContest: undefined as (Contest | undefined),
}

// Actions

export enum ActionTypes {
  ContestUpdateCurrentContest = '@contest/update-current-contest',
}

export interface ContestUpdateCurrentContest {
  type: typeof ActionTypes.ContestUpdateCurrentContest
  payload: {
    currentContest: Contest | undefined
  }
}

export type Action = ContestUpdateCurrentContest

// REDUCER

export const reducer = (state = initialState, action: Action) => {
  switch (action.type) {
    case ActionTypes.ContestUpdateCurrentContest:
      const payload = (action as ContestUpdateCurrentContest).payload
      return { ...state, currentContest: payload.currentContest }
    default:
      return state
  }
}
