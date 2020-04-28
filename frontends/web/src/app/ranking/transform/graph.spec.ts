import { aggregateReadingActivity, aggregateMediaDistribution } from './graph'

describe('aggregateReadingActivity', () => {
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

    expect(aggregateReadingActivity(logs, contest)).toStrictEqual({
      aggregated: {
        jpn: [
          {
            language: 'Japanese',
            x: '2019-06-01T00:00:00.000Z',
            y: 11,
          },
          {
            language: 'Japanese',
            x: '2019-06-02T00:00:00.000Z',
            y: 26,
          },
          {
            language: 'Japanese',
            x: '2019-06-03T00:00:00.000Z',
            y: 14.3,
          },
        ],
      },
      legend: [{ title: 'Japanese', strokeWidth: 10 }],
    })
  })
})

describe('aggregateMediaDistribution', () => {
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

    const result = aggregateMediaDistribution(logs)
    expect(result).toStrictEqual({
      aggregated: [
        {
          amount: 37,
          medium: 'Book',
          color: '#12939A',
        },
        {
          amount: 14.336201,
          medium: 'Full game',
          color: '#79C7E3',
        },
      ],
      legend: [
        { title: 'Book', color: '#12939A', strokeWidth: 10, amount: 37 },
        {
          title: 'Full game',
          color: '#79C7E3',
          strokeWidth: 10,
          amount: 14.336201,
        },
      ],
      totalAmount: 51.336201,
    })
  })
})
