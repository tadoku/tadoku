import {
  aggregateContestLogsByDays,
  aggregateContestLogsByMedium,
} from './transform'

describe('aggregateContestLogsByDays', () => {
  it('should aggregate correctly', () => {
    const contest = {
      id: 1,
      description: '2019.06 Test Round',
      start: new Date('2019-06-01T00:00:00.000Z'),
      end: new Date('2019-06-03T23:59:59.000Z'),
      open: true,
    }
    const logs = [
      {
        id: 186,
        contestId: 1,
        userId: 32,
        languageCode: 'jpn',
        mediumId: 1,
        amount: 11,
        adjustedAmount: 11,
        description: 'とらドラ！第３巻',
        date: new Date('2019-06-01T22:24:57.973Z'),
      },
      {
        id: 175,
        contestId: 1,
        userId: 32,
        languageCode: 'jpn',
        mediumId: 1,
        amount: 26,
        adjustedAmount: 26,
        description: 'とらドラ！第３巻',
        date: new Date('2019-06-02T21:43:10.580Z'),
      },
      {
        id: 173,
        contestId: 1,
        userId: 32,
        languageCode: 'jpn',
        mediumId: 4,
        amount: 86,
        adjustedAmount: 14.336201,
        description: '428 〜封鎖された渋谷で〜',
        date: new Date('2019-06-03T19:36:20.145Z'),
      },
    ]

    expect(aggregateContestLogsByDays(logs, contest)).toStrictEqual({
      aggregated: {
        jpn: [
          {
            language: 'Japanese',
            size: 2,
            x: new Date('2019-06-01T00:00:00.000Z'),
            y: 11,
          },
          {
            language: 'Japanese',
            size: 2,
            x: new Date('2019-06-02T00:00:00.000Z'),
            y: 26,
          },
          {
            language: 'Japanese',
            size: 2,
            x: new Date('2019-06-03T00:00:00.000Z'),
            y: 14.3,
          },
        ],
      },
      legend: [{ title: 'Japanese' }],
    })
  })
})

describe('aggregateContestLogsByMedium', () => {
  it('should aggregate correctly', () => {
    const logs = [
      {
        id: 186,
        contestId: 1,
        userId: 32,
        languageCode: 'jpn',
        mediumId: 1,
        amount: 11,
        adjustedAmount: 11,
        description: 'とらドラ！第３巻',
        date: new Date('2019-06-01T22:24:57.973Z'),
      },
      {
        id: 175,
        contestId: 1,
        userId: 32,
        languageCode: 'jpn',
        mediumId: 1,
        amount: 26,
        adjustedAmount: 26,
        description: 'とらドラ！第３巻',
        date: new Date('2019-06-02T21:43:10.580Z'),
      },
      {
        id: 173,
        contestId: 1,
        userId: 32,
        languageCode: 'jpn',
        mediumId: 4,
        amount: 86,
        adjustedAmount: 14.336201,
        description: '428 〜封鎖された渋谷で〜',
        date: new Date('2019-06-03T19:36:20.145Z'),
      },
    ]

    const result = aggregateContestLogsByMedium(logs)
    expect(result).toStrictEqual({
      aggregated: [
        {
          amount: 14.336201,
          medium: 'Full game',
        },
        {
          amount: 37,
          medium: 'Book',
        },
      ],
      legend: [{ title: 'Full game' }, { title: 'Book' }],
      totalAmount: 51.336201,
    })
  })
})
