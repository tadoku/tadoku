import { ShieldExclamationIcon } from '@heroicons/react/24/outline'
import { routes } from '@app/common/routes'

export default function AccessDenied() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-slate-50">
      <div className="text-center">
        <ShieldExclamationIcon className="mx-auto h-16 w-16 text-red-500" />
        <h1 className="mt-4 text-3xl font-bold text-slate-900">Access Denied</h1>
        <p className="mt-2 text-slate-600">
          You do not have permission to access this area.
        </p>
        <p className="mt-1 text-slate-500 text-sm">
          This section is restricted to administrators only.
        </p>
        <div className="mt-6">
          <a href={routes.mainApp()} className="btn primary">
            Return to Main Site
          </a>
        </div>
      </div>
    </div>
  )
}
