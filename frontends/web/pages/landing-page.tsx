import React from 'react'
import { PageTitle } from '../app/ui/components'

const LandingPage = () => {
  return (
    <>
      <PageTitle>Welcome to Tadoku!</PageTitle>
      <p>
        We'll be running a test round from June 15th until June 30th UTC. All
        existing data will be wiped after this. Registrations are open now.
        Please sign up to participate.
      </p>
    </>
  )
}

export default LandingPage
