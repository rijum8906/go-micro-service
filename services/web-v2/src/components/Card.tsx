import type { ReactNode } from 'react'
import { cn } from '#/lib/utils'

interface CardProps {
  children: ReactNode
  isDark: boolean
  className?: string
}

export function Card({ children, isDark, className }: CardProps) {
  return (
    <div
      className={cn(
        'w-full max-w-sm rounded-2xl p-8 transition-colors duration-300',
        isDark
          ? 'bg-[#30302E] shadow-[inset_0_0_0_0.2px_#F2EDE4]'
          : 'bg-white/40 shadow-[inset_0_0_0_0.3px_#9A9A9A] backdrop-blur-sm',
        className,
      )}
    >
      {children}
    </div>
  )
}