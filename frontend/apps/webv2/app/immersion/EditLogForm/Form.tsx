import { Input, TagsInput } from 'ui'
import { TagsSidebar } from '@app/immersion/components/TagsSidebar'
import { FormProvider, useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import {
  fetchTagSuggestions,
  Log,
  LogConfigurationOptions,
  UpdateLogPayload,
  useUpdateLog,
} from '@app/immersion/api'
import { useRouter } from 'next/router'
import { routes } from '@app/common/routes'
import {
  estimateScore,
  filterUnits,
  NewLogFormV2Schema,
} from '@app/immersion/NewLogFormV2/domain'
import { formatScore } from '@app/common/format'
import { useDebouncedCallback } from 'use-debounce'
import { useSessionOrRedirect } from '@app/common/session'
import { AmountWithUnit, Option } from 'ui/components/Form'
import { toast } from 'react-toastify'

interface Props {
  options: LogConfigurationOptions
  log: Log
}

export const EditLogForm = ({ options, log }: Props) => {
  const defaultValues: Partial<NewLogFormV2Schema> = {
    languageCode: log.language.code,
    activityId: log.activity.id,
    amountValue: log.amount,
    amountUnit: log.unit_id,
    tags: log.tags,
    description: log.description ?? '',
    allUnits: options.units,
  }

  const methods = useForm({
    resolver: zodResolver(NewLogFormV2Schema),
    defaultValues,
  })

  useSessionOrRedirect()

  const unitId = methods.watch('amountUnit')
  const amount = methods.watch('amountValue')

  const units = filterUnits(options.units, log.activity.id, log.language.code)
  const unitsAsOptions: Option[] = units.map(it => ({
    value: it.id,
    label: it.name,
  }))
  const currentSelectedUnit = units.find(it => it.id === unitId)
  const estimatedScore = estimateScore(amount, currentSelectedUnit)

  const router = useRouter()
  const updateLogMutation = useUpdateLog(updatedLog => {
    toast.success('Log updated')
    router.replace(routes.log(updatedLog.id))
  })

  const updateLog = useDebouncedCallback(updateLogMutation.mutate, 2500, {
    leading: true,
    trailing: false,
  })

  const onSubmit = (data: any) => {
    const payload: UpdateLogPayload = {
      amount: data.amountValue,
      unit_id: data.amountUnit,
      tags: data.tags,
      description: data.description || undefined,
    }
    updateLog({ logId: log.id, payload })
  }

  return (
    <FormProvider {...methods}>
      <div className="flex flex-col lg:flex-row lg:gap-6">
        <form
          onSubmit={methods.handleSubmit(onSubmit, errors => console.log(errors))}
          className="v-stack spaced max-w-lg flex-1"
        >
          <div className="card">
            <div className="v-stack spaced">
              <label className="label">
                <span className="label-text">Language</span>
                <input
                  type="text"
                  value={log.language.name}
                  disabled
                />
              </label>
              <label className="label">
                <span className="label-text">Activity</span>
                <input
                  type="text"
                  value={log.activity.name}
                  disabled
                />
              </label>
              <AmountWithUnit
                label="Amount"
                name="amount"
                defaultValue={log.amount}
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
                renderSuggestion={s =>
                  s.count > 0 ? `${s.tag} (${s.count}x)` : s.tag
                }
                getValue={s => s.tag}
              />
              <div className="lg:hidden">
                <TagsSidebar activityId={log.activity.id} />
              </div>
            </div>
            <div className="-mx-4 -mb-4 mt-4 px-4 py-2 md:-mx-7 md:-mb-7 md:px-7 md:py-2 bg-slate-500/5 text-center lg:text-right font-mono">
              Estimated score: <strong>{formatScore(estimatedScore)}</strong>
            </div>
          </div>
          <div className="h-stack spaced justify-end">
            <a href={routes.log(log.id)} className="btn ghost">
              Cancel
            </a>
            <button
              type="submit"
              className="btn primary"
              disabled={methods.formState.isSubmitting}
            >
              Save
            </button>
          </div>
        </form>
        <aside className="hidden lg:block lg:w-56 lg:pt-1">
          <div className="sticky top-4">
            <TagsSidebar activityId={log.activity.id} />
          </div>
        </aside>
      </div>
    </FormProvider>
  )
}
