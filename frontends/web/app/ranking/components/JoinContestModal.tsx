import React from 'react'
import Modal from '../../ui/components/Modal'

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
      TODO
    </Modal>
  )
}

export default JoinContestModal
