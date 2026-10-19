import { useStore } from '@tanstack/react-form'
import { Button, Modal, usePaperForm } from 'paper-ui'
import { useRef, useState } from 'react'

export default function ComposableModalFixture() {
  const [open, setOpen] = useState(false)
  const searchRef = useRef<HTMLInputElement>(null)
  const methods = usePaperForm({ defaultValues: { query: '' } })
  const query = useStore(methods.store, state => state.values['query'])
  const matches = ['Button', 'Input', 'Modal'].filter(name =>
    name.toLowerCase().includes(query.toLowerCase()),
  )
  return (
    <Modal
      trigger={<Button variant="ghost">Search examples</Button>}
      title="Search examples"
      description="Try Button, Input, or Modal."
      open={open}
      onOpenChange={setOpen}
      initialFocus={searchRef}
      footer={null}
    >
      <methods.AppForm>
        <methods.AppField name="query">
          {field => (
            <field.Input label="Component name" type="search" ref={searchRef} />
          )}
        </methods.AppField>
        <p role="status">
          {matches.length ? matches.join(', ') : 'No matching components.'}
        </p>
      </methods.AppForm>
    </Modal>
  )
}
