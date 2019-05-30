import React from 'react'
import { ContestLog } from '../interfaces'
import Modal from 'react-modal'
import EditLogForm from './EditLogForm'

const EditLogFormModal = ({
  log,
  setLog,
}: {
  log: ContestLog | undefined
  setLog: (log: ContestLog | undefined) => void
}) => {
  return (
    <Modal
      isOpen={!!log}
      onRequestClose={() => setLog(undefined)}
      contentLabel="Update"
    >
      {log && <EditLogForm log={log} />}
    </Modal>
  )
}

export default EditLogFormModal
