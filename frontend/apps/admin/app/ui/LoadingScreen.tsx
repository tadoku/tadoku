import { Loading } from 'ui'

export default function LoadingScreen() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-slate-50">
      <div className="text-center">
        <Loading />
        <p className="mt-4 text-slate-600">Loading...</p>
      </div>
    </div>
  )
}
