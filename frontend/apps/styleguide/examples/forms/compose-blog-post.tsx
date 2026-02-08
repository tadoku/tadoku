import { Checkbox, Input, TextArea } from 'ui'
import { FormProvider, useForm } from 'react-hook-form'

export default function ComposeBlogPostForm() {
  const methods = useForm()
  const onSubmit = (data: unknown) => console.log(data, 'submitted')

  return (
    <FormProvider {...methods}>
      <form
        onSubmit={methods.handleSubmit(onSubmit)}
        className="v-stack spaced"
      >
        <Input
          name="title"
          label="Title"
          type="text"
          options={{
            required: true,
          }}
        />
        <TextArea
          name="content"
          label="Content"
          options={{
            required: true,
          }}
        />
        <Input
          name="publishedAt"
          label="Published at"
          type="date"
          options={{
            required: true,
            valueAsDate: true,
          }}
        />
        <Checkbox name="isPublished" label="Published" />
        <button
          type="submit"
          className="btn primary"
          disabled={methods.formState.isSubmitting}
        >
          Save
        </button>
      </form>
    </FormProvider>
  )
}
