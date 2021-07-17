import { RawContest } from './interfaces'
import { createSlice, PayloadAction } from '@reduxjs/toolkit'
import { HYDRATE } from 'next-redux-wrapper'

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
  extraReducers: {
    [HYDRATE]: (_, action) => action.payload.contest,
  },
})

export const { updateRecentContests } = slice.actions

export const contestInitialState = initialState

export default slice.reducer
