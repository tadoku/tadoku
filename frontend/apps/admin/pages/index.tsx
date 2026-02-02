import Head from 'next/head'
import { NextPageWithLayout } from './_app'
import { getDashboardLayout } from '@app/ui/DashboardLayout'
import { useUserList } from '@app/common/api'
import { Loading } from 'ui'

const Page: NextPageWithLayout = () => {
  const users = useUserList({ pageSize: 5, page: 0 }, { enabled: true })

  return (
    <>
      <Head>
        <title>Dashboard - Admin - Tadoku</title>
      </Head>
      <h1 className="title">Dashboard</h1>
      <p className="mt-2 text-slate-600">Welcome to the Tadoku Admin Panel.</p>

      <div className="mt-8 grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="bg-slate-50 rounded-lg p-6 border border-slate-200">
          <h2 className="text-lg font-semibold text-slate-900">Users</h2>
          {users.isLoading ? (
            <Loading />
          ) : users.isSuccess ? (
            <p className="mt-2 text-3xl font-bold text-primary">
              {users.data.total_size}
            </p>
          ) : (
            <p className="mt-2 text-slate-500">Unable to load</p>
          )}
          <p className="mt-1 text-sm text-slate-500">Total registered users</p>
        </div>

        <div className="bg-slate-50 rounded-lg p-6 border border-slate-200">
          <h2 className="text-lg font-semibold text-slate-900">Posts</h2>
          <p className="mt-2 text-3xl font-bold text-primary">-</p>
          <p className="mt-1 text-sm text-slate-500">Blog posts</p>
        </div>

        <div className="bg-slate-50 rounded-lg p-6 border border-slate-200">
          <h2 className="text-lg font-semibold text-slate-900">Pages</h2>
          <p className="mt-2 text-3xl font-bold text-primary">-</p>
          <p className="mt-1 text-sm text-slate-500">Static pages</p>
        </div>
      </div>
    </>
  )
}

Page.getLayout = getDashboardLayout()

export default Page
