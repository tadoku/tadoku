import {
  colorForActivity,
  formatArray,
  formatScore,
  formatUnit,
} from '@app/common/format'
import { useCurrentDateTime } from '@app/common/hooks'
import { routes } from '@app/common/routes'
import { useSession, useUserRole } from '@app/common/session'
import { Log, useDeleteLog, useLog } from '@app/immersion/api'
import { LogDetailsV2 } from '@app/immersion/LogDetailsV2'
import { HomeIcon, TrashIcon } from '@heroicons/react/20/solid'
import { XMarkIcon } from '@heroicons/react/24/outline'
import { DateTime } from 'luxon'
import Link from 'next/link'
import { useRouter } from 'next/router'
import { useState } from 'react'
import { Breadcrumb, ButtonGroup, Loading, Modal } from 'ui'
import { toast } from 'react-toastify'
import Head from 'next/head'

const Page = () => {
  const router = useRouter()
  const id = router.query['id']?.toString() ?? ''
  const log = useLog(id)
  const role = useUserRole()

  if (log.isLoading || log.isIdle || role === undefined) {
    return <Loading />
  }

  if (log.isError) {
    return (
      <span className="flash error">
        Could not load page, please try again later.
      </span>
    )
  }

  if (role === 'admin') {
    return (
      <>
        <Head>
          <title>Log details - Tadoku</title>
        </Head>
        <div className="pb-4">
          <Breadcrumb
            links={[
              { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
              {
                label: log.data.user_display_name!,
                href: routes.userProfileStatistics(log.data.user_id),
              },
              {
                label: 'Log details',
                href: routes.log(log.data.id),
              },
            ]}
          />
        </div>
        <LogDetailsV2 log={log.data} />
      </>
    )
  }

  const logColor = colorForActivity(log.data.activity.id)

  return (
    <>
      <Head>
        <title>Log details - Tadoku</title>
      </Head>
      <div className="pb-4">
        <Breadcrumb
          links={[
            { label: 'Home', href: routes.home(), IconComponent: HomeIcon },
            {
              label: log.data.user_display_name!,
              href: routes.userProfileStatistics(log.data.user_id),
            },
            {
              label: 'Log details',
              href: routes.log(log.data.id),
            },
          ]}
        />
      </div>
      <div className="h-stack justify-between items-center w-full">
        <div>
          <h1 className="title">Log details</h1>
          <h2 className="subtitle">
            By {log.data.user_display_name ?? 'anonymous'},{' '}
            {DateTime.fromISO(log.data.created_at).toLocaleString(
              DateTime.DATETIME_MED,
            )}
          </h2>
        </div>
        <div>
          <ActionBar log={log.data} />
        </div>
      </div>
      <div className="max-w-2xl mt-4">
        <div className="card w-full relative">
          <div
            className={`bg-${logColor}-300 absolute top-0 left-0 right-0 h-2`}
          ></div>
          <div className="flex flex-wrap gap-3">
            <span className={`tag bg-${logColor}-300 text-${logColor}-900`}>
              {log.data.activity.name}
            </span>
            <span className="tag text-slate-900 bg-slate-200">
              {log.data.language.name}
            </span>
            {log.data.tags.map(it => (
              <span key={it} className={`tag text-slate-900 bg-slate-200`}>
                {it}
              </span>
            ))}
            {log.data.deleted ? (
              <span className={`tag text-red-900 bg-red-200`}>Deleted</span>
            ) : null}
          </div>
          {log.data.registrations && log.data.registrations.length >= 1 ? (
            <p>
              Submitted to{' '}
              {formatArray(log.data.registrations, it => (
                <Link
                  key={it.contest_id}
                  href={routes.contestLeaderboard(it.contest_id)}
                  className="font-bold"
                >
                  {it.title}
                </Link>
              ))}
            </p>
          ) : null}
          {log.data.description ? (
            <>
              <h3 className="subtitle my-2">Description</h3>
              <p className="">{log.data.description}</p>
            </>
          ) : null}
          <div className="h-stack w-full mt-4 spaced">
            <div className="w-1/2">
              <h3 className="subtitle mb-2">Score</h3>
              <div className="font-bold text-5xl">
                {formatScore(log.data.score)}
              </div>
            </div>
            <div className="w-1/2 flex flex-col items-end justify-end opacity-80">
              <h4 className="subtitle text-sm">Breakdown</h4>
              <div className="lowercase flex items-center space-x-1 text-sm">
                <strong className="text-lg">
                  {formatScore(log.data.amount)}
                </strong>
                <span className="text-slate-500">
                  {formatUnit(log.data.amount, log.data.unit_name)}
                </span>
                <XMarkIcon className="w-3 h-3 mx-2 text-secondary" />
                <strong className="text-lg">{log.data.modifier}</strong>
                <span className="text-slate-500">modifier</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  )
}

function ActionBar({ log }: { log: Log }) {
  const now = useCurrentDateTime()
  const [session] = useSession()
  const canBeDeleted = !(log.registrations ?? [])
    .map(it => {
      const end = DateTime.fromISO(it.contest_end).endOf('day')
      const hasEnded = now.diff(end).as('seconds') < 0

      return hasEnded
    })
    .some(it => it === false)
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

  return (
    <>
      <Modal
        isOpen={isDeleteModalOpen}
        setIsOpen={setIsDeleteModalOpen}
        title="Are you sure?"
      >
        <p className="modal-body">
          Deletion cannot be undone. The log will be permanently removed from
          all contests and your personal tracking history.
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
      <ButtonGroup
        actions={[
          {
            onClick: () => setIsDeleteModalOpen(true),
            href: '#',
            label: 'Delete log',
            IconComponent: TrashIcon,
            style: 'ghost',
            visible: isOwner && canBeDeleted,
          },
        ]}
        orientation="right"
      />
    </>
  )
}

export default Page
