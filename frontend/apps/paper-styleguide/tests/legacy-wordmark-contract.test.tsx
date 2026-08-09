import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
import { catalogRegistry } from 'paper-ui/catalog'
import { DocsShell } from '../src/app/DocsShell'

const legacyWordmark = readFileSync(
  resolve('../../packages/ui/components/logo.svg'),
  'utf8',
)
const paperWordmark = readFileSync(
  resolve('../../packages/paper-ui/src/assets/brand/wordmark-accent.svg'),
  'utf8',
)
const legacyReversedWordmark = readFileSync(
  resolve('../../packages/ui/components/logo-light.svg'),
  'utf8',
)
const paperReversedWordmark = readFileSync(
  resolve('../../packages/paper-ui/src/assets/brand/wordmark-reversed.svg'),
  'utf8',
)

describe('legacy design-system wordmark contract', () => {
  it('uses an exact copy of the original 158 by 29 wordmark in the shell', () => {
    expect(paperWordmark.trimEnd()).toBe(legacyWordmark.trimEnd())
    expect(paperReversedWordmark.trimEnd()).toBe(
      legacyReversedWordmark.trimEnd(),
    )

    render(
      <MemoryRouter>
        <DocsShell documents={catalogRegistry.documents}>Content</DocsShell>
      </MemoryRouter>,
    )

    const home = screen.getByRole('link', { name: 'Tadoku' })
    const wordmark = screen.getByRole('img', { name: 'Tadoku' })
    const logoImages = home.querySelectorAll('img')

    expect(wordmark).toHaveClass('docs-wordmark')
    expect(logoImages).toHaveLength(2)
    expect(home).not.toHaveTextContent('Tadoku Paper')
    expect(logoImages[0]).toHaveAttribute(
      'src',
      expect.stringContaining('wordmark-accent.svg'),
    )
    expect(logoImages[1]).toHaveAttribute(
      'src',
      expect.stringContaining('wordmark-reversed.svg'),
    )

    for (const logo of logoImages) {
      expect(logo).toHaveAttribute('alt', '')
      expect(logo).toHaveAttribute('aria-hidden', 'true')
      expect(logo).toHaveAttribute('width', '126.4')
      expect(logo).toHaveAttribute('height', '23.2')
    }
  })
})
