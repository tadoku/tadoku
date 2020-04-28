import React, { useEffect } from 'react'
import Router from 'next/router'
import { useSelector } from 'react-redux'
import { RootState } from '../app/store'
import LandingPage from '../app/landing/pages/landing'
import Blog from './blog'

const Home = () => {
  const user = useSelector((state: RootState) => state.session.user)

  useEffect(() => {
    if (user) {
      Router.replace('/blog')
    } else {
      Router.replace('/landing')
    }
  }, [user])

  if (user) {
    return <Blog />
  }

  return <LandingPage />
}

Home.getInitialProps = async function () {
  return {
    overridesLayout: true,
  }
}

export default Home
