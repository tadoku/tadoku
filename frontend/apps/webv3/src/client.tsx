import { StrictMode, startTransition } from 'react'
import { hydrateRoot } from 'react-dom/client'
import { StartClient } from '@tanstack/react-start/client'
import { listenForClientErrors, reportClientError } from './lib/client-errors'

listenForClientErrors()

// TanStack Start's default client entry, plus reporting the errors React recovers from, such as a hydration mismatch.
startTransition(() => {
  hydrateRoot(
    document,
    <StrictMode>
      <StartClient />
    </StrictMode>,
    {
      onRecoverableError: error => {
        reportClientError('hydration', error)
        console.error(error)
      },
    },
  )
})
