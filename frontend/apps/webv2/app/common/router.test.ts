import { describe, expect, it } from 'vitest'
import { getQueryStringPageParameter } from './router'

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
