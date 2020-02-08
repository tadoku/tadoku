import { RawRankingRegistration } from './interfaces'
import { createSlice, PayloadAction } from '@reduxjs/toolkit'

export const initialState = {
  rawRegistration: undefined as RawRankingRegistration | undefined,
  runEffectCount: 0,
}

const slice = createSlice({
  name: 'ranking',
  initialState,
  reducers: {
    updateRegistration(
      state,
      action: PayloadAction<RawRankingRegistration | undefined>,
    ) {
      state.rawRegistration = action.payload
    },
    runEffects(state) {
      state.runEffectCount += 1
    },
  },
})

export const { updateRegistration, runEffects } = slice.actions

export const rankingInitialState = initialState

export default slice.reducer
