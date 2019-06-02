import React from 'react'
import Modal from '../../ui/components/Modal'
import JoinContestForm from './JoinContestForm'

const JoinContestModal = ({
  isOpen,
  onSuccess,
  onCancel,
}: {
  isOpen: boolean
  onSuccess: () => void
  onCancel: () => void
}) => {
  return (
    <Modal
      isOpen={isOpen}
      onRequestClose={onCancel}
      contentLabel="Join contest"
    >
      <JoinContestForm onSuccess={onSuccess} onCancel={onCancel} />
    </Modal>
  )
}

export default JoinContestModal
