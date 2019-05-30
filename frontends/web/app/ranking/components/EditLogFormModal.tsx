import React from 'react'
import { ContestLog } from '../interfaces'
import Modal from 'react-modal'

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
    />
  )
}

export default EditLogFormModal
