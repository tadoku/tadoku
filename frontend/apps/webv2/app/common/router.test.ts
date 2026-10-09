import { describe, expect, it } from 'vitest'
import {
  getQueryStringDateParameter,
  getQueryStringPageParameter,
} from './router'

describe('getQueryStringPageParameter', () => {
  it('keeps positive pages', () => {
    expect(getQueryStringPageParameter('3')).toBe(3)
    expect(getQueryStringPageParameter(['2'])).toBe(2)
  })

  it('uses the first page for missing, invalid, zero and negative pages', () => {
    for (const param of [undefined, '', 'abc', '0', '-1', ['0']]) {
      expect(getQueryStringPageParameter(param)).toBe(1)
    }
  })
})

describe('getQueryStringDateParameter', () => {
  it('keeps calendar dates', () => {
    expect(getQueryStringDateParameter('2026-09-02')).toBe('2026-09-02')
    expect(getQueryStringDateParameter(['2026-09-02'])).toBe('2026-09-02')
  })

  it('ignores missing, malformed and impossible dates', () => {
    for (const param of [
      undefined,
      '',
      'garbage',
      '2026-02-30',
      '2026-9-2',
      '2026-09-02T00:00:00Z',
    ]) {
      expect(getQueryStringDateParameter(param)).toBeUndefined()
    }
  })
})
