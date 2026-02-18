import { TrashIcon, CheckBadgeIcon } from '@heroicons/react/20/solid'
import { XMarkIcon } from '@heroicons/react/24/outline'
import { Log, useDeleteLog } from '@app/immersion/api'
import { routes } from '@app/common/routes'
import { colorForActivity, formatScore, formatUnit } from '@app/common/format'
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
  const logColor = colorForActivity(log.activity.id)
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

      <div className="max-w-2xl card narrow">
        <div className={`bg-${logColor}-200 -mx-4 -mt-4 mb-4 px-4 py-3 rounded-t`}>
          <div className="text-sm flex items-baseline gap-2">
            <strong>{log.language.name}</strong>
            <span>&middot;</span>
            <span>{log.activity.name}</span>
          </div>
        </div>
        {log.description ? (
          <div className="mb-4">
            <h3 className="subtitle">Description</h3>
            <p className="text-sm">{log.description}</p>
          </div>
        ) : null}
        {tags.length > 0 ? (
          <div className="flex flex-wrap gap-2 mb-4">
            {tags.map(tag => (
              <span key={tag} className="tag text-slate-900 bg-slate-200">
                #{tag}
              </span>
            ))}
          </div>
        ) : null}
        <div className="h-stack w-full spaced">
          <div className="w-1/2">
            <h3 className="subtitle mb-2">Score</h3>
            <div className="font-bold text-5xl">
              {formatScore(log.score)}
            </div>
          </div>
          <div className="w-1/2 flex flex-col items-end justify-end opacity-80">
            <h4 className="subtitle text-sm">Breakdown</h4>
            <div className="lowercase flex items-center space-x-1 text-sm">
              <strong className="text-lg">
                {formatScore(log.amount)}
              </strong>
              <span className="text-slate-500">
                {formatUnit(log.amount, log.unit_name)}
              </span>
              <XMarkIcon className="w-3 h-3 mx-2 text-secondary" />
              <strong className="text-lg">{log.modifier}</strong>
              <span className="text-slate-500">modifier</span>
            </div>
          </div>
        </div>
      </div>

      {log.registrations && log.registrations.length > 0 ? (
        <>
          <div className="my-6" />
          <div className="max-w-2xl card narrow">
            <div className="flex items-center justify-between mb-4">
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
                      {reg.official ? (
                        <div className="text-xs text-gray-600 flex items-center">
                          Administered by <strong className="ml-1">Tadoku</strong>
                          <CheckBadgeIcon className="ml-1 w-4 h-4 text-lime-700" />
                        </div>
                      ) : reg.owner_user_display_name ? (
                        <div className="text-xs text-gray-600">
                          Administered by <strong>{reg.owner_user_display_name}</strong>
                        </div>
                      ) : null}
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
