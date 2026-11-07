import * as stylex from '@stylexjs/stylex'
import { Flash, iconSizes } from 'paper-ui'
import { ExclamationCircleIcon } from 'paper-ui/icons'
import type { Problem } from '../lib/identity'
import { formStyles, textLink } from './styles'

export function FormProblem({ problem, reauthenticate }: { readonly problem?: Problem; readonly reauthenticate?: string }) {
  if (!problem) return null
  return (
    <Flash variant="danger" icon={<ExclamationCircleIcon {...stylex.props(iconSizes.default)} />}>
      {problem.message}
      {problem.reauthenticate && reauthenticate ? <> <a {...stylex.props(textLink, formStyles.noticeLink)} href={reauthenticate}>Log in again</a></> : null}
    </Flash>
  )
}

export function ReturnTo({ value }: { readonly value?: string }) {
  return value ? <input type="hidden" name="return_to" value={value} /> : null
}
