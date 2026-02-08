import { useState } from 'react'
import { Modal } from 'ui'
import { toast } from 'react-toastify'

export default function StandardModalExample() {
  const [isOpen, setIsOpen] = useState(false)

  return (
    <>
      <div className="h-stack spaced">
        <button className="btn" onClick={() => setIsOpen(true)}>
          Open modal
        </button>
      </div>
      <Modal isOpen={isOpen} setIsOpen={setIsOpen} title="Are you sure?">
        <p className="modal-body">
          By forfeiting all your data for this contest will be deleted and
          cannot be undone.
        </p>

        <div className="modal-actions">
          <button
            type="button"
            className="btn danger"
            onClick={() => {
              toast.info('Your request is being processed')
              setIsOpen(false)
            }}
          >
            Yes, delete it
          </button>
          <button
            type="button"
            className="btn ghost"
            onClick={() => setIsOpen(false)}
          >
            Go back
          </button>
        </div>
      </Modal>
    </>
  )
}
