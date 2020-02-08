import { createSlice } from '@reduxjs/toolkit'

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
})

export const { startLoading, finishLoading } = slice.actions

export const appInitialState = initialState

export default slice.reducer
