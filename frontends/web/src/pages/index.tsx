import { useEffect } from 'react'
import Router from 'next/router'
import { connect } from 'react-redux'
import { User } from '../app/session/interfaces'
import { State } from '../app/store'
interface Props {
  user: User | undefined
}

const Home = ({}: Props) => {
  useEffect(() => {
    Router.replace('/blog')
  }, [])

  return null
}

const mapStateToProps = (state: State) => ({
  user: state.session.user,
})

export default connect(mapStateToProps)(Home)
