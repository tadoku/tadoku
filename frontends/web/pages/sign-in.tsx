import React from 'react'
import Layout from '../app/ui/Layout'
import SignInForm from '../components/forms/SignIn'
import { withRedirectAuthenticated } from '../app/session/components/withRedirectAuthenticated'

const SignIn = () => {
  return (
    <Layout>
      <h2>Sign in</h2>
      <SignInForm />
    </Layout>
  )
}

export default withRedirectAuthenticated(SignIn)
