import React, { Dispatch, SetStateAction, useEffect } from 'react'
import { FormProvider, useForm } from 'react-hook-form'
import { Modal, Select } from 'ui'
import { toast } from 'react-toastify'
import { UserListEntry } from '@app/common/api'
import { FeatureFlagKey } from './contracts'
import { useFeatureAccess, useUpdateFeatureAccess } from './api'

const flagOptions = [
  { value: 'release-log-entry-v2', label: 'Release log entry v2' },
]

interface FormValues {
  flagKey: FeatureFlagKey
}

interface Props {
  isOpen: boolean
  setIsOpen: Dispatch<SetStateAction<boolean>>
  user: UserListEntry | null
}

export const FeatureAccessModal = ({ isOpen, setIsOpen, user }: Props) => {
  const methods = useForm<FormValues>({
    defaultValues: { flagKey: 'release-log-entry-v2' },
  })
  const flagKey = methods.watch('flagKey')
  const access = useFeatureAccess(
    { targetUserId: user?.id ?? '', flagKey },
    { enabled: isOpen && Boolean(user) },
  )
  const update = useUpdateFeatureAccess()

  useEffect(() => {
    if (isOpen) methods.reset({ flagKey: 'release-log-entry-v2' })
  }, [isOpen, methods])

  const submit = methods.handleSubmit(async values => {
    if (!user || !access.data) return
    try {
      const result = await update.mutateAsync({
        targetUserId: user.id,
        flagKey: values.flagKey,
        enabled: !access.data.enabled,
      })
      toast.success(
        result.enabled ? 'Feature access granted' : 'Feature access removed',
      )
    } catch {
      toast.error('Could not update feature access')
    }
  })

  const close = () => {
    if (update.isLoading) return
    setIsOpen(false)
  }

  return (
    <Modal isOpen={isOpen} setIsOpen={setIsOpen} title="Feature access">
      <FormProvider {...methods}>
        <form onSubmit={submit}>
          <div className="modal-body flex flex-col gap-4">
            <div>
              <p className="font-medium">
                {user?.display_name || 'Unnamed user'}
              </p>
              <p className="text-sm text-slate-500">{user?.email}</p>
            </div>

            <Select<FormValues>
              name="flagKey"
              label="Feature"
              values={flagOptions}
            />

            {access.isLoading ? (
              <p className="text-sm text-slate-500">Checking access…</p>
            ) : access.isError ? (
              <span className="flash error">
                Could not load feature access.
              </span>
            ) : access.data ? (
              <div className="flash info">
                <p>
                  Access is currently{' '}
                  <strong>
                    {access.data.enabled ? 'granted' : 'not granted'}
                  </strong>{' '}
                  in {access.data.environment}.
                </p>
                <p className="mt-1 text-xs break-all">
                  Git revision: {access.data.revision}
                </p>
              </div>
            ) : null}
          </div>

          <div className="modal-actions">
            <button
              type="submit"
              className={access.data?.enabled ? 'btn danger' : 'btn primary'}
              disabled={!access.data || access.isError || update.isLoading}
            >
              {update.isLoading
                ? 'Saving…'
                : access.data?.enabled
                ? 'Remove access'
                : 'Grant access'}
            </button>
            <button
              type="button"
              className="btn ghost"
              onClick={close}
              disabled={update.isLoading}
            >
              Close
            </button>
          </div>
        </form>
      </FormProvider>
    </Modal>
  )
}
