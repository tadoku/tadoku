import React from 'react'
import { initializeStore } from '../app/store'
import { Provider } from 'react-redux'
import App, { Container, NextAppContext } from 'next/app'
import withRedux from 'next-redux-wrapper'
import { Store } from 'redux'
import { loadUserFromLocalStorage } from '../app/session/storage'
import * as SessionStore from '../app/session/redux'
import SessionEffects from '../app/session/components/SessionEffects'
import RankingEffects from '../app/ranking/components/RankingEffects'

class MyApp extends App<{ store: Store }> {
  static async getInitialProps({ Component, ctx }: NextAppContext) {
    const pageProps = Component.getInitialProps
      ? await Component.getInitialProps(ctx)
      : {}

    return { pageProps }
  }

  componentDidMount() {
    const payload = loadUserFromLocalStorage()

    if (payload) {
      this.props.store.dispatch({
        type: SessionStore.ActionTypes.SessionSignIn,
        payload,
      })
    }
  }

  render() {
    const { Component, pageProps, store } = this.props

    return (
      <Container>
        <Provider store={store}>
          <SessionEffects />
          <RankingEffects />
          <Component {...pageProps} />
        </Provider>
      </Container>
    )
  }
}

export default withRedux(initializeStore)(MyApp)
