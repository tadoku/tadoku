import { useEffect, useState } from 'react'
import { Loading, Tabbar } from 'ui'
import { ContentConfig } from './types'
import { useContentCreate, useContentFindById, useContentUpdate } from './api'
import { useNamespace } from './NamespaceSelector'
import { useRouter } from 'next/router'
import { useQueryClient } from 'react-query'
import { toast } from 'react-toastify'
import { v4 as uuidv4 } from 'uuid'

interface Props {
  config: ContentConfig
  id?: string // undefined = new item, string = editing existing
}

export function ContentEditor({ config, id }: Props) {
  const [namespace] = useNamespace()
  const router = useRouter()
  const queryClient = useQueryClient()
  const isNew = !id

  // Mobile tab state
  const [mobileTab, setMobileTab] = useState<'content' | 'preview'>('content')

  // Form state
  const [title, setTitle] = useState('')
  const [itemSlug, setItemSlug] = useState('')
  const [body, setBody] = useState('')
  const [publishedAt, setPublishedAt] = useState('')
  const [initialized, setInitialized] = useState(false)

  // Load existing item if editing
  const existing = useContentFindById(config, namespace, id ?? '', {
    enabled: !isNew && !!id && !!namespace,
  })

  // Populate form when data loads
  useEffect(() => {
    if (existing.data && !initialized) {
      setTitle(existing.data.title)
      setItemSlug(existing.data.slug)
      setBody(existing.data.body)
      setPublishedAt(
        existing.data.published_at
          ? existing.data.published_at.slice(0, 16) // format for datetime-local input
          : '',
      )
      setInitialized(true)
    }
  }, [existing.data, initialized])

  const createMutation = useContentCreate(
    config,
    namespace,
    () => {
      toast.success(`${config.label} created successfully`, { position: 'bottom-right' })
      queryClient.invalidateQueries([config.type])
    },
    () => {
      toast.error(`Failed to create ${config.label.toLowerCase()}`, { position: 'bottom-right' })
    },
  )

  const updateMutation = useContentUpdate(
    config,
    namespace,
    () => {
      toast.success(`${config.label} updated successfully`, { position: 'bottom-right' })
      queryClient.invalidateQueries([config.type])
    },
    () => {
      toast.error(`Failed to update ${config.label.toLowerCase()}`, { position: 'bottom-right' })
    },
  )

  const isSaving = createMutation.isLoading || updateMutation.isLoading
  const [errors, setErrors] = useState<Record<string, string>>({})

  const handleSave = () => {
    const newErrors: Record<string, string> = {}
    if (!title.trim()) newErrors.title = 'Title is required'
    if (!itemSlug.trim()) newErrors.slug = 'Slug is required'
    if (!body.trim()) newErrors.body = 'Content is required'

    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors)
      return
    }
    setErrors({})

    const itemId = isNew ? uuidv4() : existing.data!.id
    const input = {
      id: itemId,
      slug: itemSlug.trim().toLowerCase(),
      title: title.trim(),
      body: body,
      published_at: publishedAt ? new Date(publishedAt).toISOString() : null,
    }

    const onSuccess = () => router.push(config.routes.preview(itemId))

    if (isNew) {
      createMutation.mutate(input, { onSuccess })
    } else {
      updateMutation.mutate(input, { onSuccess })
    }
  }

  const handleTitleChange = (value: string) => {
    setTitle(value)
    setItemSlug(
      value
        .toLowerCase()
        .replace(/[^a-z0-9\s-]/g, '')
        .replace(/\s+/g, '-')
        .replace(/-+/g, '-')
        .replace(/^-|-$/g, ''),
    )
  }

  if (!isNew && existing.isLoading) {
    return <Loading />
  }

  if (!isNew && existing.isError) {
    return (
      <span className="flash error">
        {existing.error instanceof Error && existing.error.message === '404'
          ? `${config.label} not found.`
          : `Could not load ${config.label.toLowerCase()}.`}
      </span>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      {/* Mobile tab bar */}
      <div className="lg:hidden">
        <Tabbar
          alwaysExpanded
          links={[
            { label: 'Content', active: mobileTab === 'content', onClick: () => setMobileTab('content') },
            { label: 'Preview', active: mobileTab === 'preview', onClick: () => setMobileTab('preview') },
          ]}
        />
      </div>

      {/* Editor and preview */}
      <div className="flex flex-col lg:flex-row gap-6">
        {/* Editor */}
        <div className={`flex-1 min-w-0 flex-col ${mobileTab === 'preview' ? 'hidden lg:flex' : 'flex'}`}>
          <label className={`label flex-1 ${errors.body ? 'error' : ''}`}>
            <span className="label-text">Content</span>
            <textarea
              className="input font-mono text-sm flex-1"
              style={{ minHeight: '500px' }}
              value={body}
              onChange={e => setBody(e.target.value)}
              placeholder={`Write your ${config.label.toLowerCase()} content here...`}
            />
            <span className="error">{errors.body}</span>
          </label>
        </div>

        {/* Live preview */}
        <div className={`flex-1 min-w-0 flex-col ${mobileTab === 'content' ? 'hidden lg:flex' : 'flex'}`}>
          <span className="text-sm font-semibold text-slate-600 mb-2">
            Preview
          </span>
          <div className="card flex-1 overflow-auto" style={{ minHeight: '500px' }}>
            {body.trim() ? (
              config.renderBody(body)
            ) : (
              <p className="text-sm text-slate-400 italic">
                Start typing to see a preview...
              </p>
            )}
          </div>
        </div>
      </div>

      {/* Metadata */}
      <div className="card">
        <h2 className="subtitle mb-4">Metadata</h2>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <label className={`label ${errors.title ? 'error' : ''}`}>
            <span className="label-text">Title</span>
            <input
              type="text"
              className="input"
              value={title}
              onChange={e => handleTitleChange(e.target.value)}
              placeholder="Post title"
            />
            <span className="error">{errors.title}</span>
          </label>
          <label className={`label ${errors.slug ? 'error' : ''}`}>
            <span className="label-text">Slug</span>
            <input
              type="text"
              className="input"
              value={itemSlug}
              onChange={e => setItemSlug(e.target.value)}
              placeholder="url-friendly-slug"
            />
            <span className="error">{errors.slug}</span>
          </label>
          <label className="label">
            <span className="label-text">Publish Date (UTC)</span>
            <input
              type="datetime-local"
              className="input"
              value={publishedAt}
              onChange={e => setPublishedAt(e.target.value)}
            />
            <span className="text-xs text-slate-500">Leave empty to save as draft</span>
          </label>
        </div>
      </div>

      {/* Actions */}
      <div className="flex items-center justify-end gap-3">
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
          {isSaving ? 'Saving...' : isNew ? `Create ${config.label}` : `Save ${config.label}`}
        </button>
      </div>
    </div>
  )
}
