import { createFileRoute } from '@tanstack/react-router'
import { Settings } from '../components/Settings'
import { accountSearch } from '../lib/account'

export const Route = createFileRoute('/')({ validateSearch: accountSearch, component: Settings })
