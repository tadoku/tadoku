import React from 'react'
import LandingPage from '@app/landing/pages/landing'
import { NextPage } from 'next'

const Home: NextPage = () => {
  return <LandingPage />
}

Home.getInitialProps = async function ({ store, res }) {
  const isServer = typeof window === 'undefined'

  if (isServer) {
    const state = store.getState()
    if (state.session.user) {
      res?.writeHead(301, { Location: '/blog' })
      res?.end()
      return {}
    }
  }

  return {
    overridesLayout: true,
  }
}

export default Home
