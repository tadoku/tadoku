import { useEffect, useState } from 'react'
import { FormProvider, useForm, useController } from 'react-hook-form'
import { Input, Loading, Select, Tabbar } from 'ui'
import {
  useAnnouncementCreate,
  useAnnouncementFind,
  useAnnouncementUpdate,
} from './api'
import { useNamespace } from '@app/content/NamespaceSelector'
import { useRouter } from 'next/router'
import { useQueryClient } from 'react-query'
import { toast } from 'react-toastify'
import { v4 as uuidv4 } from 'uuid'
import { routes } from '@app/common/routes'
import { MarkdownPreview } from '@app/content/MarkdownPreview'
import { CodeEditor } from '@app/content/CodeEditor'
import { markdown } from '@codemirror/lang-markdown'
import { languages } from '@codemirror/language-data'
import { toUtcISOStringFromLocal } from '@app/common/datetime'

const mdExtensions = [markdown({ codeLanguages: languages })]

const STYLE_OPTIONS = [
  { value: 'info', label: 'Info' },
  { value: 'success', label: 'Success' },
  { value: 'warning', label: 'Warning' },
  { value: 'error', label: 'Error' },
]

interface Props {
  id?: string
}

export function AnnouncementEditor({ id }: Props) {
  const namespace = useNamespace()
  const router = useRouter()
  const queryClient = useQueryClient()
  const isNew = !id

  const [mobileTab, setMobileTab] = useState<'edit' | 'preview'>('edit')
  const [initialized, setInitialized] = useState(false)

  const methods = useForm({
    defaultValues: {
      title: '',
      content: '',
      style: 'info',
      href: '',
      startsAt: '',
      endsAt: '',
    },
  })

  const contentController = useController({
    name: 'content',
    control: methods.control,
    rules: { validate: v => v.trim() !== '' || 'Content is required' },
  })

  const watchedContent = methods.watch('content')
  const watchedStyle = methods.watch('style')

  const existing = useAnnouncementFind(namespace, id ?? '', {
    enabled: !isNew && !!id && !!namespace,
  })

  useEffect(() => {
    if (existing.data && !initialized) {
      methods.reset({
        title: existing.data.title,
        content: existing.data.content,
        style: existing.data.style,
        href: existing.data.href ?? '',
        startsAt: existing.data.starts_at.slice(0, 16),
        endsAt: existing.data.ends_at.slice(0, 16),
      })
      setInitialized(true)
    }
  }, [existing.data, initialized, methods])

  const createMutation = useAnnouncementCreate(
    namespace,
    () => {
      toast.success('Announcement created successfully', { position: 'bottom-right' })
      queryClient.removeQueries(['announcements'])
    },
    () => {
      toast.error('Failed to create announcement', { position: 'bottom-right' })
    },
  )

  const updateMutation = useAnnouncementUpdate(
    namespace,
    () => {
      toast.success('Announcement updated successfully', { position: 'bottom-right' })
      queryClient.removeQueries(['announcements'])
    },
    () => {
      toast.error('Failed to update announcement', { position: 'bottom-right' })
    },
  )

  const isSaving = createMutation.isLoading || updateMutation.isLoading

  const handleSave = methods.handleSubmit(data => {
    const input = {
      id: isNew ? uuidv4() : id!,
      title: data.title.trim(),
      content: data.content,
      style: data.style,
      href: data.href.trim() || null,
      starts_at: toUtcISOStringFromLocal(data.startsAt),
      ends_at: toUtcISOStringFromLocal(data.endsAt),
    }

    const onSuccess = () => router.push(routes.announcements(namespace))

    if (isNew) {
      createMutation.mutate(input, { onSuccess })
    } else {
      updateMutation.mutate(input, { onSuccess })
    }
  })

  if (!isNew && existing.isLoading) {
    return <Loading />
  }

  if (!isNew && existing.isError) {
    return (
      <span className="flash error">
        {existing.error instanceof Error && existing.error.message === '404'
          ? 'Announcement not found.'
          : 'Could not load announcement.'}
      </span>
    )
  }

  return (
    <FormProvider {...methods}>
      <div className="flex flex-col gap-6">
        <div className="lg:hidden">
          <Tabbar
            alwaysExpanded
            links={[
              { label: 'Edit', active: mobileTab === 'edit', onClick: () => setMobileTab('edit') },
              { label: 'Preview', active: mobileTab === 'preview', onClick: () => setMobileTab('preview') },
            ]}
          />
        </div>

        <div className="flex flex-col lg:flex-row gap-6">
          <div className={`flex-1 min-w-0 flex-col ${mobileTab === 'preview' ? 'hidden lg:flex' : 'flex'}`}>
            <div className="card flex flex-col gap-4">
              <Input name="title" type="text" label="Title" placeholder="Announcement title (admin reference)" options={{ required: 'Title is required' }} />

              <Select name="style" label="Style" values={STYLE_OPTIONS} />

              <Input name="href" type="url" label="Link URL (optional)" placeholder="https://..." hint="If set, the announcement will link to this URL" />

              <Input name="startsAt" type="datetime-local" label="Starts At (UTC)" options={{ required: 'Start date is required' }} />
              <Input name="endsAt" type="datetime-local" label="Ends At (UTC)" options={{ required: 'End date is required', validate: (v) => { const startsAt = methods.getValues('startsAt'); if (startsAt && v && new Date(v) <= new Date(startsAt)) return 'End date must be after start date'; return true; } }} />

              <div className={`label flex-1 ${contentController.fieldState.error ? 'error' : ''}`}>
                <span className="label-text">Content (Markdown)</span>
                <CodeEditor
                  value={contentController.field.value}
                  onChange={contentController.field.onChange}
                  placeholder="Write your announcement content here..."
                  extensions={mdExtensions}
                />
                <span className="error">{contentController.fieldState.error?.message}</span>
              </div>

              <div className="flex items-center justify-end gap-3 pt-2">
                <button
                  type="button"
                  className="btn ghost"
                  onClick={() => router.back()}
                >
                  Cancel
                </button>
                <button
                  type="button"
                  className="btn primary"
                  onClick={handleSave}
                  disabled={isSaving}
                >
                  {isSaving ? 'Saving...' : isNew ? 'Create Announcement' : 'Save Announcement'}
                </button>
              </div>
            </div>
          </div>

          <div className={`flex-1 min-w-0 flex-col ${mobileTab === 'edit' ? 'hidden lg:flex' : 'flex'}`}>
            <div className="card flex-1 overflow-auto" style={{ minHeight: '300px' }}>
              <p className="text-xs text-slate-500 mb-2 uppercase tracking-wide">Preview</p>
              {watchedContent.trim() ? (
                <div className={`flash ${watchedStyle} mb-4`}>
                  <MarkdownPreview content={watchedContent} />
                </div>
              ) : (
                <p className="text-sm text-slate-400 italic">
                  Start typing to see a preview...
                </p>
              )}
            </div>
          </div>
        </div>
      </div>
    </FormProvider>
  )
}
