import type { Metadata } from 'next'
import { Toaster } from 'sonner'
import './globals.css'

export const metadata: Metadata = {
  title: 'Go Micro Service',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en" className="dark">
      <head>
        <script src="/config.js" />
      </head>
      <body>
        <Toaster position="top-center" />
        {children}
      </body>
    </html>
  )
}