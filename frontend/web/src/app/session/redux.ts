import { User } from './interfaces'
import { createSlice, PayloadAction } from '@reduxjs/toolkit'
import SessionApi from './api'
import { HYDRATE } from 'next-redux-wrapper'

export const initialState = {
  expiresAt: undefined as number | undefined,
  user: undefined as User | undefined,
  loaded: false,
  runEffectCount: 0,
}

const slice = createSlice({
  name: 'session',
  initialState,
  reducers: {
    logIn(state, action: PayloadAction<{ expiresAt: number; user: User }>) {
      state.expiresAt = action.payload.expiresAt
      state.user = action.payload.user
      state.loaded = true
    },
    logOut(state) {
      state.expiresAt = undefined
      state.user = undefined
      state.runEffectCount += 1
      SessionApi.logOut()
    },
    runEffects(state) {
      state.runEffectCount += 1
    },
  },
  extraReducers: {
    [HYDRATE]: (_, action) => action.payload.session,
  },
})

export const { logIn, logOut, runEffects } = slice.actions

export const sessionInitialState = initialState

export default slice.reducer
