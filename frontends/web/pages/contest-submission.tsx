import React from 'react'
import Layout from '../app/ui/components/Layout'
import UpdateForm from '../app/ranking/components/UpdateForm'

const ContestSubmission = () => {
  return (
    <Layout>
      <h2>Update</h2>
      <UpdateForm />
    </Layout>
  )
}

// TODO: redirect on unauthenticated
export default ContestSubmission
