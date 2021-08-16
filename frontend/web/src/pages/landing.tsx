import React from 'react'
import Head from 'next/head'
import LandingPage from '@app/landing/pages/landing'

const Landing = () => {
  return (
    <>
      <Head>
        <title>Tadoku - About</title>
      </Head>
      <LandingPage />
    </>
  )
}

Landing.getInitialProps = async function () {
  return {
    overridesLayout: true,
  }
}

export default Landing
