import { createFileRoute } from '@tanstack/react-router'
import { Settings } from '../components/Settings'
import { accountSearch } from '../lib/account'

export const Route = createFileRoute('/settings')({ validateSearch: accountSearch, component: Settings })
