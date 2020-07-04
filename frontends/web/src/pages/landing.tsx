import React from 'react'
import LandingPage from '@app/landing/pages/landing'

const Landing = () => {
  return <LandingPage />
}

Landing.getInitialProps = async function () {
  return {
    overridesLayout: true,
  }
}

export default Landing
