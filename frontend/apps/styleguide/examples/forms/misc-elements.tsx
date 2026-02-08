import { Input, RadioSelect } from 'ui'
import { FormProvider, useForm } from 'react-hook-form'

export default function MiscForm() {
  const methods = useForm()
  const onSubmit = (data: unknown) => console.log(data, 'submitted')

  return (
    <FormProvider {...methods}>
      <form
        onSubmit={methods.handleSubmit(onSubmit)}
        className="v-stack spaced"
      >
        <RadioSelect
          name="cardType"
          label="Card type"
          options={{
            required: true,
            valueAsNumber: true,
          }}
          values={[
            { value: '1', label: 'Sentence card' },
            { value: '2', label: 'Vocab card' },
          ]}
        />
        <Input
          name="datetime"
          label="Datetime"
          type="datetime-local"
          options={{
            required: true,
            valueAsDate: true,
          }}
        />
        <Input
          name="time"
          label="Time"
          type="time"
          options={{
            required: true,
          }}
        />
        <Input
          name="week"
          label="Week"
          type="week"
          options={{
            required: true,
          }}
        />
        <Input
          name="month"
          label="Month"
          type="month"
          options={{
            required: true,
          }}
        />
        <Input
          name="color"
          label="Color"
          type="color"
          options={{
            required: true,
          }}
        />
        <Input
          name="email"
          label="Email"
          type="email"
          options={{
            required: true,
          }}
        />
        <Input
          name="file"
          label="File"
          type="file"
          options={{
            required: true,
          }}
        />
        <Input
          name="range"
          label="Range"
          type="range"
          options={{
            required: true,
          }}
        />
        <Input
          name="search"
          label="Search"
          type="search"
          options={{
            required: true,
          }}
        />
        <Input
          name="tel"
          label="tel"
          type="tel"
          options={{
            required: true,
          }}
        />
        <Input
          name="url"
          label="url"
          type="url"
          options={{
            required: true,
          }}
        />
        <button
          type="submit"
          className="btn primary"
          disabled={methods.formState.isSubmitting}
        >
          Submit
        </button>
      </form>
    </FormProvider>
  )
}
