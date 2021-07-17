import { configureStore, combineReducers, Store } from '@reduxjs/toolkit'
import appReducer, { appInitialState } from './redux'
import contestReducer, { contestInitialState } from '@app/contest/redux'
import rankingReducer, { rankingInitialState } from '@app/ranking/redux'
import sessionReducer, { sessionInitialState } from '@app/session/redux'
import { createWrapper, MakeStore } from 'next-redux-wrapper'

const initialState = {
  app: appInitialState,
  contest: contestInitialState,
  ranking: rankingInitialState,
  session: sessionInitialState,
}

export const reducer = combineReducers({
  app: appReducer,
  contest: contestReducer,
  ranking: rankingReducer,
  session: sessionReducer,
})

const makeStore: MakeStore<Store<RootState>> = () => {
  const store: Store = configureStore({
    reducer: reducer,
  })
  return store
}

export const wrapper = createWrapper<Store<RootState>>(makeStore, {
  debug: process.env.NODE_ENV === 'development',
})

export type RootState = typeof initialState
export type AppStore = ReturnType<typeof makeStore>
