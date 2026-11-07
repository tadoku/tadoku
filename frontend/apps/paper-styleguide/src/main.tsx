import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { rootProps } from 'paper-ui'
// Paper's font faces; the StyleX plugin appends every compiled style to this stylesheet.
import 'paper-ui/fonts.css'
import faviconUrl from 'paper-ui/assets/brand/favicon.svg'
import { App } from './app/App'

let favicon = document.querySelector<HTMLLinkElement>('link[rel="icon"]')
if (!favicon) {
  favicon = document.createElement('link')
  favicon.rel = 'icon'
  document.head.append(favicon)
}
favicon.type = 'image/svg+xml'
favicon.href = faviconUrl

// The page shell is static HTML, so its root styles are applied before React renders.
document.documentElement.className = rootProps.html().className ?? ''
document.body.className = rootProps.body().className ?? ''

const root = document.getElementById('root')

if (!root) {
  throw new Error('Paper styleguide root element was not found')
}

ReactDOM.createRoot(root).render(
  <React.StrictMode>
    <BrowserRouter>
      <App />
    </BrowserRouter>
  </React.StrictMode>,
)
