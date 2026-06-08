import { cn } from '#/lib/utils'

const sizeClasses = {
  sm: 'h-8 w-8 text-xs',
  md: 'h-12 w-12 text-sm',
  lg: 'h-20 w-20 text-xl',
} as const

interface AvatarProps {
  firstName: string
  lastName: string
  size?: keyof typeof sizeClasses
  className?: string
}

export function Avatar({ firstName, lastName, size = 'md', className }: AvatarProps) {
  const initials = `${firstName.charAt(0)}${lastName.charAt(0)}`.toUpperCase()

  return (
    <div
      className={cn(
        'rounded-full bg-[#C97D4E] text-[#F2EDE4] flex items-center justify-center font-medium select-none',
        sizeClasses[size],
        className,
      )}
      aria-label={`${firstName} ${lastName}`}
    >
      {initials}
    </div>
  )
}