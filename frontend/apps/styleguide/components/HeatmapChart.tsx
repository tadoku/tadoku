import { DateTime, Interval } from 'luxon'
import { useRef, useEffect, useState, MutableRefObject } from 'react'
import { createPortal } from 'react-dom'

interface Cell {
  x: DateTime
  y: number
  value: number
}

interface Props {
  data: { date: string; value: number }[]
  year: number
  id: string
}

const colWidth = 10
const rowHeight = 10
const padding = 3
const offset = {
  x: 30,
  y: 15,
}

function HeatmapChart({ id, data, year }: Props) {
  const start = DateTime.fromObject({ year, month: 1, day: 1 })
  const end = DateTime.fromObject({ year, month: 12, day: 31 })

  const values = new Map<string, number>()
  for (const { date, value } of data) {
    values.set(date, value)
  }

  const dates = Interval.fromDateTimes(start, end.endOf('day'))
    .splitBy({ day: 1 })
    .map(it => it.start)

  const weeks = start.weeksInWeekYear + 1
  const cols: (Cell | undefined)[][] = Array.from(Array(weeks)).map(_ =>
    Array.from(Array(7)),
  )

  let col = 0
  for (const date of dates) {
    cols[col][date.weekday - 1] = {
      x: date,
      y: date.weekday,
      value: values.get(date.toISODate()) ?? 0,
    }

    if (date.weekday === 7) {
      col += 1
    }
  }

  const allValues = cols.flatMap(it => it.map(it => it?.value ?? 0))
  const maxValue = Math.max(...allValues)

  const weekdays = ['Mon', undefined, 'Wed', undefined, 'Fri', undefined, 'Sun']

  let lastMonth = 0
  const months = cols.map(rows => {
    const shouldRender = rows
      .filter(it => it !== undefined)
      .some(cell => cell!.x.month > lastMonth)
    if (!shouldRender) {
      return undefined
    }

    lastMonth += 1
    return DateTime.fromObject({ month: lastMonth }).toFormat('LLL')
  })

  const tooltipId = `tooltip-${id}`

  return (
    <svg
      width={weeks * colWidth + (weeks - 1) * padding + offset.x}
      height={7 * rowHeight + 6 * padding + offset.y}
      className="overflow-visible"
    >
      {weekdays.map((label, row) => {
        if (!label) {
          return null
        }

        return (
          <text
            textAnchor="end"
            x={offset.x - padding * 2}
            y={offset.y + rowHeight * row + padding * row}
            className=""
            alignmentBaseline="hanging"
            style={{ fontSize: 10 }}
          >
            {label}
          </text>
        )
      })}
      {months.map((label, col) => {
        if (!label) {
          return null
        }

        return (
          <text
            textAnchor="start"
            x={offset.x + colWidth * col + padding * col}
            y={0}
            className=""
            alignmentBaseline="hanging"
            style={{ fontSize: 10 }}
          >
            {label}
          </text>
        )
      })}
      {cols.map((rows, col) =>
        rows.map((cell, row) => (
          <Cell
            value={cell?.value}
            tooltipId={tooltipId}
            maxValue={maxValue}
            col={col}
            row={row}
            tooltip={cell?.x.toLocaleString(DateTime.DATE_FULL) ?? ''}
          />
        )),
      )}
      <g id={tooltipId} className="outline-none"></g>
    </svg>
  )
}

function Cell({
  value,
  tooltipId,
  maxValue,
  tooltip,
  row,
  col,
}: {
  value: number | undefined
  tooltipId: string
  maxValue: number
  tooltip: string
  row: number
  col: number
}) {
  const [mounted, setMounted] = useState(false)
  const [hoverRef, isHovered] = useHover<SVGRectElement>()

  useEffect(() => {
    setMounted(true)

    return () => setMounted(false)
  }, [])

  if (!mounted || value === undefined) {
    return null
  }

  const x = offset.x + colWidth * col + padding * col
  const y = offset.y + rowHeight * row + padding * row

  const target = mounted ? document.getElementById(tooltipId) : null

  return (
    <>
      <rect
        width={colWidth}
        height={rowHeight}
        x={x}
        y={y}
        fill={'transparent'}
        className={`${getCellDepthClass(maxValue, value)}`}
        strokeWidth={0}
        ref={hoverRef}
      ></rect>
      {target &&
        createPortal(
          <Tooltip row={row} col={col} visible={isHovered}>
            {tooltip}
          </Tooltip>,
          target,
        )}
    </>
  )
}

function useHover<T>(): [MutableRefObject<T>, boolean] {
  const [value, setValue] = useState<boolean>(false)
  const ref: any = useRef<T | null>(null)
  const handleMouseOver = (): void => setValue(true)
  const handleMouseOut = (): void => setValue(false)
  useEffect(
    () => {
      const node: any = ref.current
      if (node) {
        node.addEventListener('mouseover', handleMouseOver)
        node.addEventListener('mouseout', handleMouseOut)
        return () => {
          node.removeEventListener('mouseover', handleMouseOver)
          node.removeEventListener('mouseout', handleMouseOut)
        }
      }
    },
    [ref.current], // Recall only if ref changes
  )
  return [ref, value]
}

function Tooltip({
  row,
  col,
  children,
  visible,
}: {
  row: number
  col: number
  children: React.ReactNode
  visible: boolean
}) {
  const ref = useRef<SVGTextElement>(null)
  const [tooltipRect, setTooltipRect] = useState({ x: 0, y: 0, w: 0, h: 0 })

  useEffect(() => {
    if (ref && ref.current) {
      const textRect = ref.current.getBoundingClientRect()
      console.log(textRect)

      const w = textRect.width + 12
      const h = textRect.height + 12
      const x = offset.x + colWidth * col + padding * col - w / 2
      const y = offset.y + rowHeight * row + padding * row - h - 2

      setTooltipRect({ x: x, y: y, w: w, h: h })
    }
  }, [ref, visible])

  return (
    <g className={`${visible ? '' : 'hidden'}`}>
      <rect
        width={tooltipRect.w}
        height={tooltipRect.h}
        x={tooltipRect.x}
        y={tooltipRect.y}
        className={`fill-secondary`}
        style={{
          filter:
            'drop-shadow(0 4px 3px rgb(0 0 0 / 0.07)) drop-shadow(0 2px 2px rgb(0 0 0 / 0.06))',
        }}
      ></rect>
      <polygon
        points={pointsForRect(tooltipRect)}
        className={`fill-secondary`}
      />
      <text
        fill={'white'}
        x={tooltipRect.x + 6}
        y={tooltipRect.y + 8}
        alignmentBaseline="hanging"
        ref={ref}
        className="text-xs"
      >
        {children}
      </text>
    </g>
  )
}

function pointsForRect({
  x,
  y,
  w,
  h,
}: {
  x: number
  y: number
  w: number
  h: number
}) {
  const size = 8

  const middle = x + w / 2 + colWidth / 2
  const left = middle - size / 2
  const right = middle + size / 2
  const top = y + h
  const bottom = top + 4

  return `${left},${top} ${right},${top} ${middle},${bottom}`
}

function getCellDepthClass(max: number, value: number) {
  if (value === 0) {
    return 'fill-stone-200'
  }

  const ratio = value / max

  if (ratio < 0.25) {
    return 'fill-teal-100'
  }
  if (ratio < 0.5) {
    return 'fill-teal-300'
  }
  if (ratio < 0.75) {
    return 'fill-teal-500'
  }

  return 'fill-teal-700'
}

export default HeatmapChart
