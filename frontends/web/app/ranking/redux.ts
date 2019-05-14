import { RankingRegistration } from './interfaces'

export const initialState = {
  registration: undefined as (RankingRegistration | undefined),
}

// Actions

export enum RankingActionTypes {
  RankingUpdateRegistration = '@ranking/ranking-registration',
}

export interface RankingUpdateRegistration {
  type: typeof RankingActionTypes.RankingUpdateRegistration
  payload: {
    registration: RankingRegistration | undefined
  }
}

export type RankingAction = RankingUpdateRegistration

// REDUCER

export const rankingReducer = (state = initialState, action: RankingAction) => {
  switch (action.type) {
    case RankingActionTypes.RankingUpdateRegistration:
      const payload = (action as RankingUpdateRegistration).payload
      return { ...state, registration: payload.registration }
    default:
      return state
  }
}
