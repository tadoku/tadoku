import { Button, Input, Modal } from 'paper-ui'
import { useRef, useState } from 'react'
import { FormProvider, useForm } from 'react-hook-form'

export default function ComposableModalFixture() {
  const [open, setOpen] = useState(false)
  const searchRef = useRef<HTMLInputElement>(null)
  const methods = useForm({ defaultValues: { query: '' } })
  const query = methods.watch('query')
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
      <FormProvider {...methods}>
        <Input
          name="query"
          label="Component name"
          type="search"
          ref={searchRef}
        />
        <p role="status">
          {matches.length ? matches.join(', ') : 'No matching components.'}
        </p>
      </FormProvider>
    </Modal>
  )
}
