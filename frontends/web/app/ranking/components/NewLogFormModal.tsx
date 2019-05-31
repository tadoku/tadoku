import React from 'react'
import Modal from 'react-modal'
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
      <EditLogForm onSuccess={onSuccess} onCancel={onCancel} />
    </Modal>
  )
}

export default NewLogFormModal
