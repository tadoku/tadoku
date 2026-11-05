import { createFileRoute, redirect } from '@tanstack/react-router'

// Webv2's permanent redirect.
export const Route = createFileRoute('/contests/')({
  beforeLoad: () => {
    throw redirect({ href: '/contests/official', statusCode: 308 })
  },
})
