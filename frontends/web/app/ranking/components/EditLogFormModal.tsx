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
      style={{
        content: {
          top: '50%',
          left: '50%',
          right: 'auto',
          bottom: 'auto',
          marginRight: '-50%',
          transform: 'translate(-50%, -50%)',
          border: 0,
          boxShadow: '4px 15px 20px 1px rgba(0, 0, 0, 0.28)',
          padding: '40px',
        },
        overlay: {
          backgroundColor: 'rgba(0, 0, 0, 0.4)',
        },
      }}
    >
      {log && <EditLogForm log={log} onSuccess={onSuccess} />}
    </Modal>
  )
}

export default EditLogFormModal
