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
  const prod = Object.fromEntries(Object.entries(config).map(([key, value]) => [key, value.replace('tadoku.dev.lab', 'tadoku.app')]))
  it.each(['alice.preview.tadoku.app', 'alice.account.preview.tadoku.app', 'ALICE.admin.preview.tadoku.app:443'])('uses the reserved production namespace for %s', host => {
    expect(resolveAppUrls(prod, host)).toEqual({
      branch: 'alice',
      homeUrl: 'https://alice.preview.tadoku.app',
      authUiUrl: 'https://alice.account.preview.tadoku.app',
      adminUrl: 'https://alice.admin.preview.tadoku.app',
      apiEndpoint: 'https://alice.preview.tadoku.app/api/internal',
    })
  })
  it.each(['tadoku.app', 'account.tadoku.app', 'admin.tadoku.app', 'alice.tadoku.app', 'alice.account.tadoku.app', 'alice.admin.tadoku.app', 'alice.preview.tadoku.app.evil.example', 'alice.account.preview.tadoku.app.evil.example', 'a.b.preview.tadoku.app', '-alice.preview.tadoku.app', 'alice-.preview.tadoku.app'])('keeps production base URLs for %s', host => {
    expect(resolveAppUrls(prod, host)).toEqual(prod)
  })
  it('preserves paths, ports and private API origins on previews', () => {
    const configured = { ...prod, authUiUrl: 'https://account.tadoku.app:8443/login/', apiEndpoint: 'http://oathkeeper-proxy.tdk-prod-oathkeeper:4455/api/internal' }
    const urls = resolveAppUrls(configured, 'alice.preview.tadoku.app')
    expect(urls.authUiUrl).toBe('https://alice.account.preview.tadoku.app:8443/login/')
    expect(urls.apiEndpoint).toBe(configured.apiEndpoint)
  })
  it('requires all sibling base hosts', () => {
    expect(() => resolveAppUrls({ ...config, adminUrl: undefined }, 'account.tadoku.dev.lab')).toThrow()
  })
})
