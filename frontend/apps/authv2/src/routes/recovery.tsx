import { createFileRoute } from '@tanstack/react-router'
import { RecoveryForm } from '../components/AccountForms'
import { accountSearch } from '../lib/account'

export const Route = createFileRoute('/recovery')({ validateSearch: accountSearch, component: RecoveryForm })
