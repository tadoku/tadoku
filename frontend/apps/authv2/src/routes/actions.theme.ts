import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/actions/theme')({
  server: {
    handlers: {
      POST: async ({ request }) => {
        const { handleThemeAction } = await import('../lib/actions.server')
        return handleThemeAction(request)
      },
    },
  },
})
