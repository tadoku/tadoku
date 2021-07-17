import React from 'react'
import { storiesOf } from '@storybook/react'
import Footer from './Footer'
import { Contest } from '@app/contest/interfaces'

const contests: Contest[] = [
  {
    id: 1,
    description: '2020 Round 1',
    start: new Date(),
    end: new Date(),
    open: false,
  },
  {
    id: 2,
    description: '2020 Round 1',
    start: new Date(),
    end: new Date(),
    open: false,
  },
]

storiesOf('Footer', module).add('default', () => (
  <div style={{ height: '300px', width: '100%', position: 'relative' }}>
    <Footer contests={contests} />
  </div>
))
