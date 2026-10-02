// @vitest-environment jsdom

import React from 'react'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { Pagination } from 'ui/components/Pagination'

Reflect.set(globalThis, 'React', React)

vi.mock('ui/node_modules/next/router', () => ({
  useRouter: () => ({ push: vi.fn() }),
}))
vi.mock('ui/node_modules/next/link', () => ({
  default: React.forwardRef<HTMLAnchorElement, React.ComponentProps<'a'>>(
    function Link(props, ref) {
      return <a {...props} ref={ref} />
    },
  ),
}))

const getHref = (page: number) => `/x/${page}`

const control = (name: 'Previous' | 'Next') =>
  screen.getByRole('link', { name })

const expectInert = (name: 'Previous' | 'Next') => {
  const element = control(name)
  expect(element.getAttribute('aria-disabled')).toBe('true')
  expect(element.hasAttribute('href')).toBe(false)

  // An unfocusable control cannot be activated with Enter.
  element.focus()
  expect(document.activeElement).not.toBe(element)
}

afterEach(() => {
  cleanup()
})

describe('Pagination boundary controls', () => {
  it('does not link to page 0 from the first page', () => {
    const { container } = render(
      <Pagination currentPage={1} totalPages={3} getHref={getHref} />,
    )
    expect(container.querySelector('[href="/x/0"]')).toBeNull()
    expectInert('Previous')
  })

  it('does not link past the last page', () => {
    const { container } = render(
      <Pagination currentPage={3} totalPages={3} getHref={getHref} />,
    )
    expect(container.querySelector('[href="/x/4"]')).toBeNull()
    expectInert('Next')
  })

  it('does not request page 0 from the first page', () => {
    const onClick = vi.fn()
    render(<Pagination currentPage={1} totalPages={3} onClick={onClick} />)
    expectInert('Previous')
    fireEvent.click(control('Previous'))
    expect(onClick).not.toHaveBeenCalled()
  })

  it('does not request a page past the last page', () => {
    const onClick = vi.fn()
    render(<Pagination currentPage={3} totalPages={3} onClick={onClick} />)
    expectInert('Next')
    fireEvent.click(control('Next'))
    expect(onClick).not.toHaveBeenCalled()
  })

  it('keeps Previous and Next clickable between the boundaries', () => {
    const onClick = vi.fn()
    render(<Pagination currentPage={2} totalPages={3} onClick={onClick} />)
    fireEvent.click(control('Previous'))
    expect(onClick).toHaveBeenLastCalledWith(1)
    fireEvent.click(control('Next'))
    expect(onClick).toHaveBeenLastCalledWith(3)
  })

  it('keeps Previous and Next linked between the boundaries', () => {
    render(<Pagination currentPage={2} totalPages={3} getHref={getHref} />)
    expect(control('Previous').getAttribute('href')).toBe('/x/1')
    expect(control('Next').getAttribute('href')).toBe('/x/3')
    expect(control('Previous').hasAttribute('aria-disabled')).toBe(false)
    expect(control('Next').hasAttribute('aria-disabled')).toBe(false)
  })
})
