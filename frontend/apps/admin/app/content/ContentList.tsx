import { Loading, Pagination } from 'ui'
import { ContentConfig, ContentItem } from './types'
import { useContentList } from './api'
import { DateTime } from 'luxon'
import Link from 'next/link'
import { useState } from 'react'
import { NamespaceSelector, useNamespace } from './NamespaceSelector'

function StatusBadge({ item }: { item: ContentItem }) {
  if (!item.published_at) {
    return (
      <span className="tag bg-amber-100 text-amber-800 justify-center w-16">
        Draft
      </span>
    )
  }
  const publishedAt = DateTime.fromISO(item.published_at)
  if (publishedAt > DateTime.now()) {
    return (
      <span className="tag bg-blue-100 text-blue-800 justify-center w-20">
        Scheduled
      </span>
    )
  }
  return (
    <span className="tag bg-emerald-100 text-emerald-800 justify-center w-20">
      Published
    </span>
  )
}

interface Props {
  config: ContentConfig
}

export function ContentList({ config }: Props) {
  const [namespace, setNamespace] = useNamespace()
  const [page, setPage] = useState(0)
  const pageSize = 20

  const list = useContentList(config, namespace, { pageSize, page })

  return (
    <>
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <NamespaceSelector value={namespace} onChange={ns => { setNamespace(ns); setPage(0) }} />
        <Link href={config.routes.new()} className="btn primary w-fit">
          New {config.label}
        </Link>
      </div>

      {list.isError ? (
        <div className="mt-4">
          {list.error instanceof Error && list.error.message === '403' ? (
            <span className="flash error">
              You do not have permission to view this page.
            </span>
          ) : list.error instanceof Error && list.error.message === '401' ? (
            <span className="flash error">
              You must be logged in to view this page.
            </span>
          ) : (
            <span className="flash error">Could not load {config.labelPlural.toLowerCase()}.</span>
          )}
        </div>
      ) : null}

      {list.isLoading ? (
        <div className="mt-4">
          <Loading />
        </div>
      ) : null}

      {list.isSuccess ? (
        <div className="mt-4">
          <div className="table-container">
            <table className="default">
              <thead>
                <tr>
                  <th className="default">Title</th>
                  <th className="default w-48">Slug</th>
                  <th className="default w-24">Status</th>
                  <th className="default w-40">Updated</th>
                </tr>
              </thead>
              <tbody>
                {list.data.items.map(item => (
                  <tr key={item.id} className="link">
                    <td className="link">
                      <Link href={config.routes.preview(item.slug)}>
                        {item.title}
                      </Link>
                    </td>
                    <td className="link">
                      <Link href={config.routes.preview(item.slug)}>
                        <code className="text-sm text-slate-500">{item.slug}</code>
                      </Link>
                    </td>
                    <td className="link">
                      <Link href={config.routes.preview(item.slug)}>
                        <StatusBadge item={item} />
                      </Link>
                    </td>
                    <td className="link">
                      <Link href={config.routes.preview(item.slug)}>
                        {item.updated_at
                          ? DateTime.fromISO(item.updated_at).toLocaleString(
                              DateTime.DATETIME_SHORT,
                            )
                          : 'N/A'}
                      </Link>
                    </td>
                  </tr>
                ))}
                {list.data.items.length === 0 ? (
                  <tr>
                    <td
                      colSpan={4}
                      className="default h-32 font-bold text-center text-xl text-slate-400"
                    >
                      No {config.labelPlural.toLowerCase()} found
                    </td>
                  </tr>
                ) : null}
              </tbody>
            </table>
          </div>

          {list.data.total_size / pageSize > 1 ? (
            <div className="mt-4">
              <Pagination
                currentPage={page + 1}
                totalPages={Math.ceil(list.data.total_size / pageSize)}
                onClick={(p: number) => setPage(p - 1)}
              />
            </div>
          ) : null}
        </div>
      ) : null}
    </>
  )
}
