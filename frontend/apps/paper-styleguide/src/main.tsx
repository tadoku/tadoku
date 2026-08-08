import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import 'paper-ui/styles.css'
import faviconUrl from 'paper-ui/assets/brand/favicon.svg'
import { App } from './app/App'
import './styles/shell.css'

let favicon = document.querySelector<HTMLLinkElement>('link[rel="icon"]')
if (!favicon) {
  favicon = document.createElement('link')
  favicon.rel = 'icon'
  document.head.append(favicon)
}
favicon.type = 'image/svg+xml'
favicon.href = faviconUrl

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
