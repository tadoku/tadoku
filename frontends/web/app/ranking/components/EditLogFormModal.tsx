import React from 'react'
import { ContestLog } from '../interfaces'
import Modal from 'react-modal'
import EditLogForm from './EditLogForm'
import { modalStyles } from '../../ui/components'

const EditLogFormModal = ({
  log,
  setLog,
  onSuccess,
  onCancel,
}: {
  log: ContestLog | undefined
  setLog: (log: ContestLog | undefined) => void
  onSuccess: () => void
  onCancel: () => void
}) => {
  return (
    <Modal
      isOpen={!!log}
      onRequestClose={() => setLog(undefined)}
      contentLabel="Update"
      style={modalStyles}
    >
      {log && (
        <EditLogForm log={log} onSuccess={onSuccess} onCancel={onCancel} />
      )}
    </Modal>
  )
}

export default EditLogFormModal
