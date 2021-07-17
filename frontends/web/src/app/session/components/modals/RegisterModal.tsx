import React from 'react'
import RegisterForm from '../forms/RegisterForm'
import Modal from '@app/ui/components/Modal'

const RegisterModal = ({
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
      contentLabel="Create a new account"
    >
      <RegisterForm onSuccess={onSuccess} onCancel={onCancel} />
    </Modal>
  )
}

export default RegisterModal
