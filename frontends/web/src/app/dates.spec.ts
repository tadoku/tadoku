import { prettyDateInUTC, getDates } from './dates'

describe('prettyDate', () => {
  it('should format dates', () => {
    const input = new Date('2020-12-08T00:00:00Z')
    const output = prettyDateInUTC(input)

    expect(output).toEqual('2020-12-8')
  })
})

describe('getDates', () => {
  it('should generate the correct set of dates in UTC', () => {
    const start = new Date('2020-12-01T00:00:00Z')
    const end = new Date('2021-01-01T00:00:00Z')
    const output = getDates(start, end)

    expect(output).toHaveLength(31)
    expect(prettyDateInUTC(output[0])).toEqual('2020-12-1')
    expect(prettyDateInUTC(output[30])).toEqual('2020-12-31')
  })
})
