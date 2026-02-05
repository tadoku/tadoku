import { Loading, Pagination } from 'ui'
import { ContentConfig, ContentItem } from './types'
import { useContentList } from './api'
import { DateTime } from 'luxon'
import Link from 'next/link'
import { useState, useEffect } from 'react'

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
  namespace: string
}

export function ContentList({ config, namespace }: Props) {
  const [page, setPage] = useState(0)
  const pageSize = 20

  // Reset to first page when namespace changes
  useEffect(() => {
    setPage(0)
  }, [namespace])

  const list = useContentList(config, namespace, { pageSize, page })

  return (
    <>

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
                  <th className="default w-24">Status</th>
                  <th className="default w-32">Updated</th>
                </tr>
              </thead>
              <tbody>
                {list.data.items.map(item => (
                  <tr key={item.id} className="link">
                    <td className="link">
                      <Link href={config.routes.preview(item.id)}>
                        {item.title}
                      </Link>
                    </td>
                    <td className="link">
                      <Link href={config.routes.preview(item.id)}>
                        <StatusBadge item={item} />
                      </Link>
                    </td>
                    <td className="link">
                      <Link href={config.routes.preview(item.id)}>
                        {item.updated_at
                          ? DateTime.fromISO(item.updated_at).toLocaleString(
                              DateTime.DATE_MED,
                            )
                          : 'N/A'}
                      </Link>
                    </td>
                  </tr>
                ))}
                {list.data.items.length === 0 ? (
                  <tr>
                    <td
                      colSpan={3}
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
