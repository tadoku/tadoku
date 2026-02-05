import { Loading } from 'ui'
import { ContentConfig, ContentItem } from './types'
import { useContentFindById, useContentUpdate, useContentVersionList, useContentVersionGet } from './api'
import { useNamespace } from './NamespaceSelector'
import { DateTime } from 'luxon'
import Link from 'next/link'
import { ArrowUturnLeftIcon, PencilIcon } from '@heroicons/react/20/solid'
import { useQueryClient } from 'react-query'
import { toast } from 'react-toastify'
import { useState } from 'react'

function StatusBadge({ item }: { item: ContentItem }) {
  if (!item.published_at) {
    return <span className="tag bg-amber-100 text-amber-800 shadow-sm">Draft</span>
  }
  const publishedAt = DateTime.fromISO(item.published_at)
  if (publishedAt > DateTime.now()) {
    return <span className="tag bg-blue-100 text-blue-800 shadow-sm">Scheduled</span>
  }
  return <span className="tag bg-emerald-100 text-emerald-800 shadow-sm">Published</span>
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
  const [selectedVersionId, setSelectedVersionId] = useState<string | null>(null)
  const [versionPage, setVersionPage] = useState(0)
  const versionsPerPage = 10

  const item = useContentFindById(config, namespace, id, {
    enabled: !!id && !!namespace,
  })

  const versions = useContentVersionList(config, namespace, id, {
    enabled: !!id && !!namespace,
  })

  const selectedVersion = useContentVersionGet(
    config,
    namespace,
    id,
    selectedVersionId,
    { enabled: !!selectedVersionId },
  )

  const togglePublishMutation = useContentUpdate(
    config,
    namespace,
    () => {
      toast.success(
        item.data?.published_at ? 'Unpublished successfully' : 'Published successfully',
        { position: 'bottom-right' },
      )
      queryClient.removeQueries([config.type, 'findById', namespace, id])
      queryClient.invalidateQueries([config.type])
    },
    () => {
      toast.error('Failed to update publish status', { position: 'bottom-right' })
    },
  )

  const restoreVersionMutation = useContentUpdate(
    config,
    namespace,
    () => {
      toast.success('Version restored successfully', { position: 'bottom-right' })
      setSelectedVersionId(null)
      setVersionPage(0)
      queryClient.removeQueries([config.type, 'findById', namespace, id])
      queryClient.invalidateQueries([config.type])
    },
    () => {
      toast.error('Failed to restore version', { position: 'bottom-right' })
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

  const handleRestoreVersion = () => {
    if (!item.data || !selectedVersion.data) return
    restoreVersionMutation.mutate({
      id: item.data.id,
      slug: item.data.slug,
      title: selectedVersion.data.title,
      body: selectedVersion.data.body,
      published_at: item.data.published_at,
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
  const latestVersionId = versions.data?.[versions.data.length - 1]?.id ?? null
  const isViewingVersion = selectedVersionId !== null && selectedVersionId !== latestVersionId
  const previewBody = isViewingVersion ? (selectedVersion.data?.body ?? data.body) : data.body
  const previewTitle = isViewingVersion ? (selectedVersion.data?.title ?? data.title) : data.title

  return (
    <>
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 mb-6">
        <div>
          <h1 className="title">{config.label} preview</h1>
          <div className="flex items-center gap-2 mt-1">
            <StatusBadge item={data} />
            <span className="text-sm text-slate-400">/{data.slug}</span>
          </div>
        </div>
        <div className="flex items-center gap-2 flex-shrink-0">
          {isViewingVersion ? (
            <>
              <button
                type="button"
                className="btn ghost"
                onClick={() => setSelectedVersionId(null)}
              >
                <ArrowUturnLeftIcon className="w-4 h-4 mr-1 inline" />
                Back to current
              </button>
              <button
                type="button"
                className="btn secondary"
                onClick={handleRestoreVersion}
                disabled={restoreVersionMutation.isLoading || !selectedVersion.data}
              >
                {restoreVersionMutation.isLoading ? 'Restoring...' : 'Restore this version'}
              </button>
            </>
          ) : (
            <>
              <button
                type="button"
                className="btn"
                onClick={handleTogglePublish}
                disabled={togglePublishMutation.isLoading}
              >
                {data.published_at ? 'Unpublish' : 'Publish'}
              </button>
              <Link href={config.routes.edit(data.id)} className="btn secondary">
                <PencilIcon className="w-4 h-4 mr-1 inline" />
                Edit
              </Link>
            </>
          )}
        </div>
      </div>

      <div className="flex flex-col lg:flex-row gap-6">
        {/* Content preview */}
        <div className="flex-1 min-w-0">
          <div className={`card ${isViewingVersion ? 'bg-amber-50' : ''}`}>
            {selectedVersion.isLoading ? (
              <Loading />
            ) : (
              <>
                <h2 className="text-xl font-bold mb-4">{previewTitle}</h2>
                {config.renderBody(previewBody)}
              </>
            )}
          </div>
        </div>

        {/* Metadata panel */}
        <div className="w-full lg:w-72 flex-shrink-0">
          <div className="card">
            <div className="flex flex-col gap-4">
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

            {/* Version history */}
            {versions.data && versions.data.length > 1 ? (() => {
              const reversed = [...versions.data].reverse()
              const totalPages = Math.ceil(reversed.length / versionsPerPage)
              const pageVersions = reversed.slice(
                versionPage * versionsPerPage,
                (versionPage + 1) * versionsPerPage,
              )
              return (
                <div className="mt-4 pt-4 border-t border-slate-100">
                  <span className="text-sm font-semibold text-slate-500 uppercase tracking-wide mb-2 block">
                    Version History
                  </span>
                  <ul className="flex flex-col gap-1">
                    {pageVersions.map(v => {
                      const isSelected = v.id === selectedVersionId
                      const isCurrent = v.id === versions.data[versions.data.length - 1].id && !isViewingVersion
                      return (
                        <li key={v.id}>
                          <button
                            type="button"
                            className={`w-full text-left px-2 py-1.5 text-sm ${
                              isSelected || (isCurrent && !isViewingVersion)
                                ? 'bg-indigo-50 text-indigo-700 font-medium'
                                : 'text-slate-700 hover:bg-slate-50'
                            }`}
                            onClick={() =>
                              setSelectedVersionId(
                                isCurrent && !isViewingVersion ? null : v.id,
                              )
                            }
                          >
                            <span>v{v.version}</span>
                            <span className="text-xs text-slate-400 ml-2">
                              {DateTime.fromISO(v.created_at).toLocaleString(DateTime.DATE_MED)}
                            </span>
                          </button>
                        </li>
                      )
                    })}
                  </ul>
                  {totalPages > 1 ? (
                    <div className="flex items-center justify-between mt-2 pt-2 border-t border-slate-100">
                      <button
                        type="button"
                        className="text-xs text-slate-500 hover:text-slate-700 disabled:opacity-30 disabled:cursor-default"
                        disabled={versionPage === 0}
                        onClick={() => setVersionPage(p => p - 1)}
                      >
                        Newer
                      </button>
                      <span className="text-xs text-slate-400">
                        {versionPage + 1} / {totalPages}
                      </span>
                      <button
                        type="button"
                        className="text-xs text-slate-500 hover:text-slate-700 disabled:opacity-30 disabled:cursor-default"
                        disabled={versionPage >= totalPages - 1}
                        onClick={() => setVersionPage(p => p + 1)}
                      >
                        Older
                      </button>
                    </div>
                  ) : null}
                </div>
              )
            })() : null}
          </div>
        </div>
      </div>

    </>
  )
}
