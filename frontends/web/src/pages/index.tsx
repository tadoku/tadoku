import React from 'react'
import LandingPage from '@app/landing/pages/landing'
import { AppContext } from 'next/app'

const Home = () => {
  return <LandingPage />
}

Home.getInitialProps = async function ({ ctx }: AppContext) {
  const isServer = typeof window === 'undefined'
  const hasContext = ctx !== undefined
  if (isServer && hasContext) {
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
