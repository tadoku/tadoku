import { FormProvider, useForm } from 'react-hook-form'
import { Input, Modal, TextArea } from 'ui'
import {
  AccountDeletionRefusal,
  UserListEntry,
  useFileAccountDeletion,
} from '@app/common/api'

const refusalMessages: Record<AccountDeletionRefusal, string> = {
  confirmation_mismatch:
    'That address does not match the account’s email address. Nothing was changed.',
  admin_account:
    'This account is an administrator. Remove the admin role first.',
  account_deletion_in_progress:
    'A deletion for this account is already filed. Nothing was changed.',
  not_found: 'This account no longer exists. It may already be deleted.',
  rejected:
    'The request was rejected. Check the email address and that the reason has 1–1,000 characters.',
  unauthorized: 'Your session has expired. Log in again and retry.',
  forbidden: 'You no longer have permission to delete accounts.',
  unavailable: 'Could not file the deletion. Try again.',
}

const normalizeEmail = (email: string) => email.trim().toLowerCase()
const reasonLength = (reason: string) => Array.from(reason.trim()).length

interface Props {
  user: UserListEntry
  onClose: () => void
}

export function DeleteAccountDialog({ user, onClose }: Props) {
  const methods = useForm({
    mode: 'onChange',
    defaultValues: { email: '', reason: '' },
  })
  const fileDeletion = useFileAccountDeletion()
  const [email, reason] = methods.watch(['email', 'reason'])
  const pending = fileDeletion.isLoading
  const outcome = fileDeletion.data
  const ready =
    normalizeEmail(email) === normalizeEmail(user.email) &&
    reasonLength(reason) >= 1 &&
    reasonLength(reason) <= 1000

  const close = () => {
    if (!pending) onClose()
  }

  const submit = methods.handleSubmit(values =>
    fileDeletion.mutate({ userId: user.id, ...values }),
  )

  return (
    <Modal isOpen setIsOpen={close} title="Delete account">
      {outcome?.kind === 'filed' ? (
        <>
          <div role="status" className="flash success">
            Deletion filed. The account is locked, and job {outcome.jobId}{' '}
            deletes its data in the background.
          </div>
          <div className="modal-actions">
            <button
              type="button"
              className="btn primary"
              onClick={onClose}
              autoFocus
            >
              Close
            </button>
          </div>
        </>
      ) : (
        <FormProvider {...methods}>
          <div className="modal-body space-y-2">
            <p>
              Delete the account of{' '}
              <strong>{user.display_name || 'this user'}</strong>? This cannot
              be undone. The account is locked at once and its data is deleted
              in the background.
            </p>
            <ul className="list-disc pl-5">
              <li>
                Deleted: sign-in, sessions, the email address, logs outside
                finished contests and registrations in unfinished contests.
              </li>
              <li>
                Kept: logs in finished contests, anonymised as “Deleted
                participant” without descriptions or tags, and contests the
                user owns, shown under “Deleted organizer”.
              </li>
            </ul>
          </div>
          <fieldset disabled={pending} className="modal-body mt-4 space-y-3">
            <Input
              name="email"
              label="Type the account’s email address to confirm"
              type="text"
              autoComplete="off"
              autoCapitalize="off"
              spellCheck={false}
            />
            <TextArea
              name="reason"
              label="Reason"
              hint="Kept in the audit log. No email address or display name."
              rows={3}
              options={{
                validate: value =>
                  reasonLength(value) <= 1000 ||
                  'The reason can be at most 1,000 characters.',
              }}
            />
          </fieldset>
          {outcome?.kind === 'refused' ? (
            <div role="alert" className="flash error mt-4">
              {refusalMessages[outcome.reason]}
            </div>
          ) : null}
          <div className="modal-actions">
            <button
              type="button"
              className="btn danger"
              onClick={submit}
              disabled={!ready || pending}
            >
              {pending ? 'Filing deletion…' : 'Delete account'}
            </button>
            <button
              type="button"
              className="btn ghost"
              onClick={close}
              disabled={pending}
            >
              Cancel
            </button>
          </div>
        </FormProvider>
      )}
    </Modal>
  )
}
