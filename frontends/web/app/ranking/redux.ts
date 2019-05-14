import { RankingRegistration } from './interfaces'

export const initialState = {
  registration: undefined as (RankingRegistration | undefined),
}

// Actions

export enum ActionTypes {
  RankingUpdateRegistration = '@ranking/ranking-registration',
}

export interface RankingUpdateRegistration {
  type: typeof ActionTypes.RankingUpdateRegistration
  payload: {
    registration: RankingRegistration | undefined
  }
}

export type Action = RankingUpdateRegistration

// REDUCER

export const reducer = (state = initialState, action: Action) => {
  switch (action.type) {
    case ActionTypes.RankingUpdateRegistration:
      const payload = (action as RankingUpdateRegistration).payload
      return { ...state, registration: payload.registration }
    default:
      return state
  }
}
