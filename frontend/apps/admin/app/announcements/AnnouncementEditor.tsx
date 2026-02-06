import { useEffect, useState } from 'react'
import { Loading, Tabbar } from 'ui'
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

  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [style, setStyle] = useState('info')
  const [href, setHref] = useState('')
  const [startsAt, setStartsAt] = useState('')
  const [endsAt, setEndsAt] = useState('')
  const [initialized, setInitialized] = useState(false)

  const existing = useAnnouncementFind(namespace, id ?? '', {
    enabled: !isNew && !!id && !!namespace,
  })

  useEffect(() => {
    if (existing.data && !initialized) {
      setTitle(existing.data.title)
      setContent(existing.data.content)
      setStyle(existing.data.style)
      setHref(existing.data.href ?? '')
      setStartsAt(existing.data.starts_at.slice(0, 16))
      setEndsAt(existing.data.ends_at.slice(0, 16))
      setInitialized(true)
    }
  }, [existing.data, initialized])

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
  const [errors, setErrors] = useState<Record<string, string>>({})

  const handleSave = () => {
    const newErrors: Record<string, string> = {}
    if (!title.trim()) newErrors.title = 'Title is required'
    if (!content.trim()) newErrors.content = 'Content is required'
    if (!startsAt) newErrors.startsAt = 'Start date is required'
    if (!endsAt) newErrors.endsAt = 'End date is required'
    if (startsAt && endsAt && new Date(endsAt) <= new Date(startsAt)) {
      newErrors.endsAt = 'End date must be after start date'
    }

    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors)
      return
    }
    setErrors({})

    const input = {
      id: isNew ? uuidv4() : id!,
      title: title.trim(),
      content: content,
      style: style,
      href: href.trim() || null,
      starts_at: toUtcISOStringFromLocal(startsAt),
      ends_at: toUtcISOStringFromLocal(endsAt),
    }

    const onSuccess = () => router.push(routes.announcements(namespace))

    if (isNew) {
      createMutation.mutate(input, { onSuccess })
    } else {
      updateMutation.mutate(input, { onSuccess })
    }
  }

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
            <label className={`label ${errors.title ? 'error' : ''}`}>
              <span className="label-text">Title</span>
              <input
                type="text"
                className="input"
                value={title}
                onChange={e => setTitle(e.target.value)}
                placeholder="Announcement title (admin reference)"
              />
              <span className="error">{errors.title}</span>
            </label>

            <label className="label">
              <span className="label-text">Style</span>
              <select
                className="input"
                value={style}
                onChange={e => setStyle(e.target.value)}
              >
                {STYLE_OPTIONS.map(opt => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
            </label>

            <label className="label">
              <span className="label-text">Link URL (optional)</span>
              <input
                type="url"
                className="input"
                value={href}
                onChange={e => setHref(e.target.value)}
                placeholder="https://..."
              />
              <span className="text-xs text-slate-500">
                If set, the announcement will link to this URL
              </span>
            </label>

            <div className="flex flex-col sm:flex-row gap-4">
              <label className={`label flex-1 ${errors.startsAt ? 'error' : ''}`}>
                <span className="label-text">Starts At (UTC)</span>
                <input
                  type="datetime-local"
                  className="input"
                  value={startsAt}
                  onChange={e => setStartsAt(e.target.value)}
                />
                <span className="error">{errors.startsAt}</span>
              </label>
              <label className={`label flex-1 ${errors.endsAt ? 'error' : ''}`}>
                <span className="label-text">Ends At (UTC)</span>
                <input
                  type="datetime-local"
                  className="input"
                  value={endsAt}
                  onChange={e => setEndsAt(e.target.value)}
                />
                <span className="error">{errors.endsAt}</span>
              </label>
            </div>

            <div className={`label flex-1 ${errors.content ? 'error' : ''}`}>
              <span className="label-text">Content (Markdown)</span>
              <CodeEditor
                value={content}
                onChange={setContent}
                placeholder="Write your announcement content here..."
                extensions={mdExtensions}
              />
              <span className="error">{errors.content}</span>
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
            {content.trim() ? (
              <div className={`flash ${style} mb-4`}>
                <MarkdownPreview content={content} />
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
  )
}
