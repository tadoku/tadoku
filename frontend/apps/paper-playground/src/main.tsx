import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { ToastProvider } from 'paper-ui'
import 'paper-ui/styles.css'
import './styles.css'
import { PlaygroundProvider } from './state'
import { App } from './App'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode><BrowserRouter><PlaygroundProvider><ToastProvider><App /></ToastProvider></PlaygroundProvider></BrowserRouter></React.StrictMode>,
)
