import React from 'react'
import Modal from '../../ui/components/Modal'
import LogForm from './LogForm'

const NewLogFormModal = ({
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
      contentLabel="Submit new pages"
    >
      <LogForm onSuccess={onSuccess} onCancel={onCancel} />
    </Modal>
  )
}

export default NewLogFormModal
