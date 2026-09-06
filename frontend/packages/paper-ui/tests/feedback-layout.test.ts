import { readFileSync } from 'node:fs'
import { expect, it } from 'vitest'

it('gives Flash content the flexible column when its optional icon is absent', () => {
  const css = readFileSync('src/components/feedback/feedback.css', 'utf8')
  // A leading auto column without an icon placed the message in a max-content
  // track, overflowing the mobile viewport instead of wrapping its text.
  expect(css).toMatch(/\.paper-flash\s*\{[^}]*grid-template-columns:\s*minmax\(0, 1fr\) auto/)
  expect(css).toMatch(/\.paper-flash:has\(> \.paper-flash__icon\)\s*\{[^}]*grid-template-columns:\s*auto minmax\(0, 1fr\) auto/)
})

it('uses compact toast typography instead of inherited heading and paragraph margins', () => {
  const css = readFileSync('src/components/feedback/feedback.css', 'utf8')
  expect(css).toMatch(/\.paper-toast__title\s*\{[^}]*margin:\s*0;[^}]*font-size:\s*var\(--paper-type-body-size\)/)
  expect(css).toMatch(/\.paper-toast__content > p\s*\{[^}]*margin:\s*0/)
})
