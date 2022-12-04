import type { NextPage } from 'next'
import {
  getInitialPropsRedirectIfLoggedOut,
  NextPageContextWithSession,
} from '../src/session'

const Home: NextPage = () => {
  return null
}

Home.getInitialProps = async (ctx: NextPageContextWithSession) => {
  getInitialPropsRedirectIfLoggedOut(ctx)
  return {}
}

export default Home
