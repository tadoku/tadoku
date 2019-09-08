import React from 'react'
import { configure } from '@storybook/react'
import { addDecorator } from '@storybook/react'
import { GlobalStyle } from '../app/ui/components/Layout'
import Constants from '../app/ui/Constants'

const req = require.context('../app', true, /.stories.tsx$/)
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
