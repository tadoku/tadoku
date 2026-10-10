import { describe, expect, it, vi } from 'vitest'
vi.mock('next/config', () => ({ default: () => ({ publicRuntimeConfig: {} }) }))
import { branchFlowUrl, branchUrl, resolveAppUrls } from './app-urls'

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

const production = {
  homeUrl: 'https://tadoku.app',
  authUiUrl: 'https://account.tadoku.app',
  adminUrl: 'https://admin.tadoku.app',
  apiEndpoint: 'https://tadoku.app/api/internal',
  kratosPublicEndpoint: 'https://account.tadoku.app/kratos',
}
const branchApp = 'https://alice-feat-1a2b3c4d.preview.tadoku.app'
const branchAccount = 'https://alice-feat-1a2b3c4d.account.preview.tadoku.app'

describe('Kratos redirects on branch hosts', () => {
  it.each([
    ['https://account.tadoku.app/login?flow=f1&return_to=x', `${branchAccount}/login?flow=f1&return_to=x`],
    ['https://account.tadoku.app/?flow=f2', `${branchAccount}/?flow=f2`],
    ['https://admin.tadoku.app/users', 'https://alice-feat-1a2b3c4d.admin.preview.tadoku.app/users'],
  ])('maps %s to the branch sibling', (url, expected) => {
    expect(branchUrl(production, 'alice-feat-1a2b3c4d.account.preview.tadoku.app', url)).toBe(expected)
  })
  it.each([
    'https://account.tadoku.app/kratos/self-service/login/browser?aal=aal2',
    'https://evil.example/login',
    '/login?flow=f1',
  ])('keeps Kratos, foreign and relative URL %s', url => {
    expect(branchUrl(production, 'alice-feat-1a2b3c4d.account.preview.tadoku.app', url)).toBe(url)
  })
  it.each(['account.tadoku.app', undefined])('keeps redirects unchanged on %s', host => {
    expect(branchUrl(production, host, 'https://account.tadoku.app/login?flow=f1')).toBe('https://account.tadoku.app/login?flow=f1')
  })
})

describe('forwarding branch flows from the production account host', () => {
  const page = 'https://account.tadoku.app/verification?flow=f1#top'
  it('forwards a flow whose return_to is on a branch host', () => {
    expect(branchFlowUrl(production, page, { return_to: `${branchApp}/profile`, request_url: '' }))
      .toBe(`${branchAccount}/verification?flow=f1#top`)
  })
  it('uses the return_to carried by request_url', () => {
    const request_url = `https://account.tadoku.app/kratos/self-service/registration/browser?return_to=${encodeURIComponent('https://alice-feat-1a2b3c4d.admin.preview.tadoku.app/')}`
    expect(branchFlowUrl(production, 'https://account.tadoku.app/?flow=f2', { request_url }))
      .toBe(`${branchAccount}/?flow=f2`)
  })
  it('forwards from another branch host to the flow branch', () => {
    expect(branchFlowUrl(production, 'https://bob-fix-0a1b2c3d.account.preview.tadoku.app/login?flow=f3', { return_to: branchApp }))
      .toBe(`${branchAccount}/login?flow=f3`)
  })
  it.each([
    undefined,
    '',
    'https://tadoku.app/',
    'https://account.tadoku.app/',
    'https://evil.example/',
    'https://alice-feat-1a2b3c4d.preview.tadoku.app.evil.example/',
    'https://a.b.preview.tadoku.app/',
    'https://alice.tadoku.app/',
    'javascript:alert(1)',
    'not a url',
  ])('stays put for return_to %s', return_to => {
    expect(branchFlowUrl(production, page, { return_to, request_url: `https://account.tadoku.app/kratos/self-service/login/browser?return_to=${return_to ?? ''}` }))
      .toBeUndefined()
  })
  it('stays put for a malformed request_url', () => {
    expect(branchFlowUrl(production, page, { request_url: 'not a url' })).toBeUndefined()
  })
  it('does not forward when already on the flow branch', () => {
    expect(branchFlowUrl(production, `${branchAccount}/verification?flow=f1`, { return_to: `${branchApp}/` })).toBeUndefined()
  })
})
