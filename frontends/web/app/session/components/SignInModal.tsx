import React from 'react'
import Modal from 'react-modal'
import SignInForm from './SignInForm'
import { modalStyles } from '../../ui/components'

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
      style={modalStyles}
    >
      <h2>Sign in</h2>
      <SignInForm />
    </Modal>
  )
}

export default NewLogFormModal
