import { User } from './interfaces'
import { createSlice, PayloadAction } from '@reduxjs/toolkit'

export const initialState = {
  token: undefined as string | undefined,
  user: undefined as User | undefined,
  loaded: false,
  runEffectCount: 0,
}

const slice = createSlice({
  name: 'session',
  initialState,
  reducers: {
    logIn(state, action: PayloadAction<{ token: string; user: User }>) {
      state.token = action.payload.token
      state.user = action.payload.user
      state.loaded = true
      debugger
    },
    logOut(state) {
      state.token = undefined
      state.user = undefined
      state.runEffectCount += 1
    },
    runEffects(state) {
      state.runEffectCount += 1
    },
  },
})

export const { logIn, logOut, runEffects } = slice.actions

export const sessionInitialState = initialState

export default slice.reducer
