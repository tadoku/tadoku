import { createFileRoute } from '@tanstack/react-router'
import { GoalPage } from '../components/GoalPage'

export const Route = createFileRoute('/plan')({
  head: () => ({
    meta: [
      { title: 'Set a daily goal · Tadoku' },
      {
        name: 'description',
        content: 'Decide how long you want to read or listen each day. Your goal stays in this browser.',
      },
    ],
  }),
  component: GoalPage,
})
