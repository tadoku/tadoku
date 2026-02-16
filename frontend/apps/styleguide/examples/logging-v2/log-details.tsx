import { PencilIcon, TrashIcon } from '@heroicons/react/20/solid'

interface ContestEntry {
  name: string
  admin: string
  score: number
}

const fields = [
  { label: 'Language', value: 'Japanese' },
  { label: 'Activity', value: 'Reading' },
  { label: 'Amount', value: '230 pages' },
  { label: 'Description', value: '無職転生 5巻' },
]

const tags = ['Book', 'Active pick']

const submittedContests: ContestEntry[] = [
  { name: 'Round 4 2025', admin: 'Tadoku', score: 230 },
  { name: "KanjiEater's Deep Weeb Club", admin: 'KanjiEater', score: 460 },
]

export default function LogDetails() {
  return (
    <div>
      <div className="flex items-center justify-between">
        <div>
          <h2 className="font-semibold text-lg">Log details</h2>
          <h2 className="subtitle">By antonve on May 22, 2025</h2>
        </div>
        <div className="h-stack gap-2">
          <button type="button" className="btn ghost gap-2">
            <PencilIcon className="w-4 h-4 mr-2" />
            Edit
          </button>
          <button type="button" className="btn danger gap-2">
            <TrashIcon className="w-4 h-4 mr-2" />
            Delete
          </button>
        </div>
      </div>

      <div className="my-6" />

      <div className="bg-neutral-50 px-4 py-3">
        <div className="v-stack gap-3">
          {fields.map(field => (
            <div key={field.label} className="flex">
              <span className="w-32 text-sm text-neutral-500 flex-shrink-0">
                {field.label}
              </span>
              <span className="text-sm font-medium">{field.value}</span>
            </div>
          ))}
          <div className="flex">
            <span className="w-32 text-sm text-neutral-500 flex-shrink-0">
              Tags
            </span>
            <span className="text-sm font-medium">
              {tags.map(tag => `#${tag.toLowerCase().replace(/\s+/g, '-')}`).join(' ')}
            </span>
          </div>
        </div>
      </div>

      <div className="my-6" />

      <div>
        <h3 className="subtitle mb-2">Submitted to contests</h3>
        <div className="v-stack gap-2">
          {submittedContests.map(contest => (
            <a key={contest.name} href="#" className="input-frame px-4 py-2 no-underline hover:bg-neutral-50 transition-colors">
              <div className="h-stack items-center w-full">
                <div className="flex-1">
                  <div className="font-bold text-secondary">{contest.name}</div>
                  <div className="text-xs text-gray-600">{contest.admin}</div>
                </div>
                <span className="text-sm font-medium text-secondary">
                  Score: {contest.score}
                </span>
              </div>
            </a>
          ))}
        </div>
      </div>
    </div>
  )
}
