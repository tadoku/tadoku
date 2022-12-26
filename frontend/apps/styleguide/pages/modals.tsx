import { useState } from 'react'
import { CodeBlock, Preview, Title } from '@components/example'
import { Modal } from 'ui'
import { toast } from 'react-toastify'

export default function Modals() {
  return (
    <>
      <h1 className="title mb-8">Modal</h1>
      <Title>Standard</Title>
      <p>
        In general modals should be used for <strong>short interactions</strong>{' '}
        such as:
      </p>
      <ul className="list">
        <li>Confirmation dialogs</li>
        <li>Single input forms</li>
      </ul>
      <p>
        <strong>Avoid</strong> in cases such as:
      </p>
      <ul className="list !mb-6">
        <li>Complex forms</li>
      </ul>

      <Preview>
        <StandardModalExample />
      </Preview>
      <CodeBlock
        language="typescript"
        code={`import Modal from '@components/Modal'
import { toast } from 'react-toastify'
import { useState } from 'react'

const StandardModalExample = () => {
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
}`}
      />
    </>
  )
}

const StandardModalExample = () => {
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
