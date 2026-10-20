import { Flash } from 'paper-ui'
import { ExclamationCircleIcon, iconClassName } from 'paper-ui/icons'
import type { Problem } from '../lib/identity'

export function FormProblem({ problem, reauthenticate }: { readonly problem?: Problem; readonly reauthenticate?: string }) {
  if (!problem) return null
  return (
    <Flash variant="danger" icon={<ExclamationCircleIcon className={iconClassName()} />}>
      {problem.message}
      {problem.reauthenticate && reauthenticate ? <> <a href={reauthenticate}>Log in again</a></> : null}
    </Flash>
  )
}

export function ReturnTo({ value }: { readonly value?: string }) {
  return value ? <input type="hidden" name="return_to" value={value} /> : null
}
