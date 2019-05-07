import React from 'react'
import Layout from '../components/ui/Layout'
import { connect } from 'react-redux'
import { State, AppActionTypes, AppActions } from '../store'
import { Dispatch, bindActionCreators } from 'redux'

const Home = ({ open, go }: { open: boolean; go: () => void }) => {
  return (
    <Layout>
      Welcome to Tadoku! We are {open ? 'open' : 'closed'}!
      <button onClick={go}>Go now!</button>
    </Layout>
  )
}

const mapStateToProps = (state: State) => ({
  open: state.open,
})

const mapDispatchToProps = (dispatch: Dispatch<AppActions>) => ({
  go: () => dispatch({ type: AppActionTypes.AppOpen }),
})

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(Home)
