import React from 'react'
import Layout from '../components/ui/Layout'
import SignInForm from '../components/forms/SignIn'
import { withRedirectAuthenticated } from '../components/hoc/withRedirectAuthenticated'

const SignIn = () => {
  return (
    <Layout>
      <h2>Sign in</h2>
      <SignInForm />
    </Layout>
  )
}

export default withRedirectAuthenticated(SignIn)
