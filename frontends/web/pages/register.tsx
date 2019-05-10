import React from 'react'
import Layout from '../components/ui/Layout'
import { withRedirectAuthenticated } from '../components/hoc/withRedirectAuthenticated'

const Register = () => {
  return (
    <Layout>
      <h2>Create a new account</h2>
    </Layout>
  )
}

export default withRedirectAuthenticated(Register)
