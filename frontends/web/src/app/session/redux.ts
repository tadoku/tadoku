import { User } from './interfaces'
import { createSlice, PayloadAction } from '@reduxjs/toolkit'
import { storeUserInLocalStorage, removeUserFromLocalStorage } from './storage'
import SessionApi from './api'

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
      storeUserInLocalStorage(action.payload)
      state.expiresAt = action.payload.expiresAt
      state.user = action.payload.user
      state.loaded = true
    },
    logOut(state) {
      removeUserFromLocalStorage()
      state.expiresAt = undefined
      state.user = undefined
      state.runEffectCount += 1
      SessionApi.logOut()
    },
    runEffects(state) {
      state.runEffectCount += 1
    },
  },
})

export const { logIn, logOut, runEffects } = slice.actions

export const sessionInitialState = initialState

export default slice.reducer
