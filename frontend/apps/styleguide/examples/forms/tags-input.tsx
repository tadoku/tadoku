import { TagsInput } from 'ui'
import { FormProvider, useForm } from 'react-hook-form'

export default function TagsInputForm() {
  const methods = useForm()
  const onSubmit = (data: unknown) => console.log(data, 'submitted')

  const allTags = [
    'Book',
    'Ebook',
    'Fiction',
    'Non-fiction',
    'Web page',
    'Lyric',
    'Manga',
    'Novel',
  ]

  // Simulated async API call with 150ms delay
  const fetchTags = async (input: string): Promise<string[]> => {
    await new Promise(resolve => setTimeout(resolve, 150))
    return allTags.filter(tag =>
      tag.toLowerCase().includes(input.toLowerCase())
    )
  }

  return (
    <FormProvider {...methods}>
      <form
        onSubmit={methods.handleSubmit(onSubmit)}
        className="v-stack spaced"
      >
        <TagsInput
          name="tags"
          label="Tags"
          hint="Add up to 5 tags"
          placeholder="Type to search or add tags..."
          maxTags={5}
          getSuggestions={fetchTags}
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
