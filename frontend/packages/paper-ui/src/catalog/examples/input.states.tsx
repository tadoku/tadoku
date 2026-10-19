import { usePaperForm } from 'paper-ui'

export default function InputStatesFixture() {
  const methods = usePaperForm({
    defaultValues: { readonly: 'Japanese', disabled: 'Archived' },
  })
  return (
    <methods.AppForm>
      <div className="paper-stack">
        <methods.AppField name="readonly">{field => <field.Input label="Language" readOnly />}</methods.AppField>
        <methods.AppField name="disabled">{field => <field.Input label="Status" disabled />}</methods.AppField>
      </div>
    </methods.AppForm>
  )
}
