import React from 'react'
import Head from 'next/head'
import ErrorPage from 'next/error'

import { contestCollectionSerializer } from '@app/contest/transform'
import { useCachedApiState, isReady } from '../app/cache'
import ContestApi from '@app/contest/api'
import { Contest } from '@app/contest/interfaces'
import { PageTitle } from '@app/ui/components'
import ContestList from '@app/contest/components/ContestList'

const Contests = () => {
  const { data: contests, status: statusContests } = useCachedApiState<
    Contest[]
  >({
    cacheKey: `contests?i=1`,
    defaultValue: [],
    fetchData: () => {
      return ContestApi.getAll()
    },
    dependencies: [],
    serializer: contestCollectionSerializer,
  })

  if (!isReady([statusContests])) {
    return <p>Loading...</p>
  }

  if (!contests) {
    return <ErrorPage statusCode={404} />
  }

  return (
    <>
      <Head>
        <title>Tadoku - Contest Archive</title>
      </Head>
      <PageTitle>Contest archive</PageTitle>
      <ContestList contests={contests} />
    </>
  )
}

export default Contests
