import { useEffect } from 'react'
import Router from 'next/router'
import { connect } from 'react-redux'
import { User } from '../src/session/interfaces'
import { State } from '../src/store'
interface Props {
  user: User | undefined
}

const Home = ({  }: Props) => {
  useEffect(() => {
    Router.replace('/ranking')
  }, [])

  return null
}

const mapStateToProps = (state: State) => ({
  user: state.session.user,
})

export default connect(mapStateToProps)(Home)
