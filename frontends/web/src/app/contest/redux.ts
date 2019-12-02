import { Contest } from './interfaces'

export const initialState = {
  latestContest: undefined as Contest | undefined,
}

// Actions

export enum ActionTypes {
  ContestUpdateLatestContest = '@contest/update-latest-contest',
}

export interface ContestUpdateLatestContest {
  type: typeof ActionTypes.ContestUpdateLatestContest
  payload: {
    latestContest: Contest | undefined
  }
}

export type Action = ContestUpdateLatestContest

// REDUCER

export const reducer = (state = initialState, action: Action) => {
  switch (action.type) {
    case ActionTypes.ContestUpdateLatestContest:
      const payload = (action as ContestUpdateLatestContest).payload
      return { ...state, latestContest: payload.latestContest }
    default:
      return state
  }
}
