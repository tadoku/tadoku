import { createFileRoute } from '@tanstack/react-router'
import { AccountSkeleton } from '../components/AccountSkeleton'

export const Route = createFileRoute('/login')({ component: AccountSkeleton })
