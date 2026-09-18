import { describe, expect, it } from 'vitest'
import { compatibleModifiers, initialLogs, personalScoreExplanation, scoreLog, scoreSubmission } from '../src/data'

describe('Log unit and modifier scoring', () => {
  it('uses the recorded reading unit and language for personal and contest scores', () => {
    const log = { activity: 'Reading' as const, amount: 800, unit: 'characters' as const, language: 'Japanese' }
    expect(scoreLog(log)).toBe(2)
    expect(scoreSubmission(log, 'round5').score).toBe(2)
    expect(personalScoreExplanation(log)).toBe('800 characters × 0.0025 points per character = 2 personal points.')
    expect(scoreLog({ ...log, language: 'French', amount: 1200 })).toBeCloseTo(1)
    expect(scoreLog({ ...log, amount: 20, unit: 'sentences' })).toBe(1)
  })

  it('applies only one compatible reading modifier and the listening modifier', () => {
    expect(compatibleModifiers('Reading', 'pages', ['Manga', 'Comic', 'Passive listening'])).toEqual(['Manga'])
    expect(compatibleModifiers('Reading', 'sentences', ['Manga'])).toEqual([])
    expect(compatibleModifiers('Listening', 'minutes', ['Manga', 'Passive listening'])).toEqual(['Passive listening'])
    const reading = { activity: 'Reading' as const, amount: 20, unit: 'pages' as const, modifiers: ['Manga'] as const }
    expect(scoreLog(reading)).toBe(4)
    expect(personalScoreExplanation(reading)).toBe('20 pages × 1 point per page × 0.2 (Manga) = 4 personal points.')
    expect(scoreLog({ ...reading, modifiers: ['Two column'] })).toBe(32)
    expect(scoreLog({ ...reading, unit: 'sentences' })).toBe(1)
    expect(scoreLog({ activity: 'Listening', amount: 40, unit: 'minutes', modifiers: ['Passive listening'] })).toBe(10)
  })

  it('keeps fixed minute-based contest rules separate from personal modifiers and saved scores', () => {
    const reading = { ...initialLogs.find(log => log.id === 'reading-konbini')!, modifiers: ['Manga'] as const }
    expect(scoreLog(reading)).toBe(8.4)
    expect(scoreSubmission(reading, 'round5').score).toBe(8.4)
    expect(scoreSubmission(reading, 'reading-circle')).toEqual({ contestId: 'reading-circle', score: 22.5, basis: '45 tracked minutes × 0.5 points per minute = 22.5 points. Sample contest rule.' })
    expect(scoreSubmission({ activity: 'Listening', amount: 40, unit: 'minutes', modifiers: ['Passive listening'] }, 'listening-circle').score).toBe(40)
    expect(reading.submissions[0].score).toBe(42)
    expect(() => scoreSubmission({ ...reading, minutes: undefined }, 'reading-circle')).toThrow(RangeError)
  })
})
