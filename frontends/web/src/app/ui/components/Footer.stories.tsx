import React from 'react'
import { storiesOf } from '@storybook/react'
import Footer from './Footer'

storiesOf('Footer', module).add('default', () => (
  <div style={{ height: '300px', width: '100%', position: 'relative' }}>
    <Footer />
  </div>
))
