import { useLogoutHandler, useSession } from '@app/common/session'

export default function BannedScreen() {
  const [session] = useSession()
  const onLogout = useLogoutHandler([session])

  return (
    <div className="min-h-screen flex items-center justify-center bg-slate-50">
      <div className="text-center p-8">
        <h1 className="text-2xl font-bold text-slate-900 mb-4">
          Account Suspended
        </h1>
        <p className="text-slate-600 mb-6">
          Your account has been suspended. If you believe this is an error,
          please contact support.
        </p>
        <button
          onClick={onLogout}
          className="px-4 py-2 bg-slate-900 text-white rounded-md hover:bg-slate-800 transition-colors"
        >
          Log out
        </button>
      </div>
    </div>
  )
}
