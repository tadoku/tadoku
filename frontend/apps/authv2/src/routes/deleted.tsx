import { createFileRoute } from '@tanstack/react-router'
import { Deleted } from '../components/StatusViews'

export const Route = createFileRoute('/deleted')({
  head: () => ({ meta: [{ title: 'Account deleted · Tadoku' }] }),
  component: Deleted,
})
