import { Separator } from '@components/example'
import { Showcase } from '@components/Showcase'

import LogPagesForm from '@examples/forms/log-pages'
import logPagesCode from '@examples/forms/log-pages.tsx?raw'

import AutocompleteForm from '@examples/forms/autocomplete'
import autocompleteCode from '@examples/forms/autocomplete.tsx?raw'

import ComposeBlogPostForm from '@examples/forms/compose-blog-post'
import composeBlogPostCode from '@examples/forms/compose-blog-post.tsx?raw'

import MiscForm from '@examples/forms/misc-elements'
import miscElementsCode from '@examples/forms/misc-elements.tsx?raw'

import BasicElements from '@examples/forms/basic-elements'
import basicElementsCode from '@examples/forms/basic-elements.tsx?raw'

import { TagsInput } from 'ui'
import { FormProvider, useForm } from 'react-hook-form'

export default function Forms() {
  return (
    <>
      <h1 className="title mb-8">Forms</h1>

      <Showcase title="React example: Log pages form" code={logPagesCode}>
        <div className="w-full max-w-xl">
          <LogPagesForm />
        </div>
      </Showcase>

      <Separator />

      <Showcase title="React example: Autocomplete" code={autocompleteCode}>
        <div className="w-96">
          <AutocompleteForm />
        </div>
      </Showcase>

      <Separator />

      <Showcase title="React example: Tags Input" code={tagsInputCode}>
        <div className="w-96">
          <TagsInputForm />
        </div>
      </Showcase>

      <Separator />

      <Showcase
        title="React example: Compose blog post form"
        code={composeBlogPostCode}
      >
        <div className="w-96">
          <ComposeBlogPostForm />
        </div>
      </Showcase>

      <Separator />

      <Showcase title="React example: other elements" code={miscElementsCode}>
        <div className="w-96">
          <MiscForm />
        </div>
      </Showcase>

      <Separator />

      <Showcase title="Basic elements" code={basicElementsCode} language="html">
        <div className="w-96">
          <BasicElements />
        </div>
      </Showcase>
    </>
  )
}

const tagsInputCode = `import { TagsInput } from 'ui'
import { FormProvider, useForm } from 'react-hook-form'

const TagsInputForm = () => {
  const methods = useForm()
  const onSubmit = (data: any) => console.log(data, 'submitted')

  // Sync example - filter a local array
  const allTags = ['Book', 'Ebook', 'Fiction', 'Non-fiction', 'Web page', 'Lyric', 'Manga', 'Novel']

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
          getSuggestions={(input) =>
            allTags.filter(tag =>
              tag.toLowerCase().includes(input.toLowerCase())
            )
          }
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

// Async example - fetch from API
const AsyncTagsInputForm = () => {
  const methods = useForm()

  return (
    <FormProvider {...methods}>
      <form className="v-stack spaced">
        <TagsInput
          name="tags"
          label="Tags"
          placeholder="Search tags..."
          debounceMs={300}
          getSuggestions={async (input) => {
            const res = await fetch(\`/api/tags/search?q=\${input}\`)
            return res.json()
          }}
        />
      </form>
    </FormProvider>
  )
}`

const TagsInputForm = () => {
  const methods = useForm()
  const onSubmit = (data: any) => console.log(data, 'submitted')

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

  // Simulated async API call with 500ms delay
  const fetchTags = async (input: string): Promise<string[]> => {
    await new Promise(resolve => setTimeout(resolve, 500))
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
          label="Tags (async)"
          hint="Add up to 5 tags"
          placeholder="Type to search or add tags..."
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
