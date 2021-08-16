import { createSlice } from '@reduxjs/toolkit'
import { HYDRATE } from 'next-redux-wrapper'

const initialState = {
  isLoading: false,
  activityCount: 0,
}

const slice = createSlice({
  name: 'app',
  initialState,
  reducers: {
    startLoading(state) {
      state.activityCount += 1
      state.isLoading = state.activityCount > 0
    },
    finishLoading(state) {
      state.activityCount -= 1
      state.isLoading = state.activityCount > 0
    },
  },
  extraReducers: {
    [HYDRATE]: (_, action) => action.payload.app,
  },
})

export const { startLoading, finishLoading } = slice.actions

export const appInitialState = initialState

export default slice.reducer
