import React from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import { ContestList } from '@app/immersion/ContestList'

Reflect.set(globalThis, 'React', React)

describe('ContestList', () => {
  it('renders the empty state when there are no contests', () => {
    const markup = renderToStaticMarkup(
      <ContestList
        list={{ contests: [], total_size: 0, next_page_token: '' }}
      />,
    )

    expect(markup).toContain('>No contests found<')
  })
})
