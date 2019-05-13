import React from 'react'
import Layout from '../app/ui/Layout'
import RegisterForm from '../components/forms/Register'
import { withRedirectAuthenticated } from '../app/session/components/withRedirectAuthenticated'

const Register = () => {
  return (
    <Layout>
      <h2>Create a new account</h2>
      <RegisterForm />
    </Layout>
  )
}

export default withRedirectAuthenticated(Register)
