import type { NextPage } from 'next'
import Router from 'next/router'
import { NextPageContextWithSession } from './_app'

const Home: NextPage = () => {
  return null
}

Home.getInitialProps = async ({ res, session }: NextPageContextWithSession) => {
  if (!session) {
    if (res) {
      res.writeHead(307, { Location: '/login' })
      res.end()
    } else {
      Router.replace('/login')
    }
  }

  return Promise.resolve({})
}

export default Home
