export const initialState = {
  isLoading: false,
  activityCount: 0,
}

// Actions

export enum ActionTypes {
  AppLoadingStart = '@app/loading-start',
  AppLoadingFinish = '@app/loading-finish',
}

export interface AppLoadingStart {
  type: typeof ActionTypes.AppLoadingStart
  payload: {
    count: number
  }
}

export interface AppLoadingFinish {
  type: typeof ActionTypes.AppLoadingFinish
  payload: {
    count: number
  }
}

export type Action = AppLoadingStart | AppLoadingFinish

// REDUCER

export const reducer = (state = initialState, action: Action) => {
  switch (action.type) {
    case ActionTypes.AppLoadingStart:
      const startPayload = (action as AppLoadingStart).payload
      const activityCountForStart = state.activityCount - startPayload.count

      return {
        ...state,
        activityCount: activityCountForStart,
        isLoading: activityCountForStart === 0,
      }
    case ActionTypes.AppLoadingFinish:
      const finishPayload = (action as AppLoadingFinish).payload
      const activityCountForFinish = state.activityCount - finishPayload.count

      return {
        ...state,
        activityCount: activityCountForFinish,
        isLoading: activityCountForFinish === 0,
      }
    default:
      return state
  }
}
