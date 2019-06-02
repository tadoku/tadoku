import React from 'react'
import Modal from '../../ui/components/Modal'
import EditLogForm from './EditLogForm'

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
      <EditLogForm onSuccess={onSuccess} onCancel={onCancel} />
    </Modal>
  )
}

export default NewLogFormModal
