import { useEffect } from 'react'
import Router from 'next/router'
import LandingPage from './landing-page'
import { connect } from 'react-redux'
import { User } from '../app/user/User'
import { State } from '../store'
interface Props {
  user: User | undefined
}

const Home = ({ user }: Props) => {
  useEffect(() => {
    if (user) {
      Router.replace('/ranking')
    }
  }, [user])

  return <LandingPage />
}

const mapStateToProps = (state: State) => ({
  user: state.user,
})

export default connect(mapStateToProps)(Home)
