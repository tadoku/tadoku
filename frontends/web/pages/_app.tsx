import React from 'react'
import { initializeStore } from '../app/store'
import { Provider } from 'react-redux'
import App, { Container, NextAppContext } from 'next/app'
import withRedux from 'next-redux-wrapper'
import { Store } from 'redux'
import { loadUserFromLocalStorage } from '../app/session/storage'
import * as SessionStore from '../app/session/redux'
import AppEffects from '../app/AppEffects'
import 'react-vis/dist/style.css'
import Modal from 'react-modal'

// Setup icons
import { library } from '@fortawesome/fontawesome-svg-core'
import {
  faChevronDown,
  faEdit,
  faTrash,
} from '@fortawesome/free-solid-svg-icons'
library.add(faChevronDown, faEdit, faTrash)

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

    Modal.setAppElement('#__next')
  }

  render() {
    const { Component, pageProps, store } = this.props

    return (
      <Container>
        <Provider store={store}>
          <AppEffects />
          <Component {...pageProps} />
        </Provider>
      </Container>
    )
  }
}

export default withRedux(initializeStore)(MyApp)
