import { Loading, Modal } from 'ui'
import { ContentConfig, ContentItem } from './types'
import { useContentFindById, useContentUpdate } from './api'
import { useNamespace } from './NamespaceSelector'
import { DateTime } from 'luxon'
import Link from 'next/link'
import { useState } from 'react'
import { PencilIcon, TrashIcon } from '@heroicons/react/20/solid'
import { useQueryClient } from 'react-query'
import { toast } from 'react-toastify'

function StatusBadge({ item }: { item: ContentItem }) {
  if (!item.published_at) {
    return <span className="tag bg-amber-100 text-amber-800">Draft</span>
  }
  const publishedAt = DateTime.fromISO(item.published_at)
  if (publishedAt > DateTime.now()) {
    return <span className="tag bg-blue-100 text-blue-800">Scheduled</span>
  }
  return <span className="tag bg-emerald-100 text-emerald-800">Published</span>
}

function MetadataRow({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-1">
      <span className="text-xs font-semibold text-slate-500 uppercase tracking-wide">
        {label}
      </span>
      <span className="text-sm text-slate-900">{children}</span>
    </div>
  )
}

interface Props {
  config: ContentConfig
  id: string
}

export function ContentPreview({ config, id }: Props) {
  const [namespace] = useNamespace()
  const queryClient = useQueryClient()
  const [deleteModalOpen, setDeleteModalOpen] = useState(false)

  const item = useContentFindById(config, namespace, id, {
    enabled: !!id && !!namespace,
  })

  const togglePublishMutation = useContentUpdate(
    config,
    namespace,
    () => {
      toast.success(
        item.data?.published_at ? 'Unpublished successfully' : 'Published successfully',
      )
      queryClient.invalidateQueries([config.type])
    },
    () => {
      toast.error('Failed to update publish status')
    },
  )

  const handleTogglePublish = () => {
    if (!item.data) return
    togglePublishMutation.mutate({
      id: item.data.id,
      slug: item.data.slug,
      title: item.data.title,
      body: item.data.body,
      published_at: item.data.published_at ? null : new Date().toISOString(),
    })
  }

  if (item.isLoading || item.isIdle) {
    return <Loading />
  }

  if (item.isError) {
    return (
      <span className="flash error">
        {item.error instanceof Error && item.error.message === '404'
          ? `${config.label} not found.`
          : `Could not load ${config.label.toLowerCase()}.`}
      </span>
    )
  }

  const data = item.data

  return (
    <>
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 mb-6">
        <h1 className="title">{data.title}</h1>
        <div className="flex items-center gap-2 flex-shrink-0">
          <Link href={config.routes.edit(data.id)} className="btn secondary">
            <PencilIcon className="w-4 h-4 mr-1 inline" />
            Edit
          </Link>
          <button
            type="button"
            className="btn ghost"
            onClick={handleTogglePublish}
            disabled={togglePublishMutation.isLoading}
          >
            {data.published_at ? 'Unpublish' : 'Publish'}
          </button>
          <button
            type="button"
            className="btn danger"
            onClick={() => setDeleteModalOpen(true)}
          >
            <TrashIcon className="w-4 h-4 mr-1 inline" />
            Delete
          </button>
        </div>
      </div>

      <div className="flex flex-col lg:flex-row gap-6">
        {/* Content preview */}
        <div className="flex-1 min-w-0">
          <div className="card">
            {config.renderBody(data.body)}
          </div>
        </div>

        {/* Metadata panel */}
        <div className="w-full lg:w-72 flex-shrink-0">
          <div className="card">
            <h2 className="subtitle mb-4">Metadata</h2>
            <div className="flex flex-col gap-4">
              <MetadataRow label="Status">
                <StatusBadge item={data} />
              </MetadataRow>
              <MetadataRow label="Slug">
                <code className="text-sm bg-slate-100 px-2 py-0.5">{data.slug}</code>
              </MetadataRow>
              <MetadataRow label="Namespace">
                {data.namespace}
              </MetadataRow>
              <MetadataRow label="ID">
                <code className="text-xs bg-slate-100 px-2 py-0.5 break-all">{data.id}</code>
              </MetadataRow>
              {data.published_at ? (
                <MetadataRow label="Published">
                  {DateTime.fromISO(data.published_at).toLocaleString(DateTime.DATETIME_FULL)}
                </MetadataRow>
              ) : null}
              <MetadataRow label="Created">
                {data.created_at
                  ? DateTime.fromISO(data.created_at).toLocaleString(DateTime.DATETIME_FULL)
                  : 'N/A'}
              </MetadataRow>
              <MetadataRow label="Updated">
                {data.updated_at
                  ? DateTime.fromISO(data.updated_at).toLocaleString(DateTime.DATETIME_FULL)
                  : 'N/A'}
              </MetadataRow>
            </div>
          </div>
        </div>
      </div>

      <Modal
        isOpen={deleteModalOpen}
        setIsOpen={setDeleteModalOpen}
        title={`Delete ${config.label}`}
      >
        <p className="modal-body">
          Are you sure you want to delete &ldquo;{data.title}&rdquo;? This action is not yet
          supported by the API. Please contact a developer to delete content directly.
        </p>
        <div className="modal-actions">
          <button
            type="button"
            className="btn ghost"
            onClick={() => setDeleteModalOpen(false)}
          >
            Close
          </button>
        </div>
      </Modal>
    </>
  )
}
