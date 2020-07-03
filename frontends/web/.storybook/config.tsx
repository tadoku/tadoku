import React from 'react'
import { configure } from '@storybook/react'
import { addDecorator } from '@storybook/react'
import { GlobalStyle } from '../src/app/ui/components/Layout'
import Constants from '../src/app/ui/Constants'

const req = require.context('../src/app', true, /.stories.tsx$/)
function loadStories() {
  req.keys().forEach(filename => req(filename))
}

const withGlobal = cb => (
  <>
    <GlobalStyle {...Constants} />
    {cb()}
  </>
)

addDecorator(withGlobal)

configure(loadStories, module)
