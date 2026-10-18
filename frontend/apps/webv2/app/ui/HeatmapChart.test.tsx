// @vitest-environment jsdom

import React from 'react'
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { HeatmapChart } from 'ui/components/HeatmapChart'

Reflect.set(globalThis, 'React', React)

vi.mock('next/link', () => ({
  default: React.forwardRef<HTMLAnchorElement, React.ComponentProps<'a'>>(
    function Link(props, ref) {
      return <a {...props} ref={ref} />
    },
  ),
}))

const getHref = (date: string) => `/updates/1?date=${date}`
const day = (date: string, value: number) => ({
  date,
  value,
  tooltip: `${value} points on ${date}`,
})

const cellClass = (date: string, value: number) =>
  screen
    .getByRole('link', { name: `${value} points on ${date}` })
    .querySelector('rect')!
    .getAttribute('class')

const swatchClass = (label: string) =>
  screen
    .getByText(label)
    .parentElement!.querySelector('rect')!
    .getAttribute('class')

afterEach(() => {
  cleanup()
})

describe('Heatmap daily point scale', () => {
  it('keeps a day the same colour when record days are added', () => {
    const data = [day('2023-01-02', 40)]
    const { rerender } = render(
      <HeatmapChart id="outliers" year={2023} data={data} getHref={getHref} />,
    )
    const originalClass = cellClass('2023-01-02', 40)

    for (const value of [520, 100000]) {
      data.push(day(`2023-01-0${data.length + 2}`, value))
      rerender(
        <HeatmapChart
          id="outliers"
          year={2023}
          data={data}
          getHref={getHref}
        />,
      )
      expect(cellClass('2023-01-02', 40)).toBe(originalClass)
    }
  })

  it('matches the legend at inclusive lower bounds and fractional values', () => {
    const values = [0.5, 29.5, 30, 99.5, 100, 299.5, 300]
    const labels = ['< 30', '< 30', '30+', '30+', '100+', '100+', '300+']
    const { container } = render(
      <HeatmapChart
        id="boundaries"
        year={2023}
        data={values.map((value, index) => day(`2023-01-0${index + 1}`, value))}
        getHref={getHref}
      />,
    )

    expect(
      new Set(['0', '< 30', '30+', '100+', '300+'].map(swatchClass)).size,
    ).toBe(5)
    values.forEach((value, index) => {
      expect(cellClass(`2023-01-0${index + 1}`, value)).toBe(
        swatchClass(labels[index]),
      )
    })
    expect(
      container
        .querySelector('svg > rect[class]:not([class=""])')!
        .getAttribute('class'),
    ).toBe(swatchClass('0'))
  })

  it('shows the zero legend colour for every day in empty and all-zero years', () => {
    for (const year of [2023, 2024]) {
      for (const data of [[], [day(`${year}-01-02`, 0)]]) {
        const { container, unmount } = render(
          <HeatmapChart id="empty" year={year} data={data} getHref={getHref} />,
        )
        expect(screen.getByText('Points per day')).toBeTruthy()
        expect(screen.queryAllByRole('link')).toHaveLength(0)

        const cells = Array.from(
          container.querySelector('svg')!.querySelectorAll('rect'),
        )
        const zeroClass = swatchClass('0')
        expect(
          cells.filter(cell => cell.getAttribute('class') === zeroClass),
        ).toHaveLength(year === 2024 ? 366 : 365)
        const padding = cells.filter(
          cell => cell.getAttribute('class') !== zeroClass,
        )
        expect(padding.length).toBeGreaterThan(0)
        padding.forEach(cell => expect(cell.getAttribute('class')).toBe(''))
        unmount()
      }
    }
  })

  it('colours a lone 40-point day with the 30+ legend swatch', () => {
    render(
      <HeatmapChart
        id="single"
        year={2023}
        data={[day('2023-01-02', 40)]}
        getHref={getHref}
      />,
    )
    expect(cellClass('2023-01-02', 40)).toBe(swatchClass('30+'))
    expect(cellClass('2023-01-02', 40)).not.toBe(swatchClass('300+'))
  })
})
