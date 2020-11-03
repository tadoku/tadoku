import React from 'react'
import { wrapper } from '../app/store'
import App from 'next/app'
import { Store } from 'redux'
import AppEffects from '../app/AppEffects'
import 'react-vis/dist/style.css'
import Modal from 'react-modal'
import '@app/ui/setup'
import Layout from '@app/ui/components/Layout'
import { parseSessionFromContext } from '@app/session/domain'

class MyApp extends App<{ store: Store }> {
  componentDidMount() {
    Modal.setAppElement('#__next')
  }

  render() {
    const { Component, pageProps } = this.props

    return (
      <>
        <AppEffects />
        <Layout overridesLayout={pageProps?.overridesLayout ?? false}>
          <Component {...pageProps} />
        </Layout>
      </>
    )
  }
}

MyApp.getInitialProps = async ({ Component, ctx }) => {
  parseSessionFromContext(ctx)

  const pageProps = Component.getInitialProps
    ? await Component.getInitialProps(ctx)
    : {}

  return { pageProps }
}

export default wrapper.withRedux(MyApp)
