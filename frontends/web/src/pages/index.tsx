import React from 'react'
import LandingPage from '@app/landing/pages/landing'
import { AppContext } from 'next/app'

const Home = () => {
  return <LandingPage />
}

Home.getInitialProps = async function ({ ctx }: AppContext) {
  const isServer = typeof window === 'undefined'
  if (isServer) {
    const state = ctx.store.getState()
    if (state.session.user) {
      ctx.res?.writeHead(301, { Location: '/blog' })
      ctx.res?.end()
      return {}
    }
  }

  return {
    overridesLayout: true,
  }
}

export default Home
