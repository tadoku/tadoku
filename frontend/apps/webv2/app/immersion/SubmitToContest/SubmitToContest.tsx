import { CheckIcon } from '@heroicons/react/20/solid'
import { FormProvider, useForm } from 'react-hook-form'
import {
  ContestRegistrationView,
  Log,
  useUpdateLogContestRegistrations,
} from '@app/immersion/api'
import { useRouter } from 'next/router'
import { routes } from '@app/common/routes'
import { formatScore } from '@app/common/format'
import { formatUnit } from '@app/common/format'
import { classifyRegistrations } from '@app/immersion/SubmitToContest/domain'
import { toast } from 'react-toastify'
import { useDebouncedCallback } from 'use-debounce'

interface Props {
  log: Log
  registrations: ContestRegistrationView[]
}

export const SubmitToContest = ({ log, registrations }: Props) => {
  const options = classifyRegistrations(log, registrations)

  const alreadyAttachedIds = new Set(
    (log.registrations ?? []).map(r => r.registration_id),
  )

  const defaultChecked: Record<string, boolean> = {}
  for (const option of options) {
    if (option.eligible) {
      defaultChecked[option.registration.id] =
        alreadyAttachedIds.has(option.registration.id) ||
        alreadyAttachedIds.size === 0
    }
  }

  const methods = useForm({ defaultValues: defaultChecked })
  const watched = methods.watch()

  const router = useRouter()
  const mutation = useUpdateLogContestRegistrations(updatedLog => {
    toast.success('Contest submissions updated')
    router.push(routes.log(updatedLog.id))
  })

  const submit = useDebouncedCallback(mutation.mutate, 2500, {
    leading: true,
    trailing: false,
  })

  const onSubmit = () => {
    const values = methods.getValues()
    const selectedIds = options
      .filter(o => o.eligible && values[o.registration.id])
      .map(o => o.registration.id)

    submit({ logId: log.id, registrationIds: selectedIds })
  }

  return (
    <FormProvider {...methods}>
      <form onSubmit={methods.handleSubmit(onSubmit)} className="v-stack spaced">
        <div className="card">
          <div className="bg-neutral-50 -mx-4 -mt-4 md:-mx-7 md:-mt-7 px-4 py-3 md:px-7 rounded-t">
            <div className="text-xs font-medium text-neutral-400 mb-1">
              Your log
            </div>
            <div className="text-sm flex items-baseline justify-between">
              <span>
                <strong>{log.language.name}</strong> &middot;{' '}
                {log.activity.name} &middot;{' '}
                {formatScore(log.amount)} {formatUnit(log.amount, log.unit_name)}
              </span>
              {log.description ? <span>{log.description}</span> : null}
            </div>
          </div>
          <div className="mt-6" />
          <div className="v-stack gap-2">
            {options.map(option => {
              const reg = option.registration
              const contest = reg.contest!
              const isChecked = watched[reg.id] ?? false

              if (!option.eligible) {
                return (
                  <div
                    key={reg.id}
                    className="input-frame px-4 py-2 pointer-events-none opacity-40"
                  >
                    <div className="h-stack items-center w-full">
                      <div className="flex-1">
                        <div className="font-bold text-secondary/30">
                          {contest.title}
                        </div>
                        <div className="text-xs text-secondary/30">
                          {contest.owner_user_display_name ?? 'Unknown'}
                        </div>
                      </div>
                      <span className="text-xs text-slate-400 italic mr-2">
                        {option.reason}
                      </span>
                      <span className="flex items-center justify-center border border-black/10 rounded-xl w-4 h-4 text-transparent">
                        <CheckIcon className="w-3 h-3" />
                      </span>
                    </div>
                  </div>
                )
              }

              return (
                <label
                  key={reg.id}
                  className={`input-frame px-4 py-2 cursor-pointer select-none transition-colors ${
                    isChecked
                      ? '!border-primary bg-primary/5 hover:bg-primary/10'
                      : 'hover:bg-neutral-50'
                  }`}
                >
                  <div className="h-stack items-center w-full">
                    <input
                      type="checkbox"
                      {...methods.register(reg.id)}
                      className="hidden"
                    />
                    <div className="flex-1">
                      <div className="font-bold text-secondary">
                        {contest.title}
                      </div>
                      <div className="text-xs text-gray-600">
                        {contest.owner_user_display_name ?? 'Unknown'}
                      </div>
                    </div>
                    <span className="text-sm font-medium text-secondary mr-4">
                      Score: {formatScore(log.score)}
                    </span>
                    <span
                      className={`flex items-center justify-center border rounded-xl w-4 h-4 ${
                        isChecked
                          ? 'bg-primary border-primary text-white'
                          : 'border-black/10 text-transparent'
                      }`}
                    >
                      <CheckIcon className="w-3 h-3" />
                    </span>
                  </div>
                </label>
              )
            })}
          </div>
          {options.length === 0 ? (
            <p className="text-sm text-slate-500">
              No contest registrations found. You can submit to contests after
              joining one.
            </p>
          ) : null}
        </div>
        <div className="h-stack spaced justify-end">
          <a href={routes.log(log.id)} className="btn ghost">
            Skip
          </a>
          <button
            type="submit"
            className="btn primary"
            disabled={mutation.isLoading}
          >
            Submit
          </button>
        </div>
      </form>
    </FormProvider>
  )
}
