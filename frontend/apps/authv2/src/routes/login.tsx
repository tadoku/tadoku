import { createFileRoute } from '@tanstack/react-router'
import { CredentialsForm } from '../components/AccountForms'
import { accountSearch } from '../lib/account'

export const Route = createFileRoute('/login')({ validateSearch: accountSearch, component: () => <CredentialsForm mode="login" /> })
