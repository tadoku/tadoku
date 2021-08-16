import React from 'react'
import { ContestLog } from '../../interfaces'
import LogForm from '../forms/LogForm'
import Modal from '@app/ui/components/Modal'

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
      {log && <LogForm log={log} onSuccess={onSuccess} onCancel={onCancel} />}
    </Modal>
  )
}

export default EditLogFormModal
