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
import { FormProvider, useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { RadioGroup } from 'ui/components/Form'
import {
  AdjustmentsHorizontalIcon,
  LinkIcon,
  UserIcon,
} from '@heroicons/react/20/solid'

export default function Forms() {
  return (
    <>
      <h1 className="title mb-8">Forms</h1>

      <Title>React example: Log pages form</Title>
      <div className="w-full">
        <div className="w-full max-w-xl">
          <Preview>
            <LogPagesForm />
          </Preview>
        </div>
        <div>
          <CodeBlock
            language="typescript"
            code={`import { Input, Select } from '@components/Form'
import { FormProvider, useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod'

const LogPagesFormSchema = z.object({
  languageCode: z.string(),
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
  const methods = useForm({
    resolver: zodResolver(LogPagesFormSchema),
  })
  const onSubmit = (data: any) => console.log(data, 'submitted')

  const languages = [
    { value: 'jpa', label: 'Japanese' },
    { value: 'zho', label: 'Chinese' },
    { value: 'kor', label: 'Korean' },
  ]
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
    <FormProvider {...methods}>
      <form
        onSubmit={methods.handleSubmit(onSubmit)}
        className="v-stack spaced"
      >
        <Select name="languageCode" label="Language" values={languages} />
        <Select
          name="activity"
          label="Activity"
          values={activities}
          options={{ valueAsNumber: true }}
        />
        <div className="h-stack spaced">
          <div className="flex-grow">
            <Input
              name="amount"
              label="Amount"
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
              values={units}
              options={{ valueAsNumber: true }}
            />
          </div>
        </div>
        <AutocompleteMultiInput
          name="tags"
          label="Tags"
          options={tags}
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
          type="text"
          placeholder="e.g. One Piece volume 45"
        />
        <button
          type="submit"
          className="btn primary"
          disabled={methods.formState.isSubmitting}
        >
          Save changes
        </button>
      </form>
    </FormProvider>
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
import { FormProvider, useForm } from 'react-hook-form'


const AutocompleteForm = () => {
  const methods = useForm()
  const onSubmit = (data: any) => console.log(data, 'submitted')

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
import { FormProvider, useForm } from 'react-hook-form'

const ComposeBlogPostForm = () => {
  const methods = useForm()
  const onSubmit = (data: any) => console.log(data, 'submitted')

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
import { FormProvider, useForm } from 'react-hook-form'

const MiscForm = () => {
  const methods = useForm()
  const onSubmit = (data: any) => console.log(data, 'submitted')

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
                <input className="input" type="text" placeholder="John Doe" />
              </label>
              <label className="label">
                <span className="label-text">Message</span>
                <textarea className="input" placeholder="Dolor sit amet..." />
              </label>
              <label className="label">
                <span className="label-text">Choose a color</span>
                <select className="input">
                  <option value="#ff0000">Red</option>
                  <option value="#00ff00">Green</option>
                  <option value="#0000ff">Blue</option>
                </select>
              </label>
              <div>
                <span className="label-text">Choose a color</span>
                <div className="v-stack">
                  <label className="label-inline">
                    <input type="radio" name="color-radio" className="input" />
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
              </div>
              <label className="label error">
                <span className="label-text">First name</span>
                <input type="text" placeholder="John Doe" className="input" />
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
  <input className="input" type="text" placeholder="John Doe" />
</label>
<label className="label">
  <span className="label-text">Message</span>
  <textarea className="input" placeholder="Dolor sit amet..." />
</label>
<label className="label">
  <span className="label-text">Choose a color</span>
  <select className="input">
    <option value="#ff0000">Red</option>
    <option value="#00ff00">Green</option>
    <option value="#0000ff">Blue</option>
  </select>
</label>
<div>
  <span className="label-text">Choose a color</span>
  <div className="v-stack">
    <label className="label-inline">
      <input type="radio" name="color-radio" className="input" />
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
</div>
<label className="label error">
  <span className="label-text">First name</span>
  <input type="text" placeholder="John Doe" className="input" />
  <span className="error">
    Should be at least 1 character long
  </span>
</label>
</form>`}
          />
        </div>
      </div>
    </>
  )
}

const LogPagesFormSchema = z.object({
  trackingModeSelection: z.enum(['automatic', 'manual', 'personal']),
  languageCode: z.string(),
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
  const methods = useForm({
    resolver: zodResolver(LogPagesFormSchema),
  })
  const onSubmit = (data: any) => console.log(data, 'submitted')

  const languages = [
    { value: 'jpa', label: 'Japanese' },
    { value: 'zho', label: 'Chinese' },
    { value: 'kor', label: 'Korean' },
  ]
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
    <FormProvider {...methods}>
      <form
        onSubmit={methods.handleSubmit(onSubmit)}
        className="v-stack spaced"
      >
        <RadioGroup
          options={[
            {
              value: 'automatic',
              label: 'Automatic',
              description: 'Submit log to all eligible contests',
              IconComponent: LinkIcon,
            },
            {
              value: 'manual',
              label: 'Manual',
              description: 'Choose which contests to submit to',
              IconComponent: AdjustmentsHorizontalIcon,
            },
            {
              value: 'personal',
              label: 'Personal',
              description: 'Do not submit to any contests',
              IconComponent: UserIcon,
            },
          ]}
          label="Where to track?"
          name="trackingModeSelection"
          defaultValue="automatic"
        />
        <Select name="languageCode" label="Language" values={languages} />
        <Select
          name="activity"
          label="Activity"
          values={activities}
          options={{ valueAsNumber: true }}
        />
        <div className="h-stack spaced">
          <div className="flex-grow">
            <Input
              name="amount"
              label="Amount"
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
              values={units}
              options={{ valueAsNumber: true }}
            />
          </div>
        </div>
        <AutocompleteMultiInput
          name="tags"
          label="Tags"
          options={tags}
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
          type="text"
          placeholder="e.g. One Piece volume 45"
        />
        <button
          type="submit"
          className="btn primary"
          disabled={methods.formState.isSubmitting}
        >
          Save changes
        </button>
      </form>
    </FormProvider>
  )
}

const ComposeBlogPostForm = () => {
  const methods = useForm()
  const onSubmit = (data: any) => console.log(data, 'submitted')

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

const AutocompleteForm = () => {
  const methods = useForm()
  const onSubmit = (data: any) => console.log(data, 'submitted')

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

const MiscForm = () => {
  const methods = useForm()
  const onSubmit = (data: any) => console.log(data, 'submitted')

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
