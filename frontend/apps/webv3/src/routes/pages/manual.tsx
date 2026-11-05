import { createFileRoute } from '@tanstack/react-router'
import { ManualPage } from '../../components/ManualPage'

export const Route = createFileRoute('/pages/manual')({
  head: () => ({ meta: [{ title: 'Manual · Tadoku' }] }),
  component: ManualPage,
})
