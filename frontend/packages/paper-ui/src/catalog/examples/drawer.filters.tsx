import { Button, Drawer, Select } from 'paper-ui'
import { useId, useState } from 'react'
import { FormProvider, useForm } from 'react-hook-form'

export default function DrawerExample() {
  const [open, setOpen] = useState(false)
  const [language, setLanguage] = useState('All languages')
  const formId = useId()
  const methods = useForm({ defaultValues: { language: 'All languages' } })
  return (
    <div className="paper-stack">
      <p role="status">Showing: {language}</p>
      <FormProvider {...methods}>
        <Drawer
          trigger={<Button variant="outline">Review filters</Button>}
          title="Entry filters"
          description="Choose which reading entries to show."
          open={open}
          onOpenChange={nextOpen => {
            if (nextOpen) methods.reset({ language })
            setOpen(nextOpen)
          }}
          footer={
            <>
              <Button variant="ghost" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <Button type="submit" form={formId}>
                Apply filters
              </Button>
            </>
          }
        >
          <form
            className="paper-stack"
            id={formId}
            onSubmit={methods.handleSubmit(values => {
              setLanguage(values.language)
              setOpen(false)
            })}
          >
            <Select
              name="language"
              label="Language"
              options={[
                { value: 'All languages', label: 'All languages' },
                { value: 'Japanese', label: 'Japanese' },
                { value: 'French', label: 'French' },
              ]}
              hint="Changes take effect only when you apply filters."
            />
          </form>
        </Drawer>
      </FormProvider>
      <h3>Start placement, without a footer</h3>
      <Drawer
        placement="start"
        trigger={<Button variant="outline">Reading help</Button>}
        title="Reading help"
        description="How entries contribute to your progress."
      >
        <p>
          Record the amount you read in the original unit. Your reading history
          stays available even when an entry is not part of a contest.
        </p>
        <p>
          Close this sheet with the close button or Escape to return to your
          entries.
        </p>
      </Drawer>
    </div>
  )
}
