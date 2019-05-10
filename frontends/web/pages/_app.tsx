import React from 'react'
import { initializeStore, SessionActionTypes } from './../store'
import { Provider } from 'react-redux'
import App, { Container, NextAppContext } from 'next/app'
import withRedux from 'next-redux-wrapper'
import { Store } from 'redux'
import { loadUser } from '../domain/Session'

class MyApp extends App<{ store: Store }> {
  static async getInitialProps({ Component, ctx }: NextAppContext) {
    const pageProps = Component.getInitialProps
      ? await Component.getInitialProps(ctx)
      : {}

    return { pageProps }
  }

  componentDidMount() {
    const payload = loadUser()

    if (payload) {
      this.props.store.dispatch({
        type: SessionActionTypes.SessionSignIn,
        payload,
      })
    }
  }

  render() {
    const { Component, pageProps, store } = this.props

    return (
      <Container>
        <Provider store={store}>
          <Component {...pageProps} />
        </Provider>
      </Container>
    )
  }
}

export default withRedux(initializeStore)(MyApp)
