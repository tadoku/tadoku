import { useEffect, useState } from 'react'
import { Loading } from 'ui'
import { ContentConfig } from './types'
import { useContentCreate, useContentFindById, useContentUpdate } from './api'
import { useNamespace } from './NamespaceSelector'
import { useRouter } from 'next/router'
import { useQueryClient } from 'react-query'
import { toast } from 'react-toastify'

interface Props {
  config: ContentConfig
  id?: string // undefined = new item, string = editing existing
}

export function ContentEditor({ config, id }: Props) {
  const [namespace] = useNamespace()
  const router = useRouter()
  const queryClient = useQueryClient()
  const isNew = !id

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
      toast.success(`${config.label} created successfully`)
      queryClient.invalidateQueries([config.type])
    },
    () => {
      toast.error(`Failed to create ${config.label.toLowerCase()}`)
    },
  )

  const updateMutation = useContentUpdate(
    config,
    namespace,
    () => {
      toast.success(`${config.label} updated successfully`)
      queryClient.invalidateQueries([config.type])
    },
    () => {
      toast.error(`Failed to update ${config.label.toLowerCase()}`)
    },
  )

  const isSaving = createMutation.isLoading || updateMutation.isLoading

  const handleSave = () => {
    if (!title.trim() || !itemSlug.trim() || !body.trim()) {
      toast.error('Title, slug, and content are required')
      return
    }

    const itemId = isNew ? crypto.randomUUID() : existing.data!.id
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

  const handleSaveAsDraft = () => {
    if (!title.trim() || !itemSlug.trim() || !body.trim()) {
      toast.error('Title, slug, and content are required')
      return
    }

    const itemId = isNew ? crypto.randomUUID() : existing.data!.id
    const input = {
      id: itemId,
      slug: itemSlug.trim().toLowerCase(),
      title: title.trim(),
      body: body,
      published_at: null,
    }

    const onSuccess = () => router.push(config.routes.preview(itemId))

    if (isNew) {
      createMutation.mutate(input, { onSuccess })
    } else {
      updateMutation.mutate(input, { onSuccess })
    }
  }

  // Auto-generate slug from title for new items
  const handleTitleChange = (value: string) => {
    setTitle(value)
    if (isNew && !initialized) {
      setItemSlug(
        value
          .toLowerCase()
          .replace(/[^a-z0-9\s-]/g, '')
          .replace(/\s+/g, '-')
          .replace(/-+/g, '-')
          .replace(/^-|-$/g, ''),
      )
    }
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
      {/* Editor and preview */}
      <div className="flex flex-col lg:flex-row gap-6">
        {/* Editor */}
        <div className="flex-1 min-w-0 flex flex-col">
          <label className="text-sm font-semibold text-slate-600 mb-2">
            Content
          </label>
          <textarea
            className="input font-mono text-sm flex-1"
            style={{ minHeight: '500px' }}
            value={body}
            onChange={e => setBody(e.target.value)}
            placeholder={`Write your ${config.label.toLowerCase()} content here...`}
          />
        </div>

        {/* Live preview */}
        <div className="flex-1 min-w-0 flex flex-col">
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
          <label className="label">
            <span className="label-text">Title</span>
            <input
              type="text"
              className="input"
              value={title}
              onChange={e => handleTitleChange(e.target.value)}
              placeholder="Post title"
            />
          </label>
          <label className="label">
            <span className="label-text">Slug</span>
            <input
              type="text"
              className="input"
              value={itemSlug}
              onChange={e => setItemSlug(e.target.value)}
              placeholder="url-friendly-slug"
            />
          </label>
          <label className="label">
            <span className="label-text">Publish Date</span>
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
      <div className="flex items-center gap-3">
        <button
          type="button"
          className="btn primary"
          onClick={handleSave}
          disabled={isSaving}
        >
          {isSaving ? 'Saving...' : isNew ? `Create ${config.label}` : `Save ${config.label}`}
        </button>
        <button
          type="button"
          className="btn secondary"
          onClick={handleSaveAsDraft}
          disabled={isSaving}
        >
          Save as Draft
        </button>
        <button
          type="button"
          className="btn ghost"
          onClick={() => router.back()}
        >
          Cancel
        </button>
      </div>
    </div>
  )
}
