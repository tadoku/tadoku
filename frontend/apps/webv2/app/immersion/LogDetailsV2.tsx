import { TrashIcon } from '@heroicons/react/20/solid'
import { Log, useDeleteLog } from '@app/immersion/api'
import { routes } from '@app/common/routes'
import { formatScore, formatUnit } from '@app/common/format'
import { useSession } from '@app/common/session'
import { useRouter } from 'next/router'
import { useState } from 'react'
import { Modal } from 'ui'
import { toast } from 'react-toastify'
import Link from 'next/link'
import { DateTime } from 'luxon'

interface Props {
  log: Log
}

export const LogDetailsV2 = ({ log }: Props) => {
  const fields = [
    { label: 'Language', value: log.language.name },
    { label: 'Activity', value: log.activity.name },
    {
      label: 'Amount',
      value: `${formatScore(log.amount)} ${formatUnit(log.amount, log.unit_name)}`,
    },
    ...(log.description
      ? [{ label: 'Description', value: log.description }]
      : []),
  ]

  const tags = log.tags

  return (
    <div>
      <div className="flex items-center justify-between">
        <div>
          <h1 className="title">Log details</h1>
          <h2 className="subtitle">
            By {log.user_display_name ?? 'anonymous'} on{' '}
            {DateTime.fromISO(log.created_at).toLocaleString(DateTime.DATE_MED)}
          </h2>
        </div>
        <div className="h-stack gap-2">
          <DeleteButton log={log} />
        </div>
      </div>

      <div className="my-6" />

      <div className="card">
        <div className="v-stack gap-3">
          {fields.map(field => (
            <div key={field.label} className="flex">
              <span className="w-32 text-sm text-neutral-500 flex-shrink-0">
                {field.label}
              </span>
              <span className="text-sm font-medium">{field.value}</span>
            </div>
          ))}
          {tags.length > 0 ? (
            <div className="flex">
              <span className="w-32 text-sm text-neutral-500 flex-shrink-0">
                Tags
              </span>
              <span className="text-sm font-medium">
                {tags
                  .map(tag => `#${tag.toLowerCase().replace(/\s+/g, '-')}`)
                  .join(' ')}
              </span>
            </div>
          ) : null}
        </div>
      </div>

      {log.registrations && log.registrations.length > 0 ? (
        <>
          <div className="my-6" />
          <div>
            <div className="flex items-center justify-between mb-2">
              <h3 className="subtitle">Submitted to contests</h3>
              <Link href={routes.logContests(log.id)} className="btn ghost text-sm">
                Edit submissions
              </Link>
            </div>
            <div className="v-stack gap-2">
              {log.registrations.map(reg => (
                <Link
                  key={reg.contest_id}
                  href={routes.contestLeaderboard(reg.contest_id)}
                  className="input-frame px-4 py-2 no-underline hover:bg-neutral-50 transition-colors"
                >
                  <div className="h-stack items-center w-full">
                    <div className="flex-1">
                      <div className="font-bold text-secondary">
                        {reg.title}
                      </div>
                    </div>
                    <span className="text-sm font-medium text-secondary">
                      Score: {formatScore(log.score)}
                    </span>
                  </div>
                </Link>
              ))}
            </div>
          </div>
        </>
      ) : null}
    </div>
  )
}

function DeleteButton({ log }: { log: Log }) {
  const [session] = useSession()
  const isOwner = log.user_id === session?.identity.id
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false)
  const router = useRouter()

  const mutation = useDeleteLog(
    () => {
      setIsDeleteModalOpen(false)
      toast.success('Deletion complete')
      router.push(routes.userProfileStatistics(log.user_id))
    },
    () => {
      setIsDeleteModalOpen(false)
      toast.error('Could not process deletion, please try again later.')
    },
  )

  if (!isOwner) {
    return null
  }

  return (
    <>
      <Modal
        isOpen={isDeleteModalOpen}
        setIsOpen={setIsDeleteModalOpen}
        title="Are you sure?"
      >
        <p className="modal-body">
          Deletion cannot be undone. The log will be permanently removed from all
          contests and your personal tracking history.
        </p>

        <div className="modal-actions spaced">
          <button
            type="button"
            className="btn danger"
            onClick={() => {
              const id = toast.info('Deleting log...')
              mutation.mutate(log.id)
              setTimeout(() => toast.dismiss(id), 200)
            }}
          >
            Yes, delete it
          </button>
          <button
            type="button"
            className="btn ghost"
            onClick={() => setIsDeleteModalOpen(false)}
          >
            Go back
          </button>
        </div>
      </Modal>
      <button
        type="button"
        className="btn danger gap-2"
        onClick={() => setIsDeleteModalOpen(true)}
      >
        <TrashIcon className="w-4 h-4 mr-2" />
        Delete
      </button>
    </>
  )
}
