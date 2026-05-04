import {
  ChevronRightIcon,
  EllipsisVerticalIcon,
  TrashIcon,
} from '@heroicons/react/20/solid'
import { routes } from '@app/common/routes'
import { DateTime } from 'luxon'
import Link from 'next/link'
import {
  Logs,
  Log,
  useContest,
  useDetachLogFromContest,
  getContestLogsQueryKey,
} from '@app/immersion/api'
import { UseQueryResult, useQueryClient } from 'react-query'
import { colorForActivity, formatScore, formatUnit } from '@app/common/format'
import { Loading, ActionMenu, Modal } from 'ui'
import { useState } from 'react'
import { useSession, useUserRole } from '@app/common/session'
import { toast } from 'react-toastify'

function truncate(text: string | undefined, len: number) {
  if (text === undefined) {
    return text
  }

  if (text.length > len - 3) {
    return text.substring(0, len - 3) + '...'
  }

  return text
}

interface Props {
  logs: UseQueryResult<Logs>
  showUsername?: boolean
  contestId?: string
}

const LogsList = ({ logs, showUsername = false, contestId }: Props) => {
  const [session] = useSession()
  const role = useUserRole()
  const contest = useContest(contestId ?? '', { enabled: !!contestId })
  const queryClient = useQueryClient()
  const [modalOpen, setModalOpen] = useState(false)
  const [selectedLog, setSelectedLog] = useState<Log | null>(null)
  const [reason, setReason] = useState('')

  const isContestOwner =
    contestId &&
    contest.data?.owner_user_id &&
    session?.identity?.id === contest.data.owner_user_id
  const canModerate = isContestOwner || role === 'admin'

  const detachMutation = useDetachLogFromContest(
    () => {
      toast.success('Log removed from contest')
      if (contestId && logs.data) {
        queryClient.invalidateQueries(
          getContestLogsQueryKey({
            contestId,
            pageSize: logs.data.total_size,
            page: 1,
            includeDeleted: false,
          }),
        )
      }
      setModalOpen(false)
      setReason('')
      setSelectedLog(null)
    },
    () => {
      toast.error('Failed to remove log')
    },
  )

  const handleDetach = () => {
    if (!selectedLog || !contestId || !reason.trim()) return
    detachMutation.mutate({ contestId, logId: selectedLog.id, reason })
  }

  if (logs.isLoading || logs.isIdle) {
    return <Loading className="pb-4" />
  }

  if (logs.isError) {
    return <span className="flash error">Could not load updates</span>
  }

  return (
    <div className="table-container shadow-transparent w-auto">
      <table className="default shadow-transparent">
        <thead>
          <tr>
            {showUsername ? (
              <th className="default w-36">Participant</th>
            ) : null}
            <th className="default w-28 hidden md:table-cell">Activity</th>
            <th className="default w-36">Date</th>
            <th className="default w-32">Language</th>
            <th className="default hidden lg:table-cell">Description</th>
            <th className="default w-36 hidden md:table-cell">Amount</th>
            <th className="default w-24 !text-right">Score</th>
            <th className="default"></th>
          </tr>
        </thead>
        <tbody>
          {logs.data.logs.map(it => (
            <tr key={it.id} className="link">
              {showUsername ? (
                <td className="default link">
                  <Link className="reset" href={routes.log(it.id)}>
                    {it.user_display_name ?? 'Unknown user'}
                  </Link>
                </td>
              ) : null}
              <td className="default link hidden md:table-cell">
                <Link className="reset" href={routes.log(it.id)}>
                  <span
                    className={`tag bg-${colorForActivity(
                      it.activity.id,
                    )}-300 text-${colorForActivity(it.activity.id)}-900`}
                  >
                    {it.activity.name}
                  </span>
                </Link>
              </td>
              <td className="default link">
                <Link className="reset" href={routes.log(it.id)}>
                  {DateTime.fromISO(it.created_at).toLocaleString(
                    DateTime.DATE_MED,
                  )}
                </Link>
              </td>
              <td className="default link">
                <Link className="reset" href={routes.log(it.id)}>
                  {it.language.name}
                </Link>
              </td>
              <td
                className={`default text-sm link hidden lg:table-cell ${
                  !it.description ? 'opacity-50' : ''
                }`}
              >
                <Link className="reset" href={routes.log(it.id)}>
                  {truncate(it.description, 38) ?? 'N/A'}
                </Link>
              </td>
              <td className="default link font-bold hidden md:table-cell">
                <Link className="reset" href={routes.log(it.id)}>
                  {it.amount != null && it.unit_name != null
                    ? `${formatScore(it.amount)} ${formatUnit(it.amount, it.unit_name)}`
                    : it.duration_seconds != null
                      ? `${Math.round(it.duration_seconds / 60)} min`
                      : 'N/A'}
                </Link>
              </td>
              <td className="default link font-bold">
                <Link className="reset justify-end" href={routes.log(it.id)}>
                  {formatScore(it.score)}
                </Link>
              </td>
              <td className="default link w-12">
                {canModerate ? (
                  <ActionMenu
                    links={[
                      {
                        label: 'Remove from contest',
                        href: '#',
                        IconComponent: TrashIcon,
                        type: 'danger',
                        onClick: () => {
                          setSelectedLog(it)
                          setModalOpen(true)
                        },
                      },
                    ]}
                  >
                    <EllipsisVerticalIcon className="w-5 h-5" />
                  </ActionMenu>
                ) : (
                  <Link className="reset flex-shrink" href={routes.log(it.id)}>
                    <ChevronRightIcon className="w-5 h-5" />
                  </Link>
                )}
              </td>
            </tr>
          ))}
          {logs.data.logs.length === 0 ? (
            <tr>
              <td
                colSpan={(showUsername ? 8 : 7) + (canModerate ? 1 : 0)}
                className="default h-32 font-bold text-center text-xl text-slate-400"
              >
                No updates submitted yet
              </td>
            </tr>
          ) : null}
        </tbody>
      </table>
      <Modal
        isOpen={modalOpen}
        setIsOpen={setModalOpen}
        title="Remove log from contest"
      >
        <p className="modal-body">
          Why are you removing this log from the contest? Please provide a
          reason.
        </p>
        <div className="modal-body">
          <label className="label">
            <span className="label-text">Reason</span>
            <textarea
              className="input"
              value={reason}
              onChange={e => setReason(e.target.value)}
              placeholder="e.g. Duplicate entry, incorrect data..."
              rows={4}
            />
          </label>
        </div>
        <div className="modal-actions">
          <button
            type="button"
            className="btn danger"
            onClick={handleDetach}
            disabled={!reason.trim() || detachMutation.isLoading}
          >
            {detachMutation.isLoading ? 'Removing...' : 'Yes, remove it'}
          </button>
          <button
            type="button"
            className="btn ghost"
            onClick={() => {
              setModalOpen(false)
              setReason('')
              setSelectedLog(null)
            }}
          >
            Cancel
          </button>
        </div>
      </Modal>
    </div>
  )
}

export default LogsList
