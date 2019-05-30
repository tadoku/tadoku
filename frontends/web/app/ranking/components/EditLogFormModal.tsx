import React from 'react'
import { ContestLog } from '../interfaces'
import Modal from 'react-modal'
import EditLogForm from './EditLogForm'

const EditLogFormModal = ({
  log,
  setLog,
  onSuccess,
}: {
  log: ContestLog | undefined
  setLog: (log: ContestLog | undefined) => void
  onSuccess: () => void
}) => {
  return (
    <Modal
      isOpen={!!log}
      onRequestClose={() => setLog(undefined)}
      contentLabel="Update"
    >
      {log && <EditLogForm log={log} onSuccess={onSuccess} />}
    </Modal>
  )
}

export default EditLogFormModal
