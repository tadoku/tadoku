// @vitest-environment jsdom

import React from 'react'
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import Navigation from './Navigation'

Reflect.set(globalThis, 'React', React)

const state = vi.hoisted(() => ({
  signedIn: true,
  role: 'user',
  pathname: '/',
  navigate: vi.fn(),
}))

vi.mock('@app/common/session', () => ({
  useSession: () => [
    state.signedIn
      ? { identity: { id: 'reader', traits: { display_name: 'Reader' } } }
      : undefined,
  ],
  useUserRole: () => state.role,
  useLogoutHandler: () => vi.fn(),
}))
vi.mock('@app/common/hooks', () => ({ useCurrentLocation: () => '/' }))
vi.mock('react-query', () => ({ useIsFetching: () => 0 }))
vi.mock('next/config', () => ({
  default: () => ({ publicRuntimeConfig: { authUiUrl: '/account' } }),
}))
vi.mock('ui/node_modules/next/router', () => ({
  useRouter: () => ({ pathname: state.pathname }),
}))
vi.mock('ui/node_modules/next/link', () => ({
  default: React.forwardRef<HTMLAnchorElement, React.ComponentProps<'a'>>(
    function Link({ onClick, ...props }, ref) {
      return (
        <a
          {...props}
          ref={ref}
          onClick={event => {
            onClick?.(event)
            if (!event.defaultPrevented) state.navigate(props.href)
            event.preventDefault()
          }}
        />
      )
    },
  ),
}))
vi.mock('ui/components/branding', () => ({ Logo: () => <span>Tadoku</span> }))

beforeEach(() => {
  state.signedIn = true
  state.role = 'user'
  state.pathname = '/'
  state.navigate.mockClear()
  vi.stubGlobal('scrollTo', vi.fn())
})

afterEach(() => {
  cleanup()
  vi.unstubAllGlobals()
})

describe('New log header action', () => {
  it('links directly to New log for signed-in readers', () => {
    render(<Navigation />)
    // JSDOM has no Tailwind layout: desktop and mobile copies are both present.
    const actions = screen.getAllByRole('link', { name: 'New log' })
    expect(actions).toHaveLength(2)
    for (const action of actions) {
      expect(action.getAttribute('href')).toBe('/logs/new')
    }
    fireEvent.click(actions[0])
    expect(state.navigate).toHaveBeenCalledWith('/logs/new')
  })

  it.each(['user', 'admin'])(
    'hides the mobile action while the %s menu is open',
    async role => {
      state.role = role
      render(<Navigation />)
      fireEvent.click(screen.getByRole('button', { name: 'Open main menu' }))

      await waitFor(() => {
        // Only the desktop copy remains; there is no duplicate in the drawer.
        expect(screen.getAllByRole('link', { name: 'New log' })).toHaveLength(1)
        expect(
          screen.getByRole('button', { name: 'Close main menu' }),
        ).toBeTruthy()
      })
      expect(screen.getByRole('link', { name: 'Profile' })).toBeTruthy()
      fireEvent.click(screen.getByRole('button', { name: 'Close main menu' }))
      await waitFor(() =>
        expect(screen.getAllByRole('link', { name: 'New log' })).toHaveLength(
          2,
        ),
      )
    },
  )

  it('does not expose New log when signed out', () => {
    state.signedIn = false
    render(<Navigation />)
    expect(screen.queryByRole('link', { name: 'New log' })).toBeNull()
    expect(screen.getByRole('link', { name: 'Log in' })).toBeTruthy()
    expect(screen.getByRole('link', { name: 'Sign up' })).toBeTruthy()
    fireEvent.click(screen.getByRole('button', { name: 'Open main menu' }))
    expect(screen.queryByRole('link', { name: 'New log' })).toBeNull()
  })

  it('does not navigate away from an in-progress New log form', () => {
    state.pathname = '/logs/new'
    render(<Navigation />)
    const action = screen.getAllByRole('link', { name: 'New log' })[0]
    expect(action.getAttribute('aria-current')).toBe('page')
    fireEvent.click(action)
    expect(state.navigate).not.toHaveBeenCalled()

    // Opening a separate tab remains possible.
    fireEvent.click(action, { ctrlKey: true })
    expect(state.navigate).toHaveBeenCalledWith('/logs/new')
  })
})
