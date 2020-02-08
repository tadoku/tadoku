import { RankingRegistration } from './interfaces'
import { createSlice, PayloadAction } from '@reduxjs/toolkit'

export const initialState = {
  registration: undefined as RankingRegistration | undefined,
  runEffectCount: 0,
}

const slice = createSlice({
  name: 'ranking',
  initialState,
  reducers: {
    updateRegistration(
      state,
      action: PayloadAction<RankingRegistration | undefined>,
    ) {
      state.registration = action.payload
    },
    runEffects(state) {
      state.runEffectCount += 1
    },
  },
})

export const { updateRegistration, runEffects } = slice.actions

export const rankingInitialState = initialState

export default slice.reducer
