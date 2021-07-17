import React from 'react'
import { Contest } from '../../interfaces'
import ContestForm from '../forms/ContestForm'
import Modal from '@app/ui/components/Modal'

const EditContestFormModal = ({
  contest,
  setContest,
  onSuccess,
  onCancel,
}: {
  contest: Contest | undefined
  setContest: (log: Contest | undefined) => void
  onSuccess: () => void
  onCancel: () => void
}) => {
  return (
    <Modal
      isOpen={!!contest}
      onRequestClose={() => setContest(undefined)}
      contentLabel={`Edit contest ${contest?.description ?? ''}`}
    >
      {contest && (
        <ContestForm
          contest={contest}
          onSuccess={onSuccess}
          onCancel={onCancel}
        />
      )}
    </Modal>
  )
}

export default EditContestFormModal
