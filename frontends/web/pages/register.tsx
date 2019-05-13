import React from 'react'
import Layout from '../app/ui/components/Layout'
import RegisterForm from '../components/forms/RegisterForm'
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
