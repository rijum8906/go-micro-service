import ReactDOM from 'react-dom/client'
import { RouterProvider } from '@tanstack/react-router'
import { bootstrapAuth } from '#/lib/auth-bootstrap'
import { getRouter } from './router'

const router = getRouter()

const rootElement = document.getElementById('app')!
const root = ReactDOM.createRoot(rootElement)

void bootstrapAuth().then(() => {
  root.render(<RouterProvider router={router} />)
})
