import React from 'react'
import LogInForm from './LogInForm'
import Modal from '../../ui/components/Modal'

const LogInModal = ({
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
      <LogInForm onSuccess={onSuccess} onCancel={onCancel} />
    </Modal>
  )
}

export default LogInModal
