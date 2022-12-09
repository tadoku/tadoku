import {
  ChevronLeftIcon,
  ChevronRightIcon,
  EllipsisHorizontalIcon,
} from '@heroicons/react/20/solid'
import classNames from 'classnames'

interface Props {
  totalPages: number
  currentPage: number
  getHref?: (page: number) => string
  onClick?: (page: number) => void
  window?: number
}

export default function Pagination({
  totalPages: total,
  currentPage: current,
  getHref,
  onClick,
  window = 4,
}: Props) {
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
    : (page: number) => () => onClick(page)

  return (
    <nav className="flex" aria-label="Breadcrumb">
      <a
        className={classNames('btn ghost', {
          'pointer-events-none disabled': !canGoPrevious,
        })}
        href={getHref?.(current - 1) ?? '#'}
        onClick={clickHandler?.(current - 1)}
      >
        <ChevronLeftIcon className="w-5 h-5 mr-2" />
        Previous
      </a>
      <ol className="inline-flex items-center justify-center space-x-1 md:space-x-3 w-full">
        {start > 1 ? (
          <>
            <Page
              href={getHref?.(1) ?? '#'}
              page={1}
              isActive={current === 1}
              onClick={clickHandler?.(1)}
            />
            {start > 2 ? <Spacer /> : null}
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
            {end < total - 1 ? <Spacer /> : null}
            <Page
              href={getHref?.(total) ?? '#'}
              page={total}
              isActive={current === total}
              onClick={clickHandler?.(total)}
            />
          </>
        ) : null}
      </ol>

      <a
        className={classNames('btn ghost', {
          'pointer-events-none disabled': !canGoNext,
        })}
        href={getHref?.(current + 1) ?? '#'}
        onClick={clickHandler?.(current + 1)}
      >
        Next
        <ChevronRightIcon className="w-5 h-5 ml-2" />
      </a>
    </nav>
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
  onClick?: () => void | undefined
}) => (
  <li className="inline-flex items-center">
    <a
      href={href}
      className={classNames(
        'reset min-w-[50px] box-border border-b-2 px-4 py-1 h-11 inline-flex items-center justify-center hover:bg-secondary/5 focus:bg-secondary/5',
        {
          'border-primary font-bold text-primary hover:text-primary': isActive,
          'font-medium border-transparent text-secondary !hover:text-secondary':
            !isActive,
        },
      )}
      onClick={onClick}
    >
      {page}
    </a>
  </li>
)

const Spacer = () => (
  <li className="inline-flex items-center text-gray-300 ">
    <EllipsisHorizontalIcon className="w-5 h-5" />
  </li>
)
