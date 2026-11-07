import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { ToastProvider, rootProps } from 'paper-ui'
// Paper's font faces; the StyleX plugin appends every compiled style to this stylesheet.
import 'paper-ui/fonts.css'
import { PlaygroundProvider } from './state'
import { App } from './App'

// The page shell is static HTML, so its root styles are applied before React renders.
document.documentElement.className = rootProps.html().className ?? ''
document.body.className = rootProps.body().className ?? ''

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode><BrowserRouter><PlaygroundProvider><ToastProvider><App /></ToastProvider></PlaygroundProvider></BrowserRouter></React.StrictMode>,
)
