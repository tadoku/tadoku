import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { ToastProvider, bodyClassName, documentClassName } from 'paper-ui'
import './styles.css'
import { PlaygroundProvider } from './state'
import { App } from './App'

document.documentElement.className = documentClassName
document.body.className = bodyClassName

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode><BrowserRouter><PlaygroundProvider><ToastProvider><App /></ToastProvider></PlaygroundProvider></BrowserRouter></React.StrictMode>,
)
