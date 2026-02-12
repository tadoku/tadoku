import { Separator } from '@components/example'
import { Showcase } from '@components/Showcase'

import LogPagesForm from '@examples/forms/log-pages'
import logPagesCode from '@examples/forms/log-pages.tsx?raw'

import AutocompleteForm from '@examples/forms/autocomplete'
import autocompleteCode from '@examples/forms/autocomplete.tsx?raw'

import TagsInputForm from '@examples/forms/tags-input'
import tagsInputCode from '@examples/forms/tags-input.tsx?raw'

import ComposeBlogPostForm from '@examples/forms/compose-blog-post'
import composeBlogPostCode from '@examples/forms/compose-blog-post.tsx?raw'

import MiscForm from '@examples/forms/misc-elements'
import miscElementsCode from '@examples/forms/misc-elements.tsx?raw'

import BasicElements from '@examples/forms/basic-elements'
import basicElementsCode from '@examples/forms/basic-elements.tsx?raw'

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
