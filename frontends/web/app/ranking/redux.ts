import { RankingRegistration } from './interfaces'

export const initialState = {
  registration: undefined as (RankingRegistration | undefined),
  runEffectCount: 0,
}

// Actions

export enum ActionTypes {
  RankingUpdateRegistration = '@ranking/ranking-registration',
  RankingRunEffects = '@ranking/run-effects',
}

export interface RankingUpdateRegistration {
  type: typeof ActionTypes.RankingUpdateRegistration
  payload: {
    registration: RankingRegistration | undefined
  }
}

export interface RankingRunEffects {
  type: typeof ActionTypes.RankingRunEffects
}

export type Action = RankingUpdateRegistration | RankingRunEffects

// REDUCER

export const reducer = (state = initialState, action: Action) => {
  switch (action.type) {
    case ActionTypes.RankingUpdateRegistration:
      const payload = (action as RankingUpdateRegistration).payload
      return { ...state, registration: payload.registration }
    case ActionTypes.RankingRunEffects:
      return { ...state, runEffectCount: state.runEffectCount + 1 }
    default:
      return state
  }
}
