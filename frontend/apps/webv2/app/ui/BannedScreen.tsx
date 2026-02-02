import { routes } from '@app/common/routes'
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
          please{' '}
          <a href={routes.discord()} target="_blank" rel="noopener noreferrer">
            contact us on Discord
          </a>
          .
        </p>
        <button onClick={onLogout} className="btn primary">
          Log out
        </button>
      </div>
    </div>
  )
}
