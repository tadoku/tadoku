import { routes } from '@app/common/routes'
import { UsersIcon } from '@heroicons/react/20/solid'
import Head from 'next/head'
import { Breadcrumb, Loading, Pagination } from 'ui'
import { NextPageWithLayout } from '../_app'
import { getAdminLayout } from '@app/manage/AdminLayout'
import { useUserList } from '@app/immersion/api'
import { useState } from 'react'
import { DateTime } from 'luxon'

const Page: NextPageWithLayout = () => {
  const [page, setPage] = useState(0)
  const [search, setSearch] = useState('')
  const [searchInput, setSearchInput] = useState('')
  const pageSize = 20

  const users = useUserList(
    {
      pageSize,
      page,
      query: search || undefined,
    },
    { enabled: true },
  )

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    setSearch(searchInput)
    setPage(0)
  }

  const handleClearSearch = () => {
    setSearchInput('')
    setSearch('')
    setPage(0)
  }

  return (
    <>
      <Head>
        <title>Users - Admin - Tadoku</title>
      </Head>
      <div className="pb-4">
        <Breadcrumb
          links={[
            {
              label: 'Users',
              href: routes.manageUsers(),
              IconComponent: UsersIcon,
            },
          ]}
        />
      </div>
      <h1 className="title">Users</h1>

      <form onSubmit={handleSearch} className="mt-4 flex gap-2">
        <input
          type="text"
          className="input flex-1"
          placeholder="Search by name or email..."
          value={searchInput}
          onChange={e => setSearchInput(e.target.value)}
        />
        <button type="submit" className="btn secondary">
          Search
        </button>
        {search ? (
          <button type="button" className="btn ghost" onClick={handleClearSearch}>
            Clear
          </button>
        ) : null}
      </form>

      {users.isError && (
        <div className="mt-4">
          {users.error instanceof Error && users.error.message === '403' ? (
            <span className="flash error">
              You do not have permission to view this page.
            </span>
          ) : users.error instanceof Error && users.error.message === '401' ? (
            <span className="flash error">
              You must be logged in to view this page.
            </span>
          ) : (
            <span className="flash error">Could not load users.</span>
          )}
        </div>
      )}

      {users.isLoading && (
        <div className="mt-4">
          <Loading />
        </div>
      )}

      {users.isSuccess && (
        <div className="mt-4">
          <div className="table-container shadow-transparent w-auto">
            <table className="default shadow-transparent">
              <thead>
                <tr>
                  <th className="default">Display Name</th>
                  <th className="default">Email</th>
                  <th className="default w-40">Created At</th>
                </tr>
              </thead>
              <tbody>
                {users.data.users.map(user => (
                  <tr key={user.id}>
                    <td className="default font-medium">
                      {user.display_name || 'N/A'}
                    </td>
                    <td className="default">{user.email}</td>
                    <td className="default">
                      {user.created_at
                        ? DateTime.fromISO(user.created_at).toLocaleString(
                            DateTime.DATE_MED,
                          )
                        : 'N/A'}
                    </td>
                  </tr>
                ))}
                {users.data.users.length === 0 && (
                  <tr>
                    <td
                      colSpan={3}
                      className="default h-32 font-bold text-center text-xl text-slate-400"
                    >
                      {search ? 'No users found matching your search' : 'No users found'}
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>

          {users.data.total_size / pageSize > 1 ? (
            <div className="mt-4">
              <Pagination
                currentPage={page + 1}
                totalPages={Math.ceil(users.data.total_size / pageSize)}
                onClick={p => setPage(p - 1)}
              />
            </div>
          ) : null}
        </div>
      )}
    </>
  )
}

Page.getLayout = getAdminLayout('users')

export default Page
