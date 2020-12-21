import React from 'react'
import Head from 'next/head'
import ErrorPage from 'next/error'
import { useSelector } from 'react-redux'

import { RootState } from '@app/store'
import { isAdmin } from '@app/session/domain'

const Manage = () => {
  const user = useSelector((state: RootState) => state.session.user)
  const isAllowed = isAdmin(user)

  if (!isAllowed) {
    return <ErrorPage statusCode={401} />
  }

  return (
    <>
      <Head>
        <title>Tadoku - Admin</title>
      </Head>
    </>
  )
}

export default Manage
