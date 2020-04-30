import React from 'react'
import { initializeStore } from '../app/store'
import { Provider } from 'react-redux'
import App, { AppContext } from 'next/app'
import withRedux from 'next-redux-wrapper'
import { Store } from 'redux'
import AppEffects from '../app/AppEffects'
import 'react-vis/dist/style.css'
import Modal from 'react-modal'
import '../app/ui/setup'
import Layout from '../app/ui/components/Layout'
import cookie from 'cookie'
import jwt from 'jsonwebtoken'
import { logIn } from '../app/session/redux'
import { userMapper } from '../app/session/transform'

class MyApp extends App<{ store: Store }> {
  static async getInitialProps({ Component, ctx }: AppContext) {
    const request = ctx.req
    if (request) {
      const cookies = cookie.parse(request.headers.cookie || '')
      const sessionCookie =
        cookies[process.env.SESSION_COOKIE_NAME || 'session_token']
      const decoded: any = jwt.decode(sessionCookie)
      if (decoded && decoded.User && decoded.exp) {
        ctx.store.dispatch(
          logIn({
            expiresAt: decoded.exp,
            user: userMapper.fromRaw(decoded.User),
          }),
        )
      }
    }

    const pageProps = Component.getInitialProps
      ? await Component.getInitialProps(ctx)
      : {}

    return { pageProps }
  }

  componentDidMount() {
    Modal.setAppElement('#__next')
  }

  render() {
    const { Component, pageProps, store } = this.props

    return (
      <Provider store={store}>
        <AppEffects />
        <Layout overridesLayout={pageProps?.overridesLayout ?? false}>
          <Component {...pageProps} />
        </Layout>
      </Provider>
    )
  }
}

export default withRedux(initializeStore)(MyApp)
