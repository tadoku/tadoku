import { useEffect } from 'react'
import Router from 'next/router'
import { connect } from 'react-redux'
import { User } from '../app/session/interfaces'
import { RootState } from '../app/store'
interface Props {
  user: User | undefined
}

const Home = ({}: Props) => {
  useEffect(() => {
    Router.replace('/blog')
  }, [])

  return null
}

const mapStateToProps = (state: RootState) => ({
  user: state.session.user,
})

export default connect(mapStateToProps)(Home)
