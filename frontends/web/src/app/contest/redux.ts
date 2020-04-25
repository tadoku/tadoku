import { RawContest } from './interfaces'
import { createSlice, PayloadAction } from '@reduxjs/toolkit'

const initialState = {
  latestContest: undefined as RawContest | undefined,
  recentContests: [] as RawContest[],
}

const slice = createSlice({
  name: 'contest',
  initialState,
  reducers: {
    updateRecentContests(state, action: PayloadAction<RawContest[]>) {
      state.latestContest = action.payload[0]
      state.recentContests = action.payload
    },
  },
})

export const { updateRecentContests } = slice.actions

export const contestInitialState = initialState

export default slice.reducer
