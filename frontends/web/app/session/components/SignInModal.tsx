import React from 'react'
import SignInForm from './SignInForm'
import Modal from '../../ui/components/Modal'

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
      <h2>Sign in</h2>
      <SignInForm />
    </Modal>
  )
}

export default NewLogFormModal
