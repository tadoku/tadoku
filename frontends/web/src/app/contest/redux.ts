import { Contest } from './interfaces'
import { createSlice, PayloadAction } from '@reduxjs/toolkit'

const initialState = {
  latestContest: undefined as Contest | undefined,
}

const slice = createSlice({
  name: 'contest',
  initialState,
  reducers: {
    updateLatestContest(state, action: PayloadAction<Contest | undefined>) {
      state.latestContest = action.payload
    },
  },
})

export const { updateLatestContest } = slice.actions

export const contestInitialState = initialState

export default slice.reducer
