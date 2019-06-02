import React from 'react'
import { ContestLog } from '../interfaces'
import EditLogForm from './EditLogForm'
import { Modal } from '../../ui/components'

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
    >
      {log && (
        <EditLogForm log={log} onSuccess={onSuccess} onCancel={onCancel} />
      )}
    </Modal>
  )
}

export default EditLogFormModal
