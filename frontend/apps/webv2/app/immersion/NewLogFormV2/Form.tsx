import { Input, TagsInput } from 'ui'
import { TagsSidebar } from '@app/immersion/components/TagsSidebar'
import { FormProvider, useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import {
  fetchTagSuggestions,
  LogConfigurationOptions,
  useCreateLogV2,
  useOngoingContestRegistrations,
} from '@app/immersion/api'
import { useRouter } from 'next/router'
import { routes } from '@app/common/routes'
import {
  estimateScore,
  filterUnits,
  NewLogFormV2Schema,
  NewLogV2APISchema,
} from '@app/immersion/NewLogFormV2/domain'
import { formatScore } from '@app/common/format'
import { useDebouncedCallback } from 'use-debounce'
import { useSessionOrRedirect } from '@app/common/session'
import { useEffect } from 'react'
import { AmountWithUnit, Option, Select } from 'ui/components/Form'

interface Props {
  options: LogConfigurationOptions
  defaultValues?: Partial<NewLogFormV2Schema>
}

export const LogFormV2 = ({ options, defaultValues: originalDefaultValues }: Props) => {
  const defaultValues: Partial<NewLogFormV2Schema> = {
    ...originalDefaultValues,
    activityId: options.activities[0].id,
    amountUnit: options.units.filter(
      it => it.log_activity_id === options.activities[0].id,
    )[0]?.id,
    allUnits: options.units,
  }

  const methods = useForm({
    resolver: zodResolver(NewLogFormV2Schema),
    defaultValues,
  })

  useSessionOrRedirect()

  const activityId = methods.watch('activityId')
  const languageCode = methods.watch('languageCode')
  const unitId = methods.watch('amountUnit')
  const amount = methods.watch('amountValue')

  const languagesAsOptions: Option[] = options.languages.map(it => ({
    value: it.code,
    label: it.name,
  }))

  const activity = options.activities.find(it => it.id === activityId)
  const units = filterUnits(options.units, activity?.id, languageCode)
  const unitsAsOptions: Option[] = units.map(it => ({
    value: it.id,
    label: it.name,
  }))
  const currentSelectedUnit = units.find(it => it.id === unitId)
  const activitiesAsOptions: Option[] = options.activities.map(it => ({
    value: it.id.toString(),
    label: it.name,
  }))
  const estimatedScore = estimateScore(amount, currentSelectedUnit)

  // Eagerly prefetch ongoing registrations (non-blocking)
  const registrations = useOngoingContestRegistrations()

  const router = useRouter()
  const createLogMutation = useCreateLogV2(log => {
    const hasRegistrations =
      registrations.data &&
      registrations.data.registrations.length > 0
    if (hasRegistrations) {
      router.replace(routes.logContests(log.id) + '?preselect=1')
    } else {
      router.replace(routes.log(log.id))
    }
  })

  const createLog = useDebouncedCallback(createLogMutation.mutate, 2500, {
    leading: true,
    trailing: false,
  })

  const onSubmit = (data: any) => {
    createLog(NewLogV2APISchema.parse(data))
  }

  useEffect(() => {
    const subscription = methods.watch((value, { name, type }) => {
      if (
        (name === 'languageCode' || name === 'activityId') &&
        type === 'change'
      ) {
        const id = filterUnits(
          options.units,
          value.activityId,
          languageCode,
        )?.[0]?.id
        if (id !== methods.getValues('amountUnit')) {
          methods.setValue('amountUnit', id)
        }
      }
    })
    return () => subscription.unsubscribe()
  }, [methods, languageCode, options.units])

  return (
    <FormProvider {...methods}>
      <div className="flex flex-col lg:flex-row lg:gap-6">
        <form
          onSubmit={methods.handleSubmit(onSubmit, errors => console.log(errors))}
          className="v-stack spaced max-w-lg flex-1"
        >
          <div className="card">
            <div className="v-stack spaced">
              <Select
                name="languageCode"
                label="Language"
                values={languagesAsOptions}
              />
              <Select
                name="activityId"
                label="Activity"
                values={activitiesAsOptions}
                options={{ valueAsNumber: true }}
              />
              <AmountWithUnit
                label="Amount"
                name="amount"
                defaultValue={0}
                min={0}
                step="any"
                units={unitsAsOptions}
                unitsLabel="Unit"
              />
              <Input
                name="description"
                label="Description"
                type="text"
                placeholder="e.g. One Piece volume 45"
              />
              <TagsInput
                name="tags"
                label="Tags"
                placeholder="Add tags..."
                maxTags={10}
                getSuggestions={fetchTagSuggestions}
                renderSuggestion={s => (s.count > 0 ? `${s.tag} (${s.count}×)` : s.tag)}
                getValue={s => s.tag}
              />
              <div className="lg:hidden">
                <TagsSidebar activityId={activityId} />
              </div>
            </div>
            <div className="-mx-4 -mb-4 mt-4 px-4 py-2 md:-mx-7 md:-mb-7 md:px-7 md:py-2 bg-slate-500/5 text-center lg:text-right font-mono">
              Estimated score: <strong>{formatScore(estimatedScore)}</strong>
            </div>
          </div>
          <div className="h-stack spaced justify-end">
            <a href={routes.home()} className="btn ghost">
              Cancel
            </a>
            <button
              type="submit"
              className="btn primary"
              disabled={methods.formState.isSubmitting}
            >
              Create
            </button>
          </div>
        </form>
        <aside className="hidden lg:block lg:w-56 lg:pt-1">
          <div className="sticky top-14 sm:top-20">
            <TagsSidebar activityId={activityId} />
          </div>
        </aside>
      </div>
    </FormProvider>
  )
}
