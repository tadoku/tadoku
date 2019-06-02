import React from 'react'
import SignInForm from './SignInForm'
import Modal from '../../ui/components/Modal'

const SignInModal = ({
  isOpen,
  onSuccess,
  onCancel,
}: {
  isOpen: boolean
  onSuccess: () => void
  onCancel: () => void
}) => {
  return (
    <Modal isOpen={isOpen} onRequestClose={onCancel} contentLabel="Sign in">
      <h2>Sign in</h2>
      <SignInForm onSuccess={onSuccess} onCancel={onCancel} />
    </Modal>
  )
}

export default SignInModal
