import { routes } from '@app/common/routes'
import { MegaphoneIcon, HomeIcon } from '@heroicons/react/20/solid'
import Head from 'next/head'
import { Breadcrumb, Loading, Pagination } from 'ui'
import { NextPageWithLayout } from '../../_app'
import { getDashboardLayout } from '@app/ui/DashboardLayout'
import { NamespaceSelector, useNamespace } from '@app/content/NamespaceSelector'
import { useAnnouncementList } from '@app/announcements/api'
import Link from 'next/link'
import { useRouter } from 'next/router'
import { useState, useEffect } from 'react'
import { DateTime } from 'luxon'

const Page: NextPageWithLayout = () => {
  const router = useRouter()
  const namespace = useNamespace()
  const [page, setPage] = useState(0)
  const pageSize = 20

  useEffect(() => {
    setPage(0)
  }, [namespace])

  const handleNamespaceChange = (ns: string) => {
    router.push(routes.announcements(ns))
  }

  const list = useAnnouncementList(namespace, { pageSize, page })

  const isActive = (startsAt: string, endsAt: string) => {
    const now = new Date()
    return new Date(startsAt) <= now && new Date(endsAt) > now
  }

  return (
    <>
      <Head>
        <title>Announcements - Admin - Tadoku</title>
      </Head>
      <div className="pb-4">
        <Breadcrumb
          links={[
            {
              label: 'Admin',
              href: routes.home(),
              IconComponent: HomeIcon,
            },
            {
              label: 'Announcements',
              href: routes.announcements(namespace),
              IconComponent: MegaphoneIcon,
            },
          ]}
        />
      </div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="title">Announcements</h1>
        <div className="flex items-center gap-2">
          <NamespaceSelector value={namespace} onChange={handleNamespaceChange} />
          <Link href={routes.announcementNew(namespace)} className="btn primary">
            New Announcement
          </Link>
        </div>
      </div>

      {list.isError ? (
        <div className="mt-4">
          {list.error instanceof Error && list.error.message === '403' ? (
            <span className="flash error">
              You do not have permission to view this page.
            </span>
          ) : (
            <span className="flash error">Could not load announcements.</span>
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
                  <th className="default w-40">Starts</th>
                  <th className="default w-40">Ends</th>
                </tr>
              </thead>
              <tbody>
                {list.data.announcements.map(item => (
                  <tr key={item.id} className="link">
                    <td className="link">
                      <Link href={routes.announcementEdit(namespace, item.id)}>
                        {item.title}
                      </Link>
                    </td>
                    <td className="link">
                      <Link href={routes.announcementEdit(namespace, item.id)}>
                        {isActive(item.starts_at, item.ends_at) ? (
                          <span className="label success">Active</span>
                        ) : new Date(item.starts_at) > new Date() ? (
                          <span className="label info">Scheduled</span>
                        ) : (
                          <span className="label">Expired</span>
                        )}
                      </Link>
                    </td>
                    <td className="link">
                      <Link href={routes.announcementEdit(namespace, item.id)} title={DateTime.fromISO(item.starts_at).toLocaleString(DateTime.DATETIME_MED)}>
                        {DateTime.fromISO(item.starts_at).toLocaleString(DateTime.DATE_MED)}
                      </Link>
                    </td>
                    <td className="link">
                      <Link href={routes.announcementEdit(namespace, item.id)} title={DateTime.fromISO(item.ends_at).toLocaleString(DateTime.DATETIME_MED)}>
                        {DateTime.fromISO(item.ends_at).toLocaleString(DateTime.DATE_MED)}
                      </Link>
                    </td>
                  </tr>
                ))}
                {list.data.announcements.length === 0 ? (
                  <tr>
                    <td
                      colSpan={4}
                      className="default h-32 font-bold text-center text-xl text-slate-400"
                    >
                      No announcements found
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

Page.getLayout = getDashboardLayout('announcements')

export default Page
