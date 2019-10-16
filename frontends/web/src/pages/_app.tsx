import React from 'react'
import { initializeStore } from '../store'
import { Provider } from 'react-redux'
import App, { AppContext } from 'next/app'
import withRedux from 'next-redux-wrapper'
import { Store } from 'redux'
import { loadUserFromLocalStorage } from '../session/storage'
import * as SessionStore from '../session/redux'
import AppEffects from '../AppEffects'
import 'react-vis/dist/style.css'
import Modal from 'react-modal'
import '../ui/setup'
import Layout from '../ui/components/Layout'

class MyApp extends App<{ store: Store }> {
  static async getInitialProps({ Component, ctx }: AppContext) {
    const pageProps = Component.getInitialProps
      ? await Component.getInitialProps(ctx)
      : {}

    return { pageProps }
  }

  componentDidMount() {
    const payload = loadUserFromLocalStorage()

    if (payload) {
      this.props.store.dispatch({
        type: SessionStore.ActionTypes.SessionLogIn,
        payload,
      })
    }

    Modal.setAppElement('#__next')
  }

  render() {
    const { Component, pageProps, store } = this.props

    return (
      <Provider store={store}>
        <AppEffects />
        <Layout>
          <Component {...pageProps} />
        </Layout>
      </Provider>
    )
  }
}

export default withRedux(initializeStore)(MyApp)
