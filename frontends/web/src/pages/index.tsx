import React from 'react'
import LandingPage from '../app/landing/pages/landing'
import { NextPageContext } from 'next'

const Home = () => {
  return <LandingPage />
}

Home.getInitialProps = async function (ctx: NextPageContext) {
  if (typeof window === 'undefined') {
    const state = ctx.store.getState()
    if (state.session.user) {
      ctx?.res?.writeHead(301, { Location: '/blog' })
      ctx?.res?.end()
      return {}
    }
  }

  return {
    overridesLayout: true,
  }
}

export default Home
