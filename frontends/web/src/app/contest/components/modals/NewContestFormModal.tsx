import React from 'react'
import ContestForm from '../forms/ContestForm'
import Modal from '@app/ui/components/Modal'

const NewContestFormModal = ({
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
      contentLabel="Create a new contest"
    >
      <ContestForm onSuccess={onSuccess} onCancel={onCancel} />
    </Modal>
  )
}

export default NewContestFormModal
