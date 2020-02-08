import { createSlice } from '@reduxjs/toolkit'

const initialState = {
  isLoading: false,
  activityCount: 0,
}

const appSlice = createSlice({
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

export const { startLoading, finishLoading } = appSlice.actions

export const appInitialState = initialState

export default appSlice.reducer
