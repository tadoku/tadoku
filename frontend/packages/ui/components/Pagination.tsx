import {
  ChevronLeftIcon,
  ChevronRightIcon,
  EllipsisHorizontalIcon,
} from '@heroicons/react/20/solid'
import classNames from 'classnames'
import Link from 'next/link'
import { Dispatch, SetStateAction, useEffect, useState } from 'react'
import { FieldValues, FormProvider, useForm } from 'react-hook-form'
import { Input } from './Form'
import { Modal } from './Modal'

interface Props {
  totalPages: number
  currentPage: number
  getHref?: (page: number) => string
  onClick?: (page: number) => void
  window?: number
}

export function Pagination({
  totalPages: total,
  currentPage: current,
  getHref,
  onClick,
  window = 4,
}: Props) {
  const [isNavigationModalOpen, setIsNavigationModalOpen] = useState(false)

  const canGoPrevious = current > 1
  const canGoNext = current < total

  let start = current - window / 2
  let end = current + window / 2

  if (start <= 0) {
    end = end - start + 1
    start = 1
  }

  if (end > total) {
    end = total
    start = Math.max(end - window, 1)
  }

  const clickHandler = !onClick
    ? undefined
    : (page: number): React.MouseEventHandler<HTMLAnchorElement> =>
        e => {
          e.preventDefault()
          onClick(page)
        }

  return (
    <>
      {onClick ? (
        <NavigateForm
          isOpen={isNavigationModalOpen}
          setIsOpen={setIsNavigationModalOpen}
          setPage={onClick}
          totalPages={total}
        />
      ) : null}
      <nav className="flex justify-between" aria-label="breadcrumb">
        <Link
          className={classNames('btn ghost', {
            'pointer-events-none disabled': !canGoPrevious,
          })}
          href={getHref?.(current - 1) ?? '#'}
          onClick={clickHandler?.(current - 1)}
        >
          <ChevronLeftIcon className="w-5 h-5 mr-2" />
          Previous
        </Link>

        <div className="md:hidden flex align-middle">
          <Spacer onClick={() => setIsNavigationModalOpen(true)} />
        </div>

        <ol className="hidden md:inline-flex items-center justify-center space-x-1 md:space-x-3 w-full">
          {start > 1 ? (
            <>
              <Page
                href={getHref?.(1) ?? '#'}
                page={1}
                isActive={current === 1}
                onClick={clickHandler?.(1)}
              />
              {start === 3 ? (
                <Page
                  href={getHref?.(2) ?? '#'}
                  page={2}
                  isActive={current === 2}
                  onClick={clickHandler?.(2)}
                />
              ) : null}
              {start > 3 ? (
                <li>
                  <Spacer onClick={() => setIsNavigationModalOpen(true)} />
                </li>
              ) : null}
            </>
          ) : null}
          {Array.from({ length: end - start + 1 }, (_, i) => i + start).map(
            page => (
              <Page
                key={page}
                href={getHref?.(page) ?? '#'}
                page={page}
                isActive={current === page}
                onClick={clickHandler?.(page)}
              />
            ),
          )}
          {end < total ? (
            <>
              {end < total - 2 ? (
                <li>
                  <Spacer onClick={() => setIsNavigationModalOpen(true)} />
                </li>
              ) : null}

              {end === total - 2 ? (
                <Page
                  href={getHref?.(total - 1) ?? '#'}
                  page={total - 1}
                  isActive={current === total - 1}
                  onClick={clickHandler?.(total - 1)}
                />
              ) : null}
              <Page
                href={getHref?.(total) ?? '#'}
                page={total}
                isActive={current === total}
                onClick={clickHandler?.(total)}
              />
            </>
          ) : null}
        </ol>

        <Link
          className={classNames('btn ghost', {
            'pointer-events-none disabled': !canGoNext,
          })}
          href={getHref?.(current + 1) ?? '#'}
          onClick={clickHandler?.(current + 1)}
        >
          Next
          <ChevronRightIcon className="w-5 h-5 ml-2" />
        </Link>
      </nav>
    </>
  )
}

const Page = ({
  href,
  page,
  isActive,
  onClick,
}: {
  href: string
  page: number
  isActive: boolean
  onClick?: React.MouseEventHandler<HTMLAnchorElement> | undefined
}) => (
  <li className="inline-flex items-center">
    <Link
      href={href}
      className={classNames(
        'reset min-w-[50px] box-border px-4 py-1 h-11 inline-flex items-center justify-center hover:bg-secondary/5 focus:bg-secondary/5',
        {
          'font-bold text-secondary bg-secondary/5 hover:bg-secondary/10':
            isActive,
          'font-medium': !isActive,
        },
      )}
      onClick={isActive ? undefined : onClick}
    >
      {page}
    </Link>
  </li>
)

const Spacer = ({ onClick }: { onClick?: () => void }) => (
  <button
    className={`flex items-center text-gray-300 ${
      onClick ? 'hover:text-secondary' : 'pointer-events-none'
    }`}
    onClick={onClick}
  >
    <EllipsisHorizontalIcon className="w-5 h-5" />
  </button>
)

const NavigateForm = ({
  isOpen,
  setIsOpen,
  setPage,
  totalPages: total,
}: {
  isOpen: boolean
  setIsOpen: Dispatch<SetStateAction<boolean>>
  setPage: (page: number) => void
  totalPages: number
}) => {
  const methods = useForm()
  const onSubmit = ({ page }: FieldValues) => {
    setPage(page)
    setIsOpen(false)
  }

  useEffect(() => methods.reset(), [isOpen])

  return (
    <Modal isOpen={isOpen} setIsOpen={setIsOpen}>
      <FormProvider {...methods}>
        <form
          onSubmit={methods.handleSubmit(onSubmit)}
          className="v-stack spaced"
        >
          <Input
            name="page"
            label="Navigate to page"
            type="number"
            options={{
              required: 'This field is required',
              valueAsNumber: true,
            }}
            min={1}
            max={total}
          />
          <p className="modal-body"></p>

          <div className="modal-actions">
            <button
              type="submit"
              className="btn secondary"
              disabled={methods.formState.isSubmitting}
            >
              Go
            </button>
            <button
              type="button"
              className="btn ghost"
              onClick={() => setIsOpen(false)}
            >
              Cancel
            </button>
          </div>
        </form>
      </FormProvider>
    </Modal>
  )
}
