import { createFileRoute } from '@tanstack/react-router'
import { Settings } from '../components/Settings'
import { settingsRoute } from '../lib/settings-route'

export const Route = createFileRoute('/')({
  ...settingsRoute,
  component: RouteComponent,
})

function RouteComponent() {
  return <Settings search={Route.useSearch()} />
}
