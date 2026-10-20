import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/actions/$operation')({
  server: {
    handlers: {
      POST: async ({ request, params }) => {
        const { handleNativeAction } = await import('../lib/actions.server')
        return handleNativeAction(request, params.operation)
      },
    },
  },
})
