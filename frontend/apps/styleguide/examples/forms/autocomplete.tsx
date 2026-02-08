import { AutocompleteInput, AutocompleteMultiInput } from 'ui'
import { FormProvider, useForm } from 'react-hook-form'

export default function AutocompleteForm() {
  const methods = useForm()
  const onSubmit = (data: unknown) => console.log(data, 'submitted')

  const tags = ['Book', 'Ebook', 'Fiction', 'Non-fiction', 'Web page', 'Lyric']
  const activities = [
    { id: 1, name: 'Reading' },
    { id: 2, name: 'Listening' },
    { id: 3, name: 'Speaking' },
    { id: 4, name: 'Writing' },
  ]

  return (
    <FormProvider {...methods}>
      <form
        onSubmit={methods.handleSubmit(onSubmit)}
        className="v-stack spaced"
      >
        <AutocompleteInput
          name="tags"
          label="Tags"
          options={tags}
          match={(option, query) =>
            option
              .toLowerCase()
              .replace(/[^a-zA-Z0-9]/g, '')
              .includes(query.toLowerCase())
          }
          format={option => option}
        />
        <AutocompleteMultiInput
          name="activities"
          label="Activities"
          options={activities}
          match={(option, query) =>
            option.name
              .toLowerCase()
              .replace(/[^a-zA-Z0-9]/g, '')
              .includes(query.toLowerCase())
          }
          getIdForOption={option => option.id}
          format={option => option.name}
          hint="Select 3 at most"
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
