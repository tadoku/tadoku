import React, { useState } from 'react'
import { ContestLog, RankingRegistrationOverview } from '../interfaces'
import EditLogFormModal from './modals/EditLogFormModal'
import { State } from '../../store'
import { connect } from 'react-redux'
import { User } from '../../session/interfaces'
import RankingApi from '../api'
import ContestLogsTable from './ContestLogsTable'
import ContestLogsList from './ContestLogsList'

interface Props {
  logs: ContestLog[]
  registration: RankingRegistrationOverview
  signedInUser?: User | undefined
  refreshData: () => void
}

const ContestLogsOverview = (props: Props) => {
  const [selectedLog, setSelectedLog] = useState(undefined as
    | ContestLog
    | undefined)

  const finishUpdate = () => {
    props.refreshData()
    setSelectedLog(undefined)
  }

  const deleteLog = (log: ContestLog) => {
    const shouldDelete = confirm('Are you sure you want to delete this?')

    if (!shouldDelete) {
      return
    }

    RankingApi.deleteLog(log.id)
    props.refreshData()
  }

  const canEdit =
    (props.signedInUser &&
      props.signedInUser.id === props.registration.userId) ||
    false

  return (
    <>
      <h1>Updates</h1>
      <EditLogFormModal
        log={selectedLog}
        setLog={setSelectedLog}
        onSuccess={finishUpdate}
        onCancel={() => setSelectedLog(undefined)}
      />
      <ContestLogsTable
        logs={props.logs}
        canEdit={canEdit}
        editLog={setSelectedLog}
        deleteLog={deleteLog}
      />
      <ContestLogsList
        logs={props.logs}
        canEdit={canEdit}
        editLog={setSelectedLog}
        deleteLog={deleteLog}
      />
    </>
  )
}

const mapStateToProps = (state: State, props: Props) => ({
  ...props,
  signedInUser: state.session.user,
})

export default connect(mapStateToProps)(ContestLogsOverview)
