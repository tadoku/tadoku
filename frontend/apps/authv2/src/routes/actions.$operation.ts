import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/actions/$operation')({
  server: {
    handlers: {
      POST: async ({ request, params }) => {
        const { handleAccountAction } = await import('../lib/actions.server')
        return handleAccountAction(request, params.operation)
      },
    },
  },
})
