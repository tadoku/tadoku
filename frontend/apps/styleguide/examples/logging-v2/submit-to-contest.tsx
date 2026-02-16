import { CheckIcon } from '@heroicons/react/20/solid'
import { FormProvider, useForm } from 'react-hook-form'

interface Contest {
  id: string
  name: string
  admin: string
  score: number
  eligible: boolean
  reason?: string
}

const contests: Contest[] = [
  {
    id: 'round4',
    name: 'Round 4 2025',
    admin: 'Tadoku',
    score: 230,
    eligible: true,
  },
  {
    id: 'deepweeb',
    name: "KanjiEater's Deep Weeb Club",
    admin: 'KanjiEater',
    score: 460,
    eligible: true,
  },
  {
    id: 'listening',
    name: '2x listening club',
    admin: 'antonve',
    score: 0,
    eligible: false,
    reason: 'Log is not eligible',
  },
]

export default function SubmitToContest() {
  const methods = useForm({
    defaultValues: {
      round4: true,
      deepweeb: true,
      listening: false,
    },
  })
  const onSubmit = (data: unknown) => console.log(data, 'submitted')
  const watched = methods.watch()

  return (
    <FormProvider {...methods}>
      <form onSubmit={methods.handleSubmit(onSubmit)}>
        <div className="bg-neutral-50 px-4 py-3">
          <div className="text-xs font-medium text-neutral-400 mb-1">Your log</div>
          <div className="text-sm flex items-baseline justify-between">
            <span>
              <strong>Japanese</strong> &middot; Reading &middot; 230 pages
            </span>
            <span>無職転生 5巻</span>
          </div>
        </div>
        <div className="my-6" />
        <div className="v-stack gap-2">
          {contests.map(contest => {
            const isChecked = watched[contest.id as keyof typeof watched]

            if (!contest.eligible) {
              return (
                <div
                  key={contest.id}
                  className="input-frame px-4 py-2 pointer-events-none opacity-40"
                >
                  <div className="h-stack items-center w-full">
                    <div className="flex-1">
                      <div className="font-bold text-secondary/30">{contest.name}</div>
                      <div className="text-xs text-secondary/30">{contest.admin}</div>
                    </div>
                    <span className="text-xs text-slate-400 italic mr-2">
                      {contest.reason}
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
                key={contest.id}
                className={`input-frame px-4 py-2 cursor-pointer select-none transition-colors ${
                  isChecked ? '!border-primary bg-primary/5 hover:bg-primary/10' : 'hover:bg-neutral-50'
                }`}
              >
                <div className="h-stack items-center w-full">
                  <input
                    type="checkbox"
                    {...methods.register(contest.id as 'round4' | 'deepweeb')}
                    className="hidden"
                  />
                  <div className="flex-1">
                    <div className="font-bold text-secondary">{contest.name}</div>
                    <div className="text-xs text-gray-600">{contest.admin}</div>
                  </div>
                  <span className="text-sm font-medium text-secondary mr-4">
                    Score: {contest.score}
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
        <div className="flex justify-end mt-6">
          <button
            type="submit"
            className="btn primary"
            disabled={methods.formState.isSubmitting}
          >
            Submit
          </button>
        </div>
      </form>
    </FormProvider>
  )
}
