import { describe, expect, it, vi } from 'vitest'
vi.mock('next/config', () => ({ default: () => ({ publicRuntimeConfig: {} }) }))
import { resolveAppUrls } from './app-urls'

const config = {
  homeUrl: 'https://tadoku.dev.lab',
  authUiUrl: 'https://account.tadoku.dev.lab',
  adminUrl: 'https://admin.tadoku.dev.lab',
  apiEndpoint: 'https://tadoku.dev.lab/api/internal',
}

describe('branch URLs', () => {
  it.each(['tadoku.dev.lab', 'account.tadoku.dev.lab', 'admin.tadoku.dev.lab', 'localhost:3000', 'evil.example', 'a.b.tadoku.dev.lab', '-x.tadoku.dev.lab', undefined])('keeps base URLs for %s', host => {
    expect(resolveAppUrls(config, host)).toEqual(config)
  })
  it.each(['alice-feat-1a2b3c4d.tadoku.dev.lab', 'alice-feat-1a2b3c4d.account.tadoku.dev.lab', 'ALICE-FEAT-1A2B3C4D.admin.tadoku.dev.lab:443'])('uses sibling branch hosts for %s', host => {
    expect(resolveAppUrls(config, host)).toEqual({
      branch: 'alice-feat-1a2b3c4d',
      homeUrl: 'https://alice-feat-1a2b3c4d.tadoku.dev.lab',
      authUiUrl: 'https://alice-feat-1a2b3c4d.account.tadoku.dev.lab',
      adminUrl: 'https://alice-feat-1a2b3c4d.admin.tadoku.dev.lab',
      apiEndpoint: 'https://alice-feat-1a2b3c4d.tadoku.dev.lab/api/internal',
    })
  })
  it('preserves relative and internal API endpoints', () => {
    for (const apiEndpoint of ['/api/internal', 'http://oathkeeper-proxy.tdk-prod-oathkeeper:4455/api/internal']) {
      expect(resolveAppUrls({ ...config, apiEndpoint }, 'alice.tadoku.dev.lab').apiEndpoint).toBe(apiEndpoint)
    }
  })
  it('recognizes production hosts', () => {
    const prod = Object.fromEntries(Object.entries(config).map(([key, value]) => [key, value.replace('tadoku.dev.lab', 'tadoku.app')]))
    expect(resolveAppUrls(prod, 'alice.account.tadoku.app').homeUrl).toBe('https://alice.tadoku.app')
  })
  it('requires all sibling base hosts', () => {
    expect(() => resolveAppUrls({ ...config, adminUrl: undefined }, 'account.tadoku.dev.lab')).toThrow()
  })
})
