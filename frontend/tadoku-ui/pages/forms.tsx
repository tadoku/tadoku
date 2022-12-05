import { CodeBlock, Preview, Separator, Title } from '@components/example'
import { Input, Select } from '@components/Form'
import { useForm } from 'react-hook-form'

export default function Forms() {
  return (
    <>
      <h1 className="title mb-8">Forms</h1>

      <Title>React example: Log pages form</Title>
      <div className="h-stack w-full">
        <div className="w-96">
          <Preview>
            <LogPagesForm />
          </Preview>
        </div>
        <div className="flex-1">
          <CodeBlock
            language="typescript"
            code={`import { Input, Select } from '@components/Form'
import { useForm } from 'react-hook-form'

const LogPagesForm = () => {
  const { register, handleSubmit, formState } = useForm()
  const onSubmit = (data: any) => console.log(data, 'submitted')

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="v-stack">
      <Input
        name="pagesRead"
        label="Pages read"
        register={register}
        formState={formState}
        type="number"
        options={{
          required: 'This field is required',
          valueAsNumber: true,
        }}
      />
      <Select
        name="medium"
        label="Medium"
        register={register}
        formState={formState}
        options={{
          required: true,
          valueAsNumber: true,
        }}
        values={[
          { value: '1', label: 'Book' },
          { value: '2', label: 'Comic' },
          { value: '3', label: 'Sentence' },
        ]}
      />
      <Input
        name="description"
        label="Description"
        register={register}
        formState={formState}
        type="text"
        placeholder="e.g. One Piece volume 45"
      />
      <button
        type="submit"
        className="btn primary"
        disabled={formState.isSubmitting}
      >
        Save changes
      </button>
    </form>
  )
}`}
          />
        </div>
      </div>

      <Separator />

      <Title>React example: Compose blog post form</Title>
      <div className="h-stack w-full">
        <div className="w-96">
          <Preview>todo</Preview>
        </div>
        <div className="flex-1">
          <CodeBlock
            language="typescript"
            code={`// TODO: Form for creating a blog post
// imports
// Title: text
// Content: textarea
// Published: checkmark
// Published at: datetime`}
          />
        </div>
      </div>

      <Separator />

      <Title>Basic elements</Title>
      <div className="h-stack w-full">
        <div className="w-96">
          <Preview>
            <form className="v-stack">
              <label className="label">
                <span className="label-text">First name</span>
                <input type="text" placeholder="John Doe" />
              </label>
              <label className="label">
                <span className="label-text">Message</span>
                <textarea placeholder="Dolor sit amet..." />
              </label>
              <label className="label">
                <span className="label-text">Choose a color</span>
                <select>
                  <option value="#ff0000" selected>
                    Red
                  </option>
                  <option value="#00ff00">Green</option>
                  <option value="#0000ff">Blue</option>
                </select>
              </label>
              <div>
                <span className="label-text">Choose a color</span>
                <label className="label-radio">
                  <input type="radio" name="color-radio" />
                  <span>Red</span>
                </label>
                <label className="label-radio">
                  <input type="radio" name="color-radio" />
                  <span>Green</span>
                </label>
                <label className="label-radio">
                  <input type="radio" name="color-radio" />
                  <span>Blue</span>
                </label>
              </div>
              <label className="label error">
                <span className="label-text">First name</span>
                <input type="text" placeholder="John Doe" />
                <span className="error">
                  Should be at least 1 character long
                </span>
              </label>
            </form>
          </Preview>
        </div>
        <div className="flex-1">
          <CodeBlock
            language="html"
            code={`<form className="v-stack">
<label className="label">
  <span className="label-text">First name</span>
  <input type="text" placeholder="John Doe" />
</label>
<label className="label">
  <span className="label-text">Message</span>
  <textarea placeholder="Dolor sit amet..." />
</label>
<label className="label">
  <span className="label-text">Choose a color</span>
  <select>
    <option value="#ff0000" selected>
      Red
    </option>
    <option value="#00ff00">Green</option>
    <option value="#0000ff">Blue</option>
  </select>
</label>
<div>
  <span className="label-text">Choose a color</span>
  <label className="label-radio">
    <input type="radio" name="color-radio" />
    <span>Red</span>
  </label>
  <label className="label-radio">
    <input type="radio" name="color-radio" />
    <span>Green</span>
  </label>
  <label className="label-radio">
    <input type="radio" name="color-radio" />
    <span>Blue</span>
  </label>
</div>
<label className="label error">
  <span className="label-text">First name</span>
  <input type="text" placeholder="John Doe" />
  <span className="error">Should be at least 1 character long</span>
</label>
</form>`}
          />
        </div>
      </div>
    </>
  )
}

const LogPagesForm = () => {
  const { register, handleSubmit, formState } = useForm()
  const onSubmit = (data: any) => console.log(data, 'submitted')

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="v-stack">
      <Input
        name="pagesRead"
        label="Pages read"
        register={register}
        formState={formState}
        type="number"
        options={{
          required: 'This field is required',
          valueAsNumber: true,
        }}
      />
      <Select
        name="medium"
        label="Medium"
        register={register}
        formState={formState}
        options={{
          required: true,
          valueAsNumber: true,
        }}
        values={[
          { value: '1', label: 'Book' },
          { value: '2', label: 'Comic' },
          { value: '3', label: 'Sentence' },
        ]}
      />
      <Input
        name="description"
        label="Description"
        register={register}
        formState={formState}
        type="text"
        placeholder="e.g. One Piece volume 45"
      />
      <button
        type="submit"
        className="btn primary"
        disabled={formState.isSubmitting}
      >
        Save changes
      </button>
    </form>
  )
}
