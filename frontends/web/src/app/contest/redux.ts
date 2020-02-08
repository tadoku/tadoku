import { RawContest } from './interfaces'
import { createSlice, PayloadAction } from '@reduxjs/toolkit'

const initialState = {
  latestContest: undefined as RawContest | undefined,
}

const slice = createSlice({
  name: 'contest',
  initialState,
  reducers: {
    updateLatestContest(state, action: PayloadAction<RawContest | undefined>) {
      state.latestContest = action.payload
    },
  },
})

export const { updateLatestContest } = slice.actions

export const contestInitialState = initialState

export default slice.reducer
