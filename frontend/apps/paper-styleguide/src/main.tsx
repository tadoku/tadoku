import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { bodyClassName, documentClassName } from 'paper-ui'
import './styles.css'
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

document.documentElement.className = documentClassName
document.body.className = bodyClassName

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
