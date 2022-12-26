import { CodeBlock, Preview, Separator, Title } from '@components/example'
import {
  AutocompleteInput,
  AutocompleteMultiInput,
  Checkbox,
  Input,
  RadioSelect,
  Select,
  TextArea,
} from 'ui'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'

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
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod'

const LogPagesFormSchema = z.object({
  pagesRead: z
    .string()
    .refine(val => !Number.isNaN(parseInt(val, 10)), {
      message: 'Amount is required',
    })
    .transform(v => parseInt(v, 10)),
  medium: z.number(),
  description: z.string().optional(),
})

const LogPagesForm = () => {
  const { register, handleSubmit, formState } = useForm({
    resolver: zodResolver(LogPagesFormSchema),
  })
  const onSubmit = (data: any) => console.log(data, 'submitted')

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="v-stack spaced">
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

      <Title>React example: Autocomplete</Title>
      <div className="h-stack w-full">
        <div className="w-96">
          <Preview>
            <AutocompleteForm />
          </Preview>
        </div>
        <div className="flex-1">
          <CodeBlock
            language="typescript"
            code={`import { AutocompleteInput, AutocompleteMultiInput } from '@components/Form'
import { useForm } from 'react-hook-form'

const AutocompleteForm = () => {
  const { handleSubmit, formState, control } = useForm()
  const onSubmit = (data: any) => console.log(data, 'submitted')

  const tags = ['Book', 'Ebook', 'Fiction', 'Non-fiction', 'Web page', 'Lyric']
  const activities = [
    { id: 1, name: 'Reading' },
    { id: 2, name: 'Listening' },
    { id: 3, name: 'Speaking' },
    { id: 4, name: 'Writing' },
  ]

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="v-stack spaced">
      <AutocompleteInput
        name="tags"
        label="Tags"
        options={tags}
        control={control}
        rules={{ required: 'Required' }}
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
        control={control}
        match={(option, query) =>
          option.name
            .toLowerCase()
            .replace(/[^a-zA-Z0-9]/g, '')
            .includes(query.toLowerCase())
        }
        getIdForOption={option => option.id}
        format={option => option.name}
      />
      <button
        type="submit"
        className="btn primary"
        disabled={formState.isSubmitting}
      >
        Submit
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
          <Preview>
            <ComposeBlogPostForm />
          </Preview>
        </div>
        <div className="flex-1">
          <CodeBlock
            language="typescript"
            code={`import { Checkbox, Input, TextArea } from '@components/Form'
import { useForm } from 'react-hook-form'

const ComposeBlogPostForm = () => {
  const { register, handleSubmit, formState } = useForm()
  const onSubmit = (data: any) => console.log(data, 'submitted')

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="v-stack spaced">
      <Input
        name="title"
        label="Title"
        register={register}
        formState={formState}
        type="text"
        options={{
          required: true,
        }}
      />
      <TextArea
        name="content"
        label="Content"
        register={register}
        formState={formState}
        options={{
          required: true,
        }}
      />
      <Input
        name="publishedAt"
        label="Published at"
        register={register}
        formState={formState}
        type="date"
        options={{
          required: true,
          valueAsDate: true,
        }}
      />
      <Checkbox
        name="isPublished"
        label="Published"
        register={register}
        formState={formState}
      />
      <button
        type="submit"
        className="btn primary"
        disabled={formState.isSubmitting}
      >
        Save
      </button>
    </form>
  )
}`}
          />
        </div>
      </div>

      <Separator />

      <Title>React example: other elements</Title>
      <div className="h-stack w-full">
        <div className="w-96">
          <Preview>
            <MiscForm />
          </Preview>
        </div>
        <div className="flex-1">
          <CodeBlock
            language="typescript"
            code={`import { RadioSelect } from '@components/Form'
import { useForm } from 'react-hook-form'

const MiscForm = () => {
  const { register, handleSubmit, formState } = useForm()
  const onSubmit = (data: any) => console.log(data, 'submitted')

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="v-stack spaced">
      <RadioSelect
        name="cardType"
        label="Card type"
        register={register}
        formState={formState}
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
        register={register}
        formState={formState}
        type="datetime-local"
        options={{
          required: true,
          valueAsDate: true,
        }}
      />
      <Input
        name="time"
        label="Time"
        register={register}
        formState={formState}
        type="time"
        options={{
          required: true,
        }}
      />
      <Input
        name="week"
        label="Week"
        register={register}
        formState={formState}
        type="week"
        options={{
          required: true,
        }}
      />
      <Input
        name="month"
        label="Month"
        register={register}
        formState={formState}
        type="month"
        options={{
          required: true,
        }}
      />
      <Input
        name="color"
        label="Color"
        register={register}
        formState={formState}
        type="color"
        options={{
          required: true,
        }}
      />
      <Input
        name="email"
        label="Email"
        register={register}
        formState={formState}
        type="email"
        options={{
          required: true,
        }}
      />
      <Input
        name="file"
        label="File"
        register={register}
        formState={formState}
        type="file"
        options={{
          required: true,
        }}
      />
      <Input
        name="range"
        label="Range"
        register={register}
        formState={formState}
        type="range"
        options={{
          required: true,
        }}
      />
      <Input
        name="search"
        label="Search"
        register={register}
        formState={formState}
        type="search"
        options={{
          required: true,
        }}
      />
      <Input
        name="tel"
        label="tel"
        register={register}
        formState={formState}
        type="tel"
        options={{
          required: true,
        }}
      />
      <Input
        name="url"
        label="url"
        register={register}
        formState={formState}
        type="url"
        options={{
          required: true,
        }}
      />
      <button
        type="submit"
        className="btn primary"
        disabled={formState.isSubmitting}
      >
        Submit
      </button>
    </form>
  )
}`}
          />
        </div>
      </div>

      <Separator />

      <Title>Basic elements</Title>
      <div className="h-stack w-full">
        <div className="w-96">
          <Preview>
            <form className="v-stack spaced">
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
                  <option value="#ff0000">Red</option>
                  <option value="#00ff00">Green</option>
                  <option value="#0000ff">Blue</option>
                </select>
              </label>
              <div>
                <span className="label-text">Choose a color</span>
                <label className="label-inline">
                  <input type="radio" name="color-radio" />
                  <span>Red</span>
                </label>
                <label className="label-inline">
                  <input type="radio" name="color-radio" />
                  <span>Green</span>
                </label>
                <label className="label-inline">
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
            code={`<form className="v-stack spaced">
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
  <label className="label-inline">
    <input type="radio" name="color-radio" />
    <span>Red</span>
  </label>
  <label className="label-inline">
    <input type="radio" name="color-radio" />
    <span>Green</span>
  </label>
  <label className="label-inline">
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

const LogPagesFormSchema = z.object({
  activity: z.number(),
  amount: z.number().positive(),
  unit: z.number(),
  tags: z
    .array(z.string())
    .min(1, 'Must select at least one tag')
    .max(3, 'Must select three or fewer'),
  description: z.string().optional(),
})

const LogPagesForm = () => {
  const { register, handleSubmit, formState, control } = useForm({
    resolver: zodResolver(LogPagesFormSchema),
  })
  const onSubmit = (data: any) => console.log(data, 'submitted')

  const units = [
    { value: '1', label: 'Pages' },
    { value: '2', label: 'Comic pages' },
  ]
  const tags = ['Book', 'Ebook', 'Fiction', 'Non-fiction', 'Web page', 'Lyric']
  const activities = [
    { value: '1', label: 'Reading' },
    { value: '2', label: 'Listening' },
    { value: '3', label: 'Speaking' },
    { value: '4', label: 'Writing' },
  ]

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="v-stack spaced">
      <Select
        name="activity"
        label="Activity"
        register={register}
        formState={formState}
        values={activities}
        options={{ valueAsNumber: true }}
      />
      <div className="h-stack spaced">
        <div className="flex-grow">
          <Input
            name="amount"
            label="Amount"
            register={register}
            formState={formState}
            type="number"
            defaultValue={0}
            options={{ valueAsNumber: true }}
            min={0}
          />
        </div>
        <div className="min-w-[150px]">
          <Select
            name="unit"
            label="Unit"
            register={register}
            formState={formState}
            values={units}
            options={{ valueAsNumber: true }}
          />
        </div>
      </div>
      <AutocompleteMultiInput
        name="tags"
        label="Tags"
        options={tags}
        control={control}
        match={(option, query) =>
          option
            .toLowerCase()
            .replace(/[^a-zA-Z0-9]/g, '')
            .includes(query.toLowerCase())
        }
        getIdForOption={option => option}
        format={option => option}
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

const ComposeBlogPostForm = () => {
  const { register, handleSubmit, formState } = useForm()
  const onSubmit = (data: any) => console.log(data, 'submitted')

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="v-stack spaced">
      <Input
        name="title"
        label="Title"
        register={register}
        formState={formState}
        type="text"
        options={{
          required: true,
        }}
      />
      <TextArea
        name="content"
        label="Content"
        register={register}
        formState={formState}
        options={{
          required: true,
        }}
      />
      <Input
        name="publishedAt"
        label="Published at"
        register={register}
        formState={formState}
        type="date"
        options={{
          required: true,
          valueAsDate: true,
        }}
      />
      <Checkbox
        name="isPublished"
        label="Published"
        register={register}
        formState={formState}
      />
      <button
        type="submit"
        className="btn primary"
        disabled={formState.isSubmitting}
      >
        Save
      </button>
    </form>
  )
}

const AutocompleteForm = () => {
  const { handleSubmit, formState, control } = useForm()
  const onSubmit = (data: any) => console.log(data, 'submitted')

  const tags = ['Book', 'Ebook', 'Fiction', 'Non-fiction', 'Web page', 'Lyric']
  const activities = [
    { id: 1, name: 'Reading' },
    { id: 2, name: 'Listening' },
    { id: 3, name: 'Speaking' },
    { id: 4, name: 'Writing' },
  ]

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="v-stack spaced">
      <AutocompleteInput
        name="tags"
        label="Tags"
        options={tags}
        control={control}
        rules={{ required: 'Required' }}
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
        control={control}
        match={(option, query) =>
          option.name
            .toLowerCase()
            .replace(/[^a-zA-Z0-9]/g, '')
            .includes(query.toLowerCase())
        }
        getIdForOption={option => option.id}
        format={option => option.name}
      />
      <button
        type="submit"
        className="btn primary"
        disabled={formState.isSubmitting}
      >
        Submit
      </button>
    </form>
  )
}

const MiscForm = () => {
  const { register, handleSubmit, formState } = useForm()
  const onSubmit = (data: any) => console.log(data, 'submitted')

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="v-stack spaced">
      <RadioSelect
        name="cardType"
        label="Card type"
        register={register}
        formState={formState}
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
        register={register}
        formState={formState}
        type="datetime-local"
        options={{
          required: true,
          valueAsDate: true,
        }}
      />
      <Input
        name="time"
        label="Time"
        register={register}
        formState={formState}
        type="time"
        options={{
          required: true,
        }}
      />
      <Input
        name="week"
        label="Week"
        register={register}
        formState={formState}
        type="week"
        options={{
          required: true,
        }}
      />
      <Input
        name="month"
        label="Month"
        register={register}
        formState={formState}
        type="month"
        options={{
          required: true,
        }}
      />
      <Input
        name="color"
        label="Color"
        register={register}
        formState={formState}
        type="color"
        options={{
          required: true,
        }}
      />
      <Input
        name="email"
        label="Email"
        register={register}
        formState={formState}
        type="email"
        options={{
          required: true,
        }}
      />
      <Input
        name="file"
        label="File"
        register={register}
        formState={formState}
        type="file"
        options={{
          required: true,
        }}
      />
      <Input
        name="range"
        label="Range"
        register={register}
        formState={formState}
        type="range"
        options={{
          required: true,
        }}
      />
      <Input
        name="search"
        label="Search"
        register={register}
        formState={formState}
        type="search"
        options={{
          required: true,
        }}
      />
      <Input
        name="tel"
        label="tel"
        register={register}
        formState={formState}
        type="tel"
        options={{
          required: true,
        }}
      />
      <Input
        name="url"
        label="url"
        register={register}
        formState={formState}
        type="url"
        options={{
          required: true,
        }}
      />
      <button
        type="submit"
        className="btn primary"
        disabled={formState.isSubmitting}
      >
        Submit
      </button>
    </form>
  )
}
