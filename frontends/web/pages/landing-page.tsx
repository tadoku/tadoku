import React from 'react'
import Layout from '../app/ui/components/Layout'

const LandingPage = () => {
  return (
    <Layout title="Welcome to Tadoku!">
      <p>
        We'll be running a test round from June 15th until June 30th UTC. All
        existing data will be wiped after this. Registrations are open now.
        Please sign up to participate.
      </p>
    </Layout>
  )
}

export default LandingPage
