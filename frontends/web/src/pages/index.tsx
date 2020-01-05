import { useEffect } from 'react'
import Router from 'next/router'
import { connect } from 'react-redux'
import { User } from '../app/session/interfaces'
import { State } from '../app/store'
import LandingPage from '../app/landing/pages/landing'
import Blog from './blog'
interface Props {
  user: User | undefined
}

const Home = ({ user }: Props) => {
  useEffect(() => {
    if (user) {
      Router.replace('/blog')
    }
  }, [user])

  if (user) {
    return <Blog />
  }

  return <LandingPage />
}

const mapStateToProps = (state: State) => ({
  user: state.session.user,
})

export default connect(mapStateToProps)(Home)
