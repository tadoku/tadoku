import { routes } from '@app/common/routes'
import {
  UsersIcon,
  EllipsisVerticalIcon,
  NoSymbolIcon,
  CheckCircleIcon,
  HomeIcon,
} from '@heroicons/react/20/solid'
import Head from 'next/head'
import { ActionMenu, Breadcrumb, Loading, Modal, Pagination, TextArea } from 'ui'
import { NextPageWithLayout } from './_app'
import { getDashboardLayout } from '@app/ui/DashboardLayout'
import { useUserList, useUpdateUserRole, UserListEntry } from '@app/common/api'
import { useEffect, useState } from 'react'
import { DateTime } from 'luxon'
import { useQueryClient } from 'react-query'
import { toast } from 'react-toastify'
import { FormProvider, useForm } from 'react-hook-form'

function RoleBadge({ role }: { role?: string }) {
  if (role === 'admin') {
    return (
      <span className="tag bg-purple-100 text-purple-800 justify-center w-16">Admin</span>
    )
  }
  if (role === 'banned') {
    return (
      <span className="tag bg-red-100 text-red-800 justify-center w-16">Banned</span>
    )
  }
  return (
    <span className="tag bg-slate-100 text-slate-600 justify-center w-16">User</span>
  )
}

const Page: NextPageWithLayout = () => {
  const [page, setPage] = useState(0)
  const [search, setSearch] = useState('')
  const [modalOpen, setModalOpen] = useState(false)
  const [selectedUser, setSelectedUser] = useState<UserListEntry | null>(null)
  const pageSize = 20
  const queryClient = useQueryClient()

  const searchMethods = useForm({
    defaultValues: { query: '' },
  })

  const modalMethods = useForm({
    defaultValues: { reason: '' },
  })

  useEffect(() => {
    if (modalOpen) {
      modalMethods.reset({ reason: '' })
    }
  }, [modalOpen, modalMethods])

  const users = useUserList(
    {
      pageSize,
      page,
      query: search || undefined,
    },
    { enabled: true },
  )

  const updateRoleMutation = useUpdateUserRole(
    () => {
      toast.success(
        selectedUser?.role === 'banned'
          ? 'User has been unbanned'
          : 'User has been banned',
      )
      queryClient.invalidateQueries(['users', 'list'])
      setModalOpen(false)
      modalMethods.reset({ reason: '' })
      setSelectedUser(null)
    },
    () => {
      toast.error('Failed to update user role')
    },
  )

  const handleSearch = searchMethods.handleSubmit(data => {
    setSearch(data.query)
    setPage(0)
  })

  const handleClearSearch = () => {
    searchMethods.reset({ query: '' })
    setSearch('')
    setPage(0)
  }

  const handleBanUnban = modalMethods.handleSubmit(data => {
    if (!selectedUser) return
    const newRole = selectedUser.role === 'banned' ? 'user' : 'banned'
    updateRoleMutation.mutate({
      userId: selectedUser.id,
      role: newRole,
      reason: data.reason,
    })
  })

  const isBanning = selectedUser?.role !== 'banned'

  return (
    <>
      <Head>
        <title>Users - Admin - Tadoku</title>
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
              label: 'Users',
              href: routes.users(),
              IconComponent: UsersIcon,
            },
          ]}
        />
      </div>
      <h1 className="title">Users</h1>

      <FormProvider {...searchMethods}>
        <form onSubmit={handleSearch} className="mt-4 flex gap-2">
          <input
            type="text"
            className="input flex-1"
            placeholder="Search by name or email..."
            {...searchMethods.register('query')}
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
      </FormProvider>

      {users.isError ? (
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
      ) : null}

      {users.isLoading ? (
        <div className="mt-4">
          <Loading />
        </div>
      ) : null}

      {users.isSuccess ? (
        <div className="mt-4">
          <div className="table-container">
            <table className="default">
              <thead>
                <tr>
                  <th className="default">Display Name</th>
                  <th className="default">Email</th>
                  <th className="default w-24">Role</th>
                  <th className="default w-40">Created At</th>
                  <th className="default w-12"></th>
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
                      <RoleBadge role={user.role} />
                    </td>
                    <td className="default">
                      {user.created_at
                        ? DateTime.fromISO(user.created_at).toLocaleString(
                            DateTime.DATE_MED,
                          )
                        : 'N/A'}
                    </td>
                    <td className="default">
                      {user.role !== 'admin' ? (
                        <ActionMenu
                          orientation="right"
                          links={[
                            user.role === 'banned'
                              ? {
                                  label: 'Unban User',
                                  href: '#',
                                  IconComponent: CheckCircleIcon,
                                  onClick: () => {
                                    setSelectedUser(user)
                                    setModalOpen(true)
                                  },
                                }
                              : {
                                  label: 'Ban User',
                                  href: '#',
                                  IconComponent: NoSymbolIcon,
                                  type: 'danger' as const,
                                  onClick: () => {
                                    setSelectedUser(user)
                                    setModalOpen(true)
                                  },
                                },
                          ]}
                        >
                          <EllipsisVerticalIcon className="w-5 h-5" />
                        </ActionMenu>
                      ) : null}
                    </td>
                  </tr>
                ))}
                {users.data.users.length === 0 ? (
                  <tr>
                    <td
                      colSpan={5}
                      className="default h-32 font-bold text-center text-xl text-slate-400"
                    >
                      {search ? 'No users found matching your search' : 'No users found'}
                    </td>
                  </tr>
                ) : null}
              </tbody>
            </table>
          </div>

          {users.data.total_size / pageSize > 1 ? (
            <div className="mt-4">
              <Pagination
                currentPage={page + 1}
                totalPages={Math.ceil(users.data.total_size / pageSize)}
                onClick={(p: number) => setPage(p - 1)}
              />
            </div>
          ) : null}
        </div>
      ) : null}

      <Modal
        isOpen={modalOpen}
        setIsOpen={setModalOpen}
        title={isBanning ? 'Ban User' : 'Unban User'}
      >
        <FormProvider {...modalMethods}>
          <p className="modal-body">
            {isBanning
              ? `Are you sure you want to ban ${selectedUser?.display_name || 'this user'}? They will no longer be able to access the site.`
              : `Are you sure you want to unban ${selectedUser?.display_name || 'this user'}? They will regain access to the site.`}
          </p>
          <div className="modal-body">
            <TextArea
              name="reason"
              label="Reason"
              placeholder={
                isBanning
                  ? 'e.g. Violated community guidelines...'
                  : 'e.g. Appeal accepted, warning issued...'
              }
              rows={4}
              options={{ required: 'Reason is required' }}
            />
          </div>
          <div className="modal-actions">
            <button
              type="button"
              className={isBanning ? 'btn danger' : 'btn primary'}
              onClick={handleBanUnban}
              disabled={updateRoleMutation.isLoading}
            >
              {updateRoleMutation.isLoading
                ? isBanning
                  ? 'Banning...'
                  : 'Unbanning...'
                : isBanning
                  ? 'Yes, ban user'
                  : 'Yes, unban user'}
            </button>
            <button
              type="button"
              className="btn ghost"
              onClick={() => {
                setModalOpen(false)
                setSelectedUser(null)
              }}
            >
              Cancel
            </button>
          </div>
        </FormProvider>
      </Modal>
    </>
  )
}

Page.getLayout = getDashboardLayout('users')

export default Page
