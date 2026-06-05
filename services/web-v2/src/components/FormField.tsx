interface FormFieldProps {
  label: string
  type?: 'text' | 'email' | 'password'
  value: string
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void
  onBlur: () => void
  placeholder?: string
  errors: ReadonlyArray<unknown>
  isDark: boolean
}

export function FormField({
  label,
  type = 'text',
  value,
  onChange,
  onBlur,
  placeholder,
  errors,
  isDark,
}: FormFieldProps) {
  return (
    <div className="flex flex-col gap-1.5">
      <label className={`text-sm font-medium ${isDark ? 'text-[#D9D9D9]' : 'text-[#262526]'}`}>
        {label}
      </label>
      <input
        type={type}
        value={value}
        onBlur={onBlur}
        onChange={onChange}
        placeholder={placeholder}
        className={`h-10 px-4 rounded-lg border text-sm outline-none focus:border-[#C97D4E]/50 transition-colors ${isDark ? 'border-[#F2EDE4]/10 bg-[#262624] text-[#F2EDE4] placeholder:text-[#F2EDE4]/30' : 'border-[#262526]/10 bg-[#F2EDE4]/60 text-[#262526] placeholder:text-[#262526]/30'}`}
      />
      {errors.length > 0 && (
        <p className="text-xs text-[#B85C5C]">
          {errors.map((e) => {
            if (typeof e === 'string') return e
            if (e && typeof e === 'object' && 'message' in e) return String(e.message)
            return ''
          }).filter(Boolean).join(', ')}
        </p>
      )}
    </div>
  )
}