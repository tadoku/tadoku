import { createFileRoute } from '@tanstack/react-router'
import { CredentialsForm } from '../components/AccountForms'
import { accountSearch } from '../lib/account'

export const Route = createFileRoute('/registration')({ validateSearch: accountSearch, component: () => <CredentialsForm mode="registration" /> })
